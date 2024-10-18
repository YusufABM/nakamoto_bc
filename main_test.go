package main

import (
	"HAND_IN_2/account"
	"HAND_IN_2/peer"
	"HAND_IN_2/rsa"
	"fmt"
	"log"
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var ledger = make([]*account.Ledger, 10)
var accounts = make([]*account.Account, 10)
var ports = []int{}
var peers = make([]*peer.Peer, 10)
var counter = 0

func GetOutboundIP2() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

// Random transaction generator

func randomTransaction(peer *peer.Peer, counter *int) {
	fromNumber := rand.Intn(6-1) + 1
	toNumber := rand.Intn(6-1) + 1
	for fromNumber == toNumber {
		toNumber = rand.Intn(6-1) + 1
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

func transactionGenerator() {
	id := "1"
	FromAccount := accounts[1]
	ToAccount := accounts[2]

	pkFrom := rsa.EncodePublicKey(FromAccount.Pk)
	pkTo := rsa.EncodePublicKey(ToAccount.Pk)
	ac := account.Transaction{ID: id, From: pkFrom, To: pkTo, Amount: 100}
	fmt.Println("pk1: ", FromAccount.Pk)
	st := account.SignTransaction(FromAccount.Sk, &ac)
	peers[1].ExecuteTransaction(st)
	peers[1].FloodTransaction(st)
}

func Test(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	//counter := 0
	fmt.Println("Test1")
	ip := GetOutboundIP2()

	//initialize ledger
	for i := 0; i < 10; i++ {
		ledger[i] = account.MakeLedger()
	}

	//initialize accounts
	for i := 0; i < 10; i++ {
		accounts[i] = account.MakeAccount()
	}

	// Generate random ports for peers
	for i := 0; i < 10; i++ {
		time.Sleep(50 * time.Millisecond)
		port, err := GetFreePort()
		if err != nil {
			log.Fatalf("Failed to get a free port: %v", err)
		}
		ports = append(ports, port)
	}

	//Initialize peers
	for i := 0; i < 10; i++ {
		peers[i] = peer.NewPeer(ports[i], ledger[i], "Peer"+strconv.Itoa(i), ip.String())
	}

	// Connect peers to the network
	for i := 0; i < 10; i++ {
		go peers[i].Connect(ip.String(), peers[0].Port)
		time.Sleep(500 * time.Millisecond)
		go peers[i].AskForPeers(peers[0].Port)
	}
	time.Sleep(3 * time.Second)

	//Test if peer5 has connected to all other peers
	t.Run("Connections", func(t *testing.T) {
		if peers[4].GetAmountOfConnections() != 10 {
			t.Errorf("Expected more than 10 connections, got %d", peers[4].GetAmountOfConnections())
		}
	})
	/*
		// Runs 10 transactions on all peers at the same time
		for i := 0; i < 5; i++ {
			go randomTransaction(peers[i], &counter)
			time.Sleep(200 * time.Millisecond)
		}

		//Test if the transaction is flooded correctly
		t.Run("Flooding", func(t *testing.T) {
			time.Sleep(1 * time.Second)
			for i := 1; i < 10; i++ {
				if !reflect.DeepEqual(ledger[0].Accounts, ledger[i].Accounts) {
					t.Errorf("Expected all ledgers to be equal, got %v\n, %v\n", ledger[0].Accounts, ledger[i].Accounts)
				}
			}
		})
	*/

	transactionGenerator() // Runs 10 transactions on all peers at the same time
	t.Run("Flooding1Message", func(t *testing.T) {
		time.Sleep(1 * time.Second)
		for i := 1; i < 10; i++ {
			if !reflect.DeepEqual(ledger[0].Accounts, ledger[i].Accounts) {
				t.Errorf("Expected all ledgers to be equal")
				//t.Errorf("Expected all ledgers to be equal, got %v\n, %v\n", ledger[0].Accounts, ledger[i].Accounts)
			}
		}
	})

	//Test if public keys can be encoded and decoded
	t.Run("EncodeDecode", func(t *testing.T) {
		pk, _, _ := rsa.Keygen(2048)
		encodedKey := rsa.EncodePublicKey(pk)
		decodedKey := rsa.DecodePublicKey(encodedKey)
		if !reflect.DeepEqual(pk, decodedKey) {
			t.Errorf("Expected key to be equal, got %v\n, %v\n", pk, decodedKey)
		}
	})

	t.Run("signTransaction", func(t *testing.T) {
		ledger10 := account.MakeLedger()
		account1 := account.MakeAccount()
		account2 := account.MakeAccount()
		tx := account.Transaction{ID: "1", From: rsa.EncodePublicKey(account1.Pk), To: rsa.EncodePublicKey(account2.Pk), Amount: 100}
		st := account.SignTransaction(account1.Sk, &tx)
		ledger10.ProcessSignedTransaction(&st)
		if ledger10.Accounts[rsa.EncodePublicKey(account2.Pk)] != 100 {
			t.Errorf("Expected account2 to have 100, got %d", ledger10.Accounts[rsa.EncodePublicKey(account2.Pk)])
		}
	})

	t.Run("signWrongKeyTransaction", func(t *testing.T) {
		ledger10 := account.MakeLedger()
		account1 := account.MakeAccount()
		account2 := account.MakeAccount()
		tx := account.Transaction{ID: "1", From: rsa.EncodePublicKey(account1.Pk), To: rsa.EncodePublicKey(account2.Pk), Amount: 100}
		st := account.SignTransaction(account2.Sk, &tx)
		ledger10.ProcessSignedTransaction(&st)
		if ledger10.Accounts[rsa.EncodePublicKey(account2.Pk)] != 0 {
			t.Errorf("Expected account2 to have 0, got %d", ledger10.Accounts[rsa.EncodePublicKey(account2.Pk)])
		}
	})

}
