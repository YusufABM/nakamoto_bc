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
	conn   net.Conn
	Ports  []int
	name   string
}

// NewPeer creates a new instance of Peer
func NewPeer(port int, ledger *account.Ledger, name string) *Peer {
	peer := new(Peer)
	peer.Port = port
	peer.ledger = ledger
	peer.Ports = []int{port}
	peer.name = name
	return peer
}

// Connects peer to a peer
func (peer *Peer) Connect(addr string, port int) {
	address := fmt.Sprintf("%s:%d", addr, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Println("Error connecting to peer: ", err)
		peer.StartNewNetwork()
	}
	peer.conn = conn
	fmt.Printf("%s Connected to peer on port: %d\n", peer.name, port)
}

// HandleConnection handles incoming connections from peers
func HandleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var msg account.Transaction
	for {
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Error decoding message:", err)
			return
		}
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
		go HandleConnection(conn)
	}
}
