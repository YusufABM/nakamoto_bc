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

}
