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

// Random transaction generator
func randomMaliciousTransaction(peer *peer.Peer, counter *int) {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	randomAmount := rand.Intn(1000)
	malicousIntent := rand.Intn(4)
	fromNumber := rand.Intn(len(accounts))
	toNumber := rand.Intn(len(accounts))
	for fromNumber == toNumber {
		toNumber = rand.Intn(len(accounts))
	}

	FromAccount := accounts[fromNumber]
	ToAccount := accounts[toNumber]

	pkFrom := rsa.EncodePublicKey(FromAccount.Pk)
	pkTo := rsa.EncodePublicKey(ToAccount.Pk)

	signingSk := FromAccount.Sk

	//switch case to determine the malicious intent
	switch malicousIntent {
	case 0: //send a transaction with a negative amount
		randomAmount = -1
		fmt.Println("sending a transaction with a negative amount")
	case 1: //send a transaction with a 0 amount
		randomAmount = 0
		fmt.Println("sending a transaction with a 0 amount")
	case 2: //send a transaction with an invalid signature
		signingSk = ToAccount.Sk
		fmt.Println("sending a transaction with an invalid signature")
	case 3: //send a transaction which will cause the account to go into negative balance
		randomAmount = 10000000
		fmt.Println("sending a transaction which will cause the account to go into negative balance")
	}

	ac := account.Transaction{ID: strconv.Itoa(*counter), From: pkFrom, To: pkTo, Amount: randomAmount}
	st := account.SignTransaction(signingSk, &ac)
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

func Test(t *testing.T) {

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

	fmt.Println("ports: ", ports)

	//create 10 peers and set their ledger to genesisLedger
	genesisLedger := account.MakeLedger()

	// Create 10 accounts
	for i := 0; i < 10; i++ {

		// Create a new account
		accounts[i] = account.MakeAccount()
		// Add the account to the ledger
		genesisLedger.Accounts[rsa.EncodePublicKey(accounts[i].Pk)] = 1000000
	}
	genesisTime := time.Now()
	for i := 0; i < 10; i++ {
		peers[i] = peer.NewPeer(ports[i], genesisLedger, "Peer"+strconv.Itoa(i), ip.String(), accounts[i], genesisTime)
	}

	// Start a new network
	for i := 0; i < 10; i++ {
		go peers[i].StartNewNetwork()
		fmt.Println("Peer ", i, " started a new network")
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("genesisLedger: ", genesisLedger)

	// Connect all peers
	// Connect peers to the network
	for i := 0; i < 10; i++ {
		go peers[i].Connect(ip.String(), peers[0].Port)
		fmt.Println("Peer ", i, " connected to peer 0")
		time.Sleep(50 * time.Millisecond)
		go peers[i].AskForPeers(peers[0].Port)
	}

	time.Sleep(1 * time.Second)

	//create transactions, add them to a block and sign the block and then verify the block
	t.Run("blockVerify", func(t *testing.T) {
		account1 := account.MakeAccount()
		account2 := account.MakeAccount()
		transaction := account.Transaction{ID: "1", From: rsa.EncodePublicKey(account1.Pk), To: rsa.EncodePublicKey(account2.Pk), Amount: 100}
		signedTransaction := account.SignTransaction(account1.Sk, &transaction)
		transactions := []account.SignedTransaction{signedTransaction}
		block := block.NewBlock(&block.Block{Hash: ""}, transactions, account1.Pk)
		signedBlock := block.SignBlock(account1.Pk, account1.Sk)

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
		transactions := []account.SignedTransaction{signedTransaction1, signedTransaction2, signedTransaction3}
		block := block.NewBlock(&block.Block{Hash: ""}, transactions, account1.Pk)

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
		transactions := []account.SignedTransaction{signedTransaction1, signedTransaction2, signedTransaction3}
		block := block.NewBlock(&block.Block{Hash: ""}, transactions, account1.Pk)

		if block.VerifyBlockTransactions() {
			t.Errorf("Block verification succeeded with an invalid Transaction")
		}
	})

	//test that height of a block is correctly set
	t.Run("blockHeight", func(t *testing.T) {
		account1 := account.MakeAccount()
		block1 := block.NewBlock(nil, make([]account.SignedTransaction, 0), account1.Pk)
		block2 := block.NewBlock(block1, make([]account.SignedTransaction, 0), account1.Pk)

		if block2.Height != 2 {
			t.Errorf("Block height is not correctly set")
		}
	})

	//test that we can compare two heights of blocks
	t.Run("blockHeightComparison", func(t *testing.T) {
		account1 := account.MakeAccount()
		block1 := block.NewBlock(nil, make([]account.SignedTransaction, 0), account1.Pk)
		block2 := block.NewBlock(block1, make([]account.SignedTransaction, 0), account1.Pk)

		if block1.Height >= block2.Height {
			t.Errorf("Block height comparison failed")
		}
	})

	//test that if we create a blockchain, the chainHead of the blockchain is updated
	t.Run("chainHeadUpdate", func(t *testing.T) {
		ledger1 := account.MakeLedger()
		account1 := account.MakeAccount()
		block1 := block.NewBlock(nil, make([]account.SignedTransaction, 0), account1.Pk)
		block2 := block.NewBlock(block1, make([]account.SignedTransaction, 0), account1.Pk)
		block3 := block.NewBlock(block2, make([]account.SignedTransaction, 0), account1.Pk)

		blockchain := block.NewBlockchain(ledger1, time.Now())
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
		block1 := block.NewBlock(nil, make([]account.SignedTransaction, 0), account1.Pk)
		block2 := block.NewBlock(block1, make([]account.SignedTransaction, 0), account1.Pk)
		block3 := block.NewBlock(block2, make([]account.SignedTransaction, 0), account1.Pk)
		block4 := block.NewBlock(block3, make([]account.SignedTransaction, 0), account1.Pk)

		block5 := block.NewBlock(block1, make([]account.SignedTransaction, 0), account1.Pk)
		block6 := block.NewBlock(block5, make([]account.SignedTransaction, 0), account1.Pk)

		blockchain := block.NewBlockchain(ledger1, time.Now())
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
		transactions := []account.SignedTransaction{signedTransaction1, signedTransaction2, signedTransaction3}
		block1 := block.NewBlock(&block.Block{Hash: ""}, transactions, accounts[1].Pk)

		blockchain := block.NewBlockchain(genesisLedger, time.Now())

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

		blockchain := block.NewBlockchain(genesisLedger, time.Now())

		transaction1 := account.Transaction{ID: "1", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 50}
		transaction2 := account.Transaction{ID: "2", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[2].Pk), Amount: 30}
		signedTransaction1 := account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 := account.SignTransaction(accounts[1].Sk, &transaction2)
		transactions := []account.SignedTransaction{signedTransaction1, signedTransaction2}
		block1 := block.NewBlock(&block.Block{Hash: ""}, transactions, accounts[1].Pk)

		transaction1 = account.Transaction{ID: "3", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 20}
		transaction2 = account.Transaction{ID: "4", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[0].Pk), Amount: 30}
		signedTransaction1 = account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 = account.SignTransaction(accounts[1].Sk, &transaction2)
		transactions = []account.SignedTransaction{signedTransaction1, signedTransaction2}
		block2 := block.NewBlock(block1, transactions, accounts[1].Pk)

		transaction1 = account.Transaction{ID: "5", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 30}
		transaction2 = account.Transaction{ID: "6", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[0].Pk), Amount: 10}
		signedTransaction1 = account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 = account.SignTransaction(accounts[1].Sk, &transaction2)
		transactions = []account.SignedTransaction{signedTransaction1, signedTransaction2}
		block3 := block.NewBlock(block1, transactions, accounts[1].Pk)

		transaction1 = account.Transaction{ID: "7", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 10}
		transaction2 = account.Transaction{ID: "8", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[0].Pk), Amount: 20}
		signedTransaction1 = account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 = account.SignTransaction(accounts[1].Sk, &transaction2)
		transactions = []account.SignedTransaction{signedTransaction1, signedTransaction2}
		block4 := block.NewBlock(block3, transactions, accounts[1].Pk)

		blockchain.AddBlock(*block1)
		blockchain.AddBlock(*block2)
		blockchain.AddBlock(*block3)
		blockchain.AddBlock(*block4)

		fmt.Println("ledger amount for account 1 ", blockchain.Ledger.Accounts[rsa.EncodePublicKey(accounts[0].Pk)])

		if blockchain.Ledger.Accounts[rsa.EncodePublicKey(accounts[0].Pk)] != 38 {
			t.Errorf("Account balance not correct")
		}

	})

	//test that we can send a block without transaction to peers
	t.Run("sendEmptyBlockToPeers", func(t *testing.T) {
		transactions := []account.SignedTransaction{}
		block1 := block.NewBlock(&block.Block{Hash: ""}, transactions, accounts[0].Pk)
		// create lotteryBlock
		lotteryBlock := block.NewLotteryBlock(*block1, accounts[0].Pk, accounts[0].Sk, []byte{1}, []byte{1})
		initialState := block1.VerifyBlock(accounts[0].Pk, lotteryBlock.Signature)
		if !initialState {
			t.Errorf("Block not signed correctly")
		}
		peers[0].SendBlockToPeers(*lotteryBlock)
		time.Sleep(1 * time.Second)
		if len(peers[1].Blockchain.Blocks) == 0 {
			t.Errorf("Block not received by peer")
		}
	})

	//test that we can send a block with a transaction to peers
	t.Run("sendBlockToPeers", func(t *testing.T) {

		transaction1 := account.Transaction{ID: "1", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 50}
		transaction2 := account.Transaction{ID: "2", From: rsa.EncodePublicKey(accounts[1].Pk), To: rsa.EncodePublicKey(accounts[2].Pk), Amount: 30}
		signedTransaction1 := account.SignTransaction(accounts[0].Sk, &transaction1)
		signedTransaction2 := account.SignTransaction(accounts[1].Sk, &transaction2)
		transactions := []account.SignedTransaction{signedTransaction1, signedTransaction2}
		block1 := block.NewBlock(&block.Block{Hash: ""}, transactions, accounts[0].Pk)

		fmt.Println(peers[1].Blockchain.Ledger.Accounts)

		// create lotteryBlock
		lotteryBlock := block.Lottery{Block: block1, Slot: []byte{1}, Pk: accounts[0].Pk, Draw: []byte{1}, Signature: block1.SignBlock(accounts[0].Pk, accounts[0].Sk)}
		peers[0].SendBlockToPeers(lotteryBlock)
		time.Sleep(3 * time.Second)
		fmt.Println(peers[1].Blockchain.Ledger.Accounts)
		if len(peers[1].Blockchain.Blocks[block1.Hash].Transactions) == 0 {
			t.Errorf("Block not proccessed by peer")
		}
	})

	//test that we can flood a transaction to peers
	t.Run("floodTransactionToPeers", func(t *testing.T) {
		transaction1 := account.Transaction{ID: "1", From: rsa.EncodePublicKey(accounts[0].Pk), To: rsa.EncodePublicKey(accounts[1].Pk), Amount: 50}
		signedTransaction1 := account.SignTransaction(accounts[0].Sk, &transaction1)
		peers[0].FloodTransaction(signedTransaction1)
		time.Sleep(1 * time.Second)
		if len(peers[1].Transactions) == 0 {
			t.Errorf("Transaction not received by peer")
		}
	})

	//test that we can flood random transactions, that they are received by peers and that they are processed.
	t.Run("floodRandomMessages", func(t *testing.T) {
		time.Sleep(6 * time.Second)
		for i := 0; i < 10; i++ {
			for j := 0; j < 20; j++ {
				//create a random number between 0 and 10
				rand := rand.New(rand.NewSource(time.Now().UnixNano()))
				malicious := rand.Intn(100)
				if malicious < 10 {
					go randomMaliciousTransaction(peers[i], &counter)
				} else {
					go randomTransaction(peers[i], &counter)
				}
				time.Sleep(200 * time.Millisecond)
			}
		}
		time.Sleep(50 * time.Second)
		if len(peers[1].Transactions) == 0 {
			t.Errorf("Transaction not received by peer")
		}
	})

}
