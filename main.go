package main

import (
	"HAND_IN_2/account"
	"HAND_IN_2/peer"
	"log"
	"net"
	"time"
)

// GetOutboundIP preferred outbound ip of this machine
// based on code taken from https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go/37382208#37382208
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

func main() {
	ip := GetOutboundIP()

	//initialize ledgers and peers
	ledger1 := account.MakeLedger()
	ledger2 := account.MakeLedger()
	ledger3 := account.MakeLedger()
	peer1 := peer.NewPeer(8081, ledger1, "Peer1")
	peer2 := peer.NewPeer(8082, ledger2, "Peer2")
	peer3 := peer.NewPeer(8083, ledger3, "Peer3")

	//peers := []*peer.Peer{peer1, peer2, peer3}

	go peer1.Connect(ip, 8081)
	time.Sleep(10 * time.Millisecond)
	go peer2.Connect(ip, 8081)
	go peer3.Connect(ip, 8081)

	select {}
}
