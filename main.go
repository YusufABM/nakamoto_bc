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
	// Create a slice of ports

	// Create a slice of peers
	peers := []*peer.Peer{peer1, peer2, peer3, peer4, peer5, peer6, peer7, peer8, peer9, peer10}

	// Connect peers to the network
	// Generate random ports for peers
	for i := 0; i < len(peers); i++ {
		port, err := GetFreePort()
		if err != nil {
			log.Fatalf("Failed to get a free port: %v", err)
		}
		peers[i].Port = port
	}

	go peer1.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer2.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer2.AskForPeers(peer1.Port)
	go peer3.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer3.AskForPeers(peer1.Port)
	go peer4.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer4.AskForPeers(peer1.Port)
	go peer5.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer5.AskForPeers(peer1.Port)
	go peer6.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer6.AskForPeers(peer1.Port)
	go peer7.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer7.AskForPeers(peer1.Port)
	go peer8.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer8.AskForPeers(peer1.Port)
	go peer9.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer9.AskForPeers(peer1.Port)
	go peer10.Connect(ip, peer1.Port)
	time.Sleep(500 * time.Millisecond)
	go peer10.AskForPeers(peer1.Port)

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
