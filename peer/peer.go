package peer

import (
	"HAND_IN_2/account"
	"HAND_IN_2/block"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

var HARDNESS int = 4
var SLOTLENGTH int = 10

// Peer is a struct that contains the IP address of the peer and the ledger
type Peer struct {
	Port        int
	Ledger      *account.Ledger
	open        bool
	name        string
	ip          string
	connections map[int]net.Conn
	Account     account.Account
	mu          sync.Mutex
}

type Message struct {
	Action string
	Ports  []int
	Port   int
	St     account.SignedTransaction
	block  block.Block
}

// NewPeer creates a new instance of Peer
func NewPeer(port int, ledger *account.Ledger, name string, ip string, account *account.Account) *Peer {
	peer := new(Peer)
	peer.Port = port
	peer.Ledger = ledger
	peer.Account = *account
	peer.name = name
	peer.open = false
	peer.ip = ip
	peer.connections = make(map[int]net.Conn)
	peer.connections[port] = nil
	return peer
}

// Connects peer to a peer
// If the peer is already connected to the peer, it returns the connection
// If the peer is not connected to the peer, it connects to the peer
// If the peer is not connected to any peers, it starts a new network
func (peer *Peer) Connect(addr string, port int) net.Conn {

	if conn, exists := peer.connections[port]; exists {
		fmt.Printf(" %s Already connected to peer on port: %d\n", peer.name, port)
		if conn == nil && peer.open == false {
			go peer.StartNewNetwork()
			peer.open = true
		} else {
			return conn
		}
	}

	address := fmt.Sprintf("%s:%d", addr, port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		if !peer.open {
			go peer.StartNewNetwork()
			peer.open = true
		}
		return nil
	} else {
		peer.addPeer(port, conn)
		if len(peer.connections) == 5 {
			fmt.Println("Max number of connections reached for peer:", peer.name)
		}
	}

	if !peer.open {
		go peer.StartNewNetwork()
		peer.open = true
	}

	fmt.Printf("%s with port: %d, Connected to peer on port: %d\n", peer.name, peer.Port, port)
	return conn
}

// HandleConnection handles incoming connections from peers
// It reads the message and calls handleMessage on the msg body
func (peer *Peer) HandleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		var msg Message
		ReceivedMessage := make([]byte, 8192)
		n, errcon := conn.Read(ReceivedMessage)
		if errcon != nil {
			fmt.Println("Error reading message:", errcon)
			return
		}
		//fmt.Println("Received raw message:", string(ReceivedMessage[:n])) // Print the raw message
		err := json.Unmarshal(ReceivedMessage[:n], &msg)
		if err != nil {
			fmt.Println("Error unmarshalling message:", err)
			return
		}
		//fmt.Println("Received message:", msg.Action)
		go peer.handleMessage(msg)
	}
}

func (peer *Peer) handleMessage(msg Message) {
	//a peer joins the network
	//we connect to the new peer
	if msg.Action == "join" {
		if !peer.KnownPeer(msg.Port) && msg.Port != peer.Port {
			conn := peer.Connect(peer.ip, msg.Port)
			if conn != nil {
				go peer.HandleConnection(conn)
				peer.addPeer(msg.Port, conn)
			}
		}
	}
	//a peer asks for peers
	//we send the ports to the new peer
	if msg.Action == "askForPeers" {
		conn := peer.Connect(peer.ip, msg.Port)
		if conn != nil {
			peer.sendPeersToNewPeer(conn)
			fmt.Println("Sent ports to new peer")
			go peer.HandleConnection(conn)
		}
	}
	//a peer sends a list of ports
	//we then flood a message to all the peers
	//for each port, if we don't know the peer, we connect to it
	if msg.Action == "peers" {
		for _, port := range msg.Ports {
			if !peer.KnownPeer(port) && port != peer.Port {
				fmt.Printf("Peer: %s added %d to known peers \n", peer.name, port)

				conn := peer.Connect(peer.ip, port)
				if conn != nil {
					go peer.HandleConnection(conn)
				}
			}
		}
		msg := Message{Action: "join", Ports: peer.getPorts(), Port: peer.Port}
		peer.floodMessage(msg)
	}
	if msg.Action == "transaction" {
		peer.ExecuteTransaction(msg.St)
	}
	if msg.Action == "block" {

		//peer.Ledger.ProcessBlock(msg.St)
	}
}

// FloodTransaction sends a transaction to all known peers
func (peer *Peer) FloodTransaction(st account.SignedTransaction) {
	//flood transaction
	for port, conn := range peer.connections {
		if conn != nil && port != peer.Port {
			msg := Message{Action: "transaction", Ports: nil, Port: peer.Port, St: st}
			//fmt.Printf("Message: %v \n", msg)
			b, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Error marshalling message:", err)
				continue
			}
			_, err = conn.Write(b)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				continue
			}
		}
	}
}

// ExecuteTransaction executes a transaction
// takes a transaction and updates the ledger
func (peer *Peer) ExecuteTransaction(st account.SignedTransaction) {
	//fmt.Printf("length of %s 's connections: %d \n", peer.name, len(peer.connections))
	//fmt.Print(peer.getPorts())
	//fmt.Printf("from peer %s Executing transaction \n", peer.name)
	peer.Ledger.ProcessSignedTransaction(&st)
	//fmt.Println(peer.ledger.Accounts)
	//fmt.Println("From", st.From, "to", st.To, "Amount:", st.Amount)
}

// AskForPeers asks a specific peer for its peers
// has to be a known peer
func (peer *Peer) AskForPeers(port int) {
	conn := peer.connections[port]
	if conn != nil {
		msg := Message{"askForPeers", nil, peer.Port, account.SignedTransaction{}, block.Block{}}
		b, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Error marshalling message:", err)
		}
		_, err = conn.Write(b)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
		}
	}

}

// StartNewNetwork starts a new network with the peer itself as the only member
func (peer *Peer) StartNewNetwork() {
	fmt.Println("Starting new network on port", peer.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", peer.Port))
	if err != nil {
		fmt.Println("can't connect to port", peer.Port)
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go peer.HandleConnection(conn)
	}
}

func (peer *Peer) sendPeersToNewPeer(conn net.Conn) {
	msg1 := Message{"peers", peer.getPorts(), peer.Port, account.SignedTransaction{}, block.Block{}}

	b, err := json.Marshal(msg1)
	if err != nil {
		fmt.Println("Error marshalling ports:", err)
	}
	conn.Write(b)
}

func (peer *Peer) KnownPeer(port int) bool {
	for k := range peer.getPorts() {
		if k == port {
			return true
		}
	}
	return false
}

func (peer *Peer) getPorts() []int {
	keys := make([]int, 0, len(peer.connections))
	for k := range peer.connections {
		keys = append(keys, k)
	}
	return keys
}

// Get amount of connections
func (peer *Peer) GetAmountOfConnections() int {
	return len(peer.connections)
}

func (peer *Peer) floodMessage(msg Message) {
	for _, conn := range peer.connections {
		if conn != nil {
			b, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Error marshalling message:", err)
				continue
			}
			_, err = conn.Write(b)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				continue
			}
		}
	}
}

func (peer *Peer) addPeer(port int, conn net.Conn) {
	peer.connections[port] = conn
}
