package main

import (
	"HAND_IN_2/account"
	"HAND_IN_2/block"
	"HAND_IN_2/peer"
	"HAND_IN_2/rsa"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"testing"
	"time"
)

var ledger = make([]*account.Ledger, 10)
var accounts = make([]*account.Account, 10)
var ports = []int{}
var peers = make([]*peer.Peer, 10)
var counter = 0
var n = 0

// GetOutboundIP preferred outbound ip of this machine
func GetOutboundIP2() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

// Random transaction generator
func randomTransaction(peer *peer.Peer, counter *int) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	fromNumber := rand.Intn(len(accounts))
	toNumber := rand.Intn(len(accounts))
	for fromNumber == toNumber {
		toNumber = rand.Intn(len(accounts))
	}

	FromAccount := accounts[fromNumber]
	ToAccount := accounts[toNumber]

	pkFrom := rsa.EncodePublicKey(FromAccount.Pk)
	pkTo := rsa.EncodePublicKey(ToAccount.Pk)

	randomAmount := rand.Intn(1000)
	ac := account.Transaction{ID: strconv.Itoa(*counter), From: pkFrom, To: pkTo, Amount: randomAmount}
	st := account.SignTransaction(FromAccount.Sk, &ac)
	peer.ExecuteTransaction(st)
	*counter++
	peer.FloodTransaction(st)
}

// Prints the ledgers of all peers
func PrintLedgers(peers []*peer.Peer) {
	for i, peer := range peers {
		fmt.Printf("Peer %d Ledger:\n", i+1)
		balances := peer.Ledger.Accounts

		accountNames := make([]string, 0, len(balances))
		for account := range balances {
			accountNames = append(accountNames, account)
		}
		sort.Strings(accountNames)

		accountNumber := 1
		for _, account := range accountNames {
			fmt.Printf("Account %d: %d\n", accountNumber, balances[account])
			accountNumber++
		}
		fmt.Println()
	}
}

func setup() {
	ip := GetOutboundIP2()
	fmt.Println("IP: ", ip.String())
	// Generate random ports for peers
	for i := 0; i < 10; i++ {
		time.Sleep(50 * time.Millisecond)
		port, err := GetFreePort()
		if err != nil {
			log.Fatalf("Failed to get a free port: %v", err)
		}
		ports = append(ports, port)
	}

	//create 10 peers and set their ledger to genesisLedger
	genesisLedger := account.MakeLedger()
	fmt.Println("genesisLedger: ", genesisLedger)
	// Create 10 accounts
	for i := 0; i < 10; i++ {

		// Create a new account
		accounts[i] = account.MakeAccount()
		// Add the account to the ledger
		genesisLedger.Accounts[rsa.EncodePublicKey(accounts[i].Pk)] = 1000000
	}

	for i := 0; i < 10; i++ {
		peers[i] = peer.NewPeer(ports[i], genesisLedger, "Peer"+strconv.Itoa(i), ip.String(), accounts[i])
	}

	// Start a new network
	for i := 0; i < 10; i++ {
		go peers[i].StartNewNetwork()
		fmt.Println("Peer ", i, " started a new network")
	}

	fmt.Println("genesisLedger: ", genesisLedger)

	// Connect all peers
	// Connect peers to the network
	for i := 0; i < 10; i++ {
		go peers[i].Connect(ip.String(), peers[0].Port)
		fmt.Println("Peer ", i, " connected to peer 0")
		time.Sleep(500 * time.Millisecond)
		go peers[i].AskForPeers(peers[0].Port)
	}

	// Print the ledgers of all peers
	PrintLedgers(peers)
}

// create a block with random signedTransactions and return the block
func createBlock(accounts []*account.Account, prevHash string) block.Block {
	block := block.Block{PrevHash: prevHash}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 2; i++ {
		fromNumber := rand.Intn(len(accounts))
		toNumber := rand.Intn(len(accounts))
		for fromNumber == toNumber {
			toNumber = rand.Intn(len(accounts))
		}

		FromAccount := accounts[fromNumber]
		ToAccount := accounts[toNumber]

		pkFrom := rsa.EncodePublicKey(FromAccount.Pk)
		pkTo := rsa.EncodePublicKey(ToAccount.Pk)

		randomAmount := rand.Intn(10)
		ac := account.Transaction{ID: strconv.Itoa(i), From: pkFrom, To: pkTo, Amount: randomAmount}
		st := account.SignTransaction(FromAccount.Sk, &ac)
		block.AddTransaction(&st)
	}
	return block
}

