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

	ac := account.Transaction{
		ID:     "1",
		From:   "Peer1",
		To:     "Peer2",
		Amount: 10,
	}

	peer1.FloodTransaction(ac)

	time.Sleep(1 * time.Second)

	fmt.Println("Peer1 ledger:", ledger1.Accounts)
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
