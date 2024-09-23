package main

import (
	"HAND_IN_2/account"
	"HAND_IN_2/peer"
	"fmt"
	"log"
	"net"
	"time"
)

// GetOutboundIP preferred outbound ip of this machine
// based on code taken from https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go/37382208#37382208

func main() {
	ip := GetOutboundIP()
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
	peer1 := peer.NewPeer(8091, ledger1, "Peer1", ip)
	peer2 := peer.NewPeer(8092, ledger2, "Peer2", ip)
	peer3 := peer.NewPeer(8093, ledger3, "Peer3", ip)
	peer4 := peer.NewPeer(8094, ledger4, "Peer4", ip)
	peer5 := peer.NewPeer(8095, ledger5, "Peer5", ip)
	peer6 := peer.NewPeer(8096, ledger6, "Peer6", ip)
	peer7 := peer.NewPeer(8097, ledger7, "Peer7", ip)
	peer8 := peer.NewPeer(8098, ledger8, "Peer8", ip)
	peer9 := peer.NewPeer(8099, ledger9, "Peer9", ip)
	peer10 := peer.NewPeer(8100, ledger10, "Peer10", ip)

	go peer1.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer2.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer2.AskForPeers(8091)
	go peer3.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer3.AskForPeers(8091)
	go peer4.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer4.AskForPeers(8091)
	go peer5.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer5.AskForPeers(8091)
	go peer6.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer6.AskForPeers(8091)
	go peer7.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer7.AskForPeers(8091)
	go peer8.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer8.AskForPeers(8091)
	go peer9.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer9.AskForPeers(8091)
	go peer10.Connect(ip, 8091)
	time.Sleep(500 * time.Millisecond)
	go peer10.AskForPeers(8091)

	time.Sleep(3 * time.Second)

	ac1 := account.Transaction{
		ID:     "1",
		From:   "account1",
		To:     "account2",
		Amount: 823,
	}
	peer1.ExecuteTransaction(ac1)
	ac2 := account.Transaction{
		ID:     "2",
		From:   "account2",
		To:     "account4",
		Amount: 129,
	}
	peer1.ExecuteTransaction(ac2)
	ac3 := account.Transaction{
		ID:     "3",
		From:   "account3",
		To:     "account4",
		Amount: 398,
	}
	peer1.ExecuteTransaction(ac3)
	ac4 := account.Transaction{
		ID:     "4",
		From:   "account1",
		To:     "account4",
		Amount: 989,
	}
	peer1.ExecuteTransaction(ac4)
	ac5 := account.Transaction{
		ID:     "5",
		From:   "account5",
		To:     "account2",
		Amount: 321,
	}
	peer1.ExecuteTransaction(ac5)
	ac6 := account.Transaction{
		ID:     "6",
		From:   "account3",
		To:     "account5",
		Amount: 590,
	}
	peer1.ExecuteTransaction(ac6)
	ac7 := account.Transaction{
		ID:     "7",
		From:   "account4",
		To:     "account5",
		Amount: 147,
	}
	peer1.ExecuteTransaction(ac7)
	ac8 := account.Transaction{
		ID:     "8",
		From:   "account1",
		To:     "account3",
		Amount: 289,
	}
	peer1.ExecuteTransaction(ac8)
	ac9 := account.Transaction{
		ID:     "9",
		From:   "account5",
		To:     "account2",
		Amount: 900,
	}
	peer1.ExecuteTransaction(ac9)
	ac10 := account.Transaction{
		ID:     "10",
		From:   "account2",
		To:     "account3",
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

	fmt.Println("Ledger 1 has the following accounts: ")

	for account := range ledger1.Accounts {
		fmt.Println(account, ledger1.Accounts[account])
	}
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	hostip, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		log.Fatal(err)
	}

	return hostip
}