func Test(t *testing.T) {

	//create transactions, add them to a block and sign the block and then verify the block
	t.Run("blockVerify", func(t *testing.T) {
		account1 := account.MakeAccount()
		account2 := account.MakeAccount()
		transaction := account.Transaction{ID: "1", From: rsa.EncodePublicKey(account1.Pk), To: rsa.EncodePublicKey(account2.Pk), Amount: 100}
		signedTransaction := account.SignTransaction(account1.Sk, &transaction)
		block := block.Block{PrevHash: "0"}
		block.AddTransaction(&signedTransaction)
		signedBlock := block.SignBlock(*account1)

		fmt.Println("Block: ", block)
		fmt.Println("SignedBlock: ", signedBlock)
		fmt.Println("BlockHash: ", block.HashBlock(account1.Pk))
		fmt.Println("verifyBlock: ", block.VerifyBlock(account1.Pk, signedBlock))
		if !block.VerifyBlock(account1.Pk, signedBlock) {
			t.Errorf("Block verification failed")
		}
	})

	// Tests that correctly signed transactions in a block  are verified
	t.Run("blockVerifyAllTransactions", func(t *testing.T) {
		account1 := account.MakeAccount()
		account2 := account.MakeAccount()
		account3 := account.MakeAccount()
		transaction1 := account.Transaction{ID: "1", From: rsa.EncodePublicKey(account1.Pk), To: rsa.EncodePublicKey(account2.Pk), Amount: 100}
		transaction2 := account.Transaction{ID: "2", From: rsa.EncodePublicKey(account2.Pk), To: rsa.EncodePublicKey(account3.Pk), Amount: 100}
		transaction3 := account.Transaction{ID: "3", From: rsa.EncodePublicKey(account3.Pk), To: rsa.EncodePublicKey(account1.Pk), Amount: 100}
		signedTransaction1 := account.SignTransaction(account1.Sk, &transaction1)
		signedTransaction2 := account.SignTransaction(account2.Sk, &transaction2)
		signedTransaction3 := account.SignTransaction(account3.Sk, &transaction3)
		block := block.Block{PrevHash: "0"}
		block.AddTransaction(&signedTransaction1)
		block.AddTransaction(&signedTransaction2)
		block.AddTransaction(&signedTransaction3)

		if !block.VerifyBlockTransactions() {
			t.Errorf("Block verification failed")
		}
	})

	// Tests that incorrectly signed transactions in a block are not verified
	t.Run("oneInvalidTransactionVerifyBlockTransactions", func(t *testing.T) {
		account1 := account.MakeAccount()
		account2 := account.MakeAccount()
		account3 := account.MakeAccount()
		transaction1 := account.Transaction{ID: "1", From: rsa.EncodePublicKey(account1.Pk), To: rsa.EncodePublicKey(account2.Pk), Amount: 100}
		transaction2 := account.Transaction{ID: "2", From: rsa.EncodePublicKey(account2.Pk), To: rsa.EncodePublicKey(account3.Pk), Amount: 100}
		transaction3 := account.Transaction{ID: "3", From: rsa.EncodePublicKey(account3.Pk), To: rsa.EncodePublicKey(account1.Pk), Amount: 100}
		signedTransaction1 := account.SignTransaction(account1.Sk, &transaction1)
		signedTransaction2 := account.SignTransaction(account1.Sk, &transaction2)
		signedTransaction3 := account.SignTransaction(account3.Sk, &transaction3)
		block := block.Block{PrevHash: "0"}
		block.AddTransaction(&signedTransaction1)
		block.AddTransaction(&signedTransaction2)
		block.AddTransaction(&signedTransaction3)

		if block.VerifyBlockTransactions() {
			t.Errorf("Block verification succeeded with an invalid Transaction")
		}
	})

	//test that height of a block is correctly set
	t.Run("blockHeight", func(t *testing.T) {
		account1 := account.MakeAccount()
		block1 := block.NewBlock(nil, account1.Pk)
		block2 := block.NewBlock(block1, account1.Pk)

		if block2.Height != 2 {
			t.Errorf("Block height is not correctly set")
		}
	})

	//test that we can compare two heights of blocks
	t.Run("blockHeightComparison", func(t *testing.T) {
		account1 := account.MakeAccount()
		block1 := block.NewBlock(nil, account1.Pk)
		block2 := block.NewBlock(block1, account1.Pk)

		if block1.Height >= block2.Height {
			t.Errorf("Block height comparison failed")
		}
	})

	//test that if we create a blockchain, the chainHead of the blockchain is updated
	t.Run("chainHeadUpdate", func(t *testing.T) {
		ledger1 := account.MakeLedger()
		account1 := account.MakeAccount()
		block1 := block.NewBlock(nil, account1.Pk)
		block2 := block.NewBlock(block1, account1.Pk)
		block3 := block.NewBlock(block2, account1.Pk)

		blockchain := block.NewBlockchain(ledger1)
		blockchain.AddBlock(*block1)
		blockchain.AddBlock(*block2)
		blockchain.AddBlock(*block3)

		if blockchain.ChainHead.Hash != block3.Hash {
			t.Errorf("ChainHead not updated")
		}
	})

	//test that if we create a blockchain, with 2 chains of blocks where one is longer the chainHead of the blockchain is updated
	t.Run("chainHeadUpdateLongerChain", func(t *testing.T) {
		ledger1 := account.MakeLedger()
		account1 := account.MakeAccount()
		block1 := block.NewBlock(nil, account1.Pk)
		block2 := block.NewBlock(block1, account1.Pk)
		block3 := block.NewBlock(block2, account1.Pk)
		block4 := block.NewBlock(block3, account1.Pk)

		block5 := block.NewBlock(block1, account1.Pk)
		block6 := block.NewBlock(block5, account1.Pk)

		blockchain := block.NewBlockchain(ledger1)
		blockchain.AddBlock(*block1)
		blockchain.AddBlock(*block2)
		blockchain.AddBlock(*block3)
		blockchain.AddBlock(*block4)
		blockchain.AddBlock(*block5)
		blockchain.AddBlock(*block6)

		if blockchain.ChainHead.Hash != block4.Hash {
			t.Errorf("ChainHead not updated")
		}
	})

	//test that transactions of a block are added to the ledger
	t.Run("transactionsToLedger", func(t *testing.T) {
		genesisLedger := account.MakeLedger()
		// Create 10 accounts
		for i := 0; i < 10; i++ {
			// Create a new account
			accounts[i] = account.MakeAccount()
			// Add the account to the ledger
			genesisLedger.Accounts[rsa.EncodePublicKey(accounts[i].Pk)] = 100
		}

		transaction1 := account.Transaction{ID: "1", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[2].Pk), Amount: 50}
		transaction2 := account.Transaction{ID: "2", From: rsa.EncodePublicKey(accounts[2].Pk), To: rsa.EncodePublicKey(accounts[3].Pk), Amount: 30}
		transaction3 := account.Transaction{ID: "3", From: rsa.EncodePublicKey(accounts[3].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 20}
		signedTransaction1 := account.SignTransaction(accounts[1].Sk, &transaction1)
		signedTransaction2 := account.SignTransaction(accounts[2].Sk, &transaction2)
		signedTransaction3 := account.SignTransaction(accounts[3].Sk, &transaction3)
		block1 := block.NewBlock(&block.Block{Hash: ""}, accounts[1].Pk)
		block1.AddTransaction(&signedTransaction1)
		block1.AddTransaction(&signedTransaction2)
		block1.AddTransaction(&signedTransaction3)

		blockchain := block.NewBlockchain(genesisLedger)

		blockchain.AddBlock(*block1)

		fmt.Println("blockchain ledger: ", blockchain.Ledger.Accounts)
		if blockchain.Ledger.Accounts[rsa.EncodePublicKey(accounts[1].Pk)] != 69 {
			t.Errorf("Account balance not correct")
		}
	})

	//test if we add a longer chain, the blockchain ledger is updated with the new blocks
	t.Run("longerChainToLedger", func(t *testing.T) {
		var accounts = make([]*account.Account, 3)
		genesisLedger := account.MakeLedger()
		for i := 0; i < 3; i++ {
			// Create a new account
			accounts[i] = account.MakeAccount()
			// Add the account to the ledger
			genesisLedger.Accounts[rsa.EncodePublicKey(accounts[i].Pk)] = 100
		}

		blockchain := block.NewBlockchain(genesisLedger)

		transaction1 := account.Transaction{ID: "1", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 50}
		transaction2 := account.Transaction{ID: "2", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[2].Pk), Amount: 30}
		signedTransaction1 := account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 := account.SignTransaction(accounts[1].Sk, &transaction2)
		block1 := block.NewBlock(&block.Block{Hash: ""}, accounts[1].Pk)
		block1.AddTransaction(&signedTransaction1)
		block1.AddTransaction(&signedTransaction2)

		transaction1 = account.Transaction{ID: "3", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 20}
		transaction2 = account.Transaction{ID: "4", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[0].Pk), Amount: 30}
		signedTransaction1 = account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 = account.SignTransaction(accounts[1].Sk, &transaction2)
		block2 := block.NewBlock(block1, accounts[1].Pk)
		block2.AddTransaction(&signedTransaction1)
		block2.AddTransaction(&signedTransaction2)

		transaction1 = account.Transaction{ID: "5", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 30}
		transaction2 = account.Transaction{ID: "6", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[0].Pk), Amount: 10}
		signedTransaction1 = account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 = account.SignTransaction(accounts[1].Sk, &transaction2)
		block3 := block.NewBlock(block1, accounts[1].Pk)
		block3.AddTransaction(&signedTransaction1)
		block3.AddTransaction(&signedTransaction2)

		transaction1 = account.Transaction{ID: "7", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 10}
		transaction2 = account.Transaction{ID: "8", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[0].Pk), Amount: 20}
		signedTransaction1 = account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 = account.SignTransaction(accounts[1].Sk, &transaction2)
		block4 := block.NewBlock(block3, accounts[1].Pk)
		block4.AddTransaction(&signedTransaction1)
		block4.AddTransaction(&signedTransaction2)

		blockchain.AddBlock(*block1)
		blockchain.AddBlock(*block2)
		blockchain.AddBlock(*block3)
		blockchain.AddBlock(*block4)

		fmt.Println("ledger amount for account 1 ", blockchain.Ledger.Accounts[rsa.EncodePublicKey(accounts[0].Pk)])

		if blockchain.Ledger.Accounts[rsa.EncodePublicKey(accounts[0].Pk)] != 38 {
			t.Errorf("Account balance not correct")
		}

	})
}
