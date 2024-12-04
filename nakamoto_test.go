package main

import (
	"HAND_IN_2/account"
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

	t.Run("Connections", func(t *testing.T) {
		if peers[4].GetAmountOfConnections() != 10 {
			t.Errorf("Expected more than 10 connections, got %d", peers[4].GetAmountOfConnections())
		}
	})
}
