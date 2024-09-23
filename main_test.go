package main

import (
	"HAND_IN_2/account"
	"HAND_IN_2/peer"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func GetOutboundIP2() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func Test1(t *testing.T) {
	fmt.Println("Test1")
	ip := GetOutboundIP2()
	//initialize ledgers and peers
	ledger1 := account.MakeLedger()
	ledger2 := account.MakeLedger()
	ledger3 := account.MakeLedger()
	ledger4 := account.MakeLedger()
	ledger5 := account.MakeLedger()
	ledger6 := account.MakeLedger()
	ledger7 := account.MakeLedger()
	ledger8 := account.MakeLedger()
	ledger9 := account.MakeLedger()
	ledger10 := account.MakeLedger()
	peer1 := peer.NewPeer(8091, ledger1, "Peer1", ip.String())
	peer2 := peer.NewPeer(8092, ledger2, "Peer2", ip.String())
	peer3 := peer.NewPeer(8093, ledger3, "Peer3", ip.String())
	peer4 := peer.NewPeer(8094, ledger4, "Peer4", ip.String())
	peer5 := peer.NewPeer(8095, ledger5, "Peer5", ip.String())
	peer6 := peer.NewPeer(8096, ledger6, "Peer6", ip.String())
	peer7 := peer.NewPeer(8097, ledger7, "Peer7", ip.String())
	peer8 := peer.NewPeer(8098, ledger8, "Peer8", ip.String())
	peer9 := peer.NewPeer(8099, ledger9, "Peer9", ip.String())
	peer10 := peer.NewPeer(8100, ledger10, "Peer10", ip.String())

	go peer1.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer2.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer2.AskForPeers(8091)
	go peer3.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer3.AskForPeers(8091)
	go peer4.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer4.AskForPeers(8091)
	go peer5.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer5.AskForPeers(8091)
	go peer6.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer6.AskForPeers(8091)
	go peer7.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer7.AskForPeers(8091)
	go peer8.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer8.AskForPeers(8091)
	go peer9.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer9.AskForPeers(8091)
	go peer10.Connect(ip.String(), 8091)
	time.Sleep(500 * time.Millisecond)
	go peer10.AskForPeers(8091)

	time.Sleep(3 * time.Second)

	//Test if peer5 has connected to all other peers
	t.Run("Connections", func(t *testing.T) {
		if peer5.GetAmountOfConnections() != 10 {
			t.Errorf("Expected more than 10 connections, got %d", peer5.GetAmountOfConnections())
		}
	})

	ac1 := account.Transaction{
		ID:     "1",
		From:   "Brain",
		To:     "Paul",
		Amount: 823,
	}
	peer1.ExecuteTransaction(ac1)
	ac2 := account.Transaction{
		ID:     "2",
		From:   "Brain",
		To:     "Felix",
		Amount: 129,
	}
	peer1.ExecuteTransaction(ac2)
	ac3 := account.Transaction{
		ID:     "3",
		From:   "Thomas",
		To:     "Felix",
		Amount: 398,
	}
	peer1.ExecuteTransaction(ac3)
	ac4 := account.Transaction{
		ID:     "4",
		From:   "Paul",
		To:     "Felix",
		Amount: 989,
	}
	peer1.ExecuteTransaction(ac4)
	ac5 := account.Transaction{
		ID:     "5",
		From:   "Yusuf",
		To:     "Brain",
		Amount: 321,
	}
	peer1.ExecuteTransaction(ac5)
	ac6 := account.Transaction{
		ID:     "6",
		From:   "Thomas",
		To:     "Yusuf",
		Amount: 590,
	}
	peer1.ExecuteTransaction(ac6)
	ac7 := account.Transaction{
		ID:     "7",
		From:   "Felix",
		To:     "Yusuf",
		Amount: 147,
	}
	peer1.ExecuteTransaction(ac7)
	ac8 := account.Transaction{
		ID:     "8",
		From:   "Paul",
		To:     "Thomas",
		Amount: 289,
	}
	peer1.ExecuteTransaction(ac8)
	ac9 := account.Transaction{
		ID:     "9",
		From:   "Yusuf",
		To:     "Paul",
		Amount: 900,
	}
	peer1.ExecuteTransaction(ac9)
	ac10 := account.Transaction{
		ID:     "10",
		From:   "Brain",
		To:     "Thomas",
		Amount: 540,
	}
	peer1.ExecuteTransaction(ac10)
	peer1.FloodTransaction(ac1)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac2)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac3)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac4)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac5)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac6)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac7)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac8)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac9)
	time.Sleep(50 * time.Millisecond)
	peer1.FloodTransaction(ac10)
	time.Sleep(50 * time.Millisecond)

	//Test if the transaction is executed correctly
	/*
		t.Run("Transaction", func(t *testing.T) {
			if ledger1.Accounts["Paul"] != 823 {
				t.Errorf("Expected 823, got %d", ledger1.Accounts["Paul"])
			}
		})
	*/
	//Test if the transaction is flooded correctly
	t.Run("Flooding", func(t *testing.T) {
		time.Sleep(1 * time.Second)
		if !reflect.DeepEqual(ledger1.Accounts, ledger2.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger3.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger4.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger5.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger6.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger7.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger8.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger9.Accounts) || !reflect.DeepEqual(ledger1.Accounts, ledger10.Accounts) {
			t.Errorf("Expected all ledgers to be equal, got %v\n, %v\n, %v\n, %v\n, %v\n, %v\n, %v\n, %v\n, %v\n, %v\n", ledger1.Accounts, ledger2.Accounts, ledger3.Accounts, ledger4.Accounts, ledger5.Accounts, ledger6.Accounts, ledger7.Accounts, ledger8.Accounts, ledger9.Accounts, ledger10.Accounts)
		}

	})

}
