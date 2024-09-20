package peer

import (
	"HAND_IN_2/account"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Peer is a struct that contains the IP address of the peer and the ledger
type Peer struct {
	Port   int
	ledger *account.Ledger
	open   bool
	Ports  []int
	name   string
}

// NewPeer creates a new instance of Peer
func NewPeer(port int, ledger *account.Ledger, name string) *Peer {
	peer := new(Peer)
	peer.Port = port
	peer.ledger = ledger
	peer.Ports = []int{}
	peer.name = name
	peer.open = false
	return peer
}

// Connects peer to a peer
func (peer *Peer) Connect(addr string, port int) {
	address := fmt.Sprintf("%s:%d", addr, port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		fmt.Println("Error connecting to peer, starting new network")
		peer.StartNewNetwork()
	}
	if peer.open == false {
		go peer.StartNewNetwork()
		peer.open = true
	}

	go peer.HandleConnection(conn)
	peer.Ports = append(peer.Ports, port)
	fmt.Printf("%s with port: %d, Connected to peer on port: %d\n", peer.name, peer.Port, port)
}

type Message struct {
	Test  string
	Ports []int
	Port  int
}

// HandleConnection handles incoming connections from peers
func (peer *Peer) HandleConnection(conn net.Conn) {
	msg1 := Message{peer.name, peer.Ports, peer.Port}

	defer conn.Close()
	b, err := json.Marshal(msg1)
	if err != nil {
		fmt.Println("Error marshalling ports:", err)
	}
	conn.Write(b)

	for {
		var msg Message
		ReceivedMessage := make([]byte, 1024)
		n, errcon := conn.Read(ReceivedMessage)
		if errcon != nil {
			fmt.Println("Error reading message:", errcon)
		}
		//err = json.Unmarshal([]byte(m), &msg)
		err = json.Unmarshal(ReceivedMessage[:n], &msg)
		if err != nil {
			fmt.Println("Error unmarshalling message:", err)
			return
		}
		fmt.Println("Received message:", msg)
		fmt.Println(peer.name, "'s Ports ", msg.Ports)
		peer.Ports = append(peer.Ports, msg.Port)
		time.Sleep(1 * time.Second)

	}
}

// StartNewNetwork starts a new network with the peer itself as the only member
func (p *Peer) StartNewNetwork() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	fmt.Println("Started new network on", "127.0.0.1", ":", p.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		//p.Ports = append(p.Ports, port)
		fmt.Printf("%s connected to ports: %v\n", p.name, p.Ports)

		go p.HandleConnection(conn)
	}
}
