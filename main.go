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
	fmt.Println("IP:", ip)
	//initialize ledgers and peers
	ledger1 := account.MakeLedger()
	ledger2 := account.MakeLedger()
	ledger3 := account.MakeLedger()
	ledger4 := account.MakeLedger()
	ledger5 := account.MakeLedger()

	peer1 := peer.NewPeer(8091, ledger1, "Peer1", ip)
	peer2 := peer.NewPeer(8092, ledger2, "Peer2", ip)
	peer3 := peer.NewPeer(8093, ledger3, "Peer3", ip)
	peer4 := peer.NewPeer(8094, ledger4, "Peer4", ip)
	peer5 := peer.NewPeer(8095, ledger5, "Peer5", ip)

	//peers := []*peer.Peer{peer1, peer2, peer3}

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

	time.Sleep(3 * time.Second)

	ac := account.Transaction{
		ID:     "First",
		From:   "Brain",
		To:     "Paul",
		Amount: 834,
	}
	peer1.ExecuteTransaction(ac)
	peer1.FloodTransaction(ac)

	time.Sleep(1 * time.Second)

	fmt.Println("Peer1 ledger:", ledger1.Accounts)
	fmt.Println("Peer2 ledger:", ledger2.Accounts)
	fmt.Println("Peer3 ledger:", ledger3.Accounts)
	fmt.Println("Peer4 ledger:", ledger4.Accounts)
	fmt.Println("Peer5 ledger:", ledger5.Accounts)

	select {}
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
