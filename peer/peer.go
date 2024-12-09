package peer

import (
	"HAND_IN_2/account"
	"HAND_IN_2/block"
	"HAND_IN_2/rsa"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"time"
)

var HARDNESS int = 4
var SLOTLENGTH int = 1

// Peer is a struct that contains the IP address of the peer and the ledger
type Peer struct {
	Port         int
	Ledger       *account.Ledger
	open         bool
	name         string
	ip           string
	connections  map[int]net.Conn
	Account      account.Account
	mu           sync.Mutex
	Blockchain   block.Blockchain
	Transactions []account.SignedTransaction
}

type Message struct {
	Action       string
	Ports        []int
	SendingPort  int
	St           account.SignedTransaction
	LotteryBlock block.Lottery
}

// NewPeer creates a new instance of Peer
func NewPeer(port int, ledger *account.Ledger, name string, ip string, account1 *account.Account, time time.Time) *Peer {
	peer := new(Peer)
	peer.Port = port
	peer.Account = *account1
	peer.Blockchain = *block.NewBlockchain(ledger, time)
	peer.name = name
	peer.open = false
	peer.ip = ip
	peer.connections = make(map[int]net.Conn)
	peer.connections[port] = nil
	//sets peer.transactions to an empty slice
	peer.Transactions = make([]account.SignedTransaction, 0)
	return peer
}

// Connects peer to a peer
// If the peer is already connected to the peer, it returns the connection
// If the peer is not connected to the peer, it connects to the peer
// If the peer is not connected to any peers, it starts a new network
func (peer *Peer) Connect(addr string, port int) net.Conn {

	if conn, exists := peer.connections[port]; exists {
		fmt.Printf(" %s Already connected to peer on port: %d\n", peer.name, port)
		if conn == nil && !peer.open {
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
		ReceivedMessage := make([]byte, 8192*64)
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
		if !peer.KnownPeer(msg.SendingPort) && msg.SendingPort != peer.Port {
			conn := peer.Connect(peer.ip, msg.SendingPort)
			if conn != nil {
				go peer.HandleConnection(conn)
				peer.addPeer(msg.SendingPort, conn)
			}
		}
	}
	//a peer asks for peers
	//we send the ports to the new peer
	if msg.Action == "askForPeers" {
		conn := peer.Connect(peer.ip, msg.SendingPort)
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
		msg := Message{Action: "join", Ports: peer.getPorts(), SendingPort: peer.Port}
		peer.floodMessage(msg)
	}
	if msg.Action == "transaction" {
		peer.ExecuteTransaction(msg.St)
	}
	if msg.Action == "block" {
		fmt.Println("Received block from peer:")
		lotteryBlock := msg.LotteryBlock
		peer.Blockchain.ProcessLotteryBlock(lotteryBlock)
	}
	if msg.Action == "lottery" {
		pk := msg.LotteryBlock.Pk
		slotNum := msg.LotteryBlock.Slot
		draw := msg.LotteryBlock.Draw
		verifyDraw := peer.verifyDraw(draw, slotNum, pk)

		if verifyDraw {
			hash := peer.HashTicket(slotNum, pk, draw)
			stake := peer.Blockchain.GenesisLedger.Accounts[rsa.EncodePublicKey(pk)]
			var num uint64
			err := binary.Read(bytes.NewReader(hash), binary.BigEndian, &num)
			if err != nil {
				fmt.Println("binary.Read failed:", err)
			}
			tickets := int(math.Abs(float64(int(num)*stake))) / 10000000
			if tickets > block.HARDNESS {
				fmt.Println("received Lottery Block created")
				peer.Blockchain.ProcessLotteryBlock(msg.LotteryBlock)
				peer.Blockchain.AddMinerReward(msg.LotteryBlock)
				peer.DeleteTransactions()
			} else {
				fmt.Println("Block not created invalid tickets")
			}
		} else {
			fmt.Println("Invalid signature on draw")
		}
	}
}

func (p *Peer) verifyDraw(signature []byte, slot []byte, pk rsa.PublicKey) bool {
	data := "lottery" + strconv.Itoa(p.Blockchain.Seed) + string(slot)
	return rsa.VerifySignature([]byte(data), signature, pk)
}

func (p *Peer) lottery() {
	time.Sleep(5 * time.Second)
	// Create a ticker that triggers every second
	ticker := time.NewTicker(time.Duration(block.SLOTLENGTH) * time.Second)
	defer ticker.Stop() // Clean up the ticker when the program exits

	// Use a channel to signal program termination
	done := make(chan bool)

	// Start a goroutine to handle the ticks
	go func() {
		for {
			select {
			case <-done:
				return // Exit the goroutine when signaled
			case t := <-ticker.C:
				// Code to run every second
				slotNum := []byte(strconv.Itoa(p.CurrenetSlotNum()))
				draw := p.signDraw(slotNum)
				hash := p.HashTicket(slotNum, p.Account.Pk, draw)
				stake := p.Blockchain.GenesisLedger.Accounts[rsa.EncodePublicKey(p.Account.Pk)]
				var num uint64
				err := binary.Read(bytes.NewReader(hash), binary.BigEndian, &num)
				if err != nil {
					fmt.Println("binary.Read failed:", err)
				}
				tickets := int(math.Abs(float64(int(num)*stake))) / 10000000

				if tickets > block.HARDNESS {
					fmt.Println("Block created")
					fmt.Println("Ticket:", tickets)
					fmt.Println("Peer:", p.name)
					fmt.Println("time:", t)
					fmt.Println("slot:", p.CurrenetSlotNum())
					transactionsToSend := int(math.Min(float64(p.Blockchain.BlockSize), float64(len(p.Transactions))))
					winnerTransactions := p.Transactions[:transactionsToSend]
					winnerBlock := block.NewBlock(&p.Blockchain.ChainHead, winnerTransactions, p.Account.Pk)
					lotteryBlock := block.NewLotteryBlock(*winnerBlock, p.Account.Pk, p.Account.Sk, slotNum, draw)
					p.SendLottery(*lotteryBlock)
					p.Blockchain.AddMinerReward(*lotteryBlock)
					p.Blockchain.ProcessLotteryBlock(*lotteryBlock)
					p.DeleteTransactions()
				}
			}
		}
	}()

	// Simulate running the program for 10 seconds
	time.Sleep(100 * time.Second)
	fmt.Println("Program finished")
}

// FloodTransaction sends a transaction to all known peers
func (peer *Peer) FloodTransaction(st account.SignedTransaction) {
	//flood transaction
	for port, conn := range peer.connections {
		if conn != nil && port != peer.Port {
			msg := Message{Action: "transaction", SendingPort: peer.Port, St: st}
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
	peer.Transactions = append(peer.Transactions, st)
}

// delete blocksize transactions
func (peer *Peer) DeleteTransactions() {
	length := len(peer.Transactions)
	if length > peer.Blockchain.BlockSize {
		length = peer.Blockchain.BlockSize
	}

	peer.Transactions = peer.Transactions[length:]
}

// AskForPeers asks a specific peer for its peers
// has to be a known peer
func (peer *Peer) AskForPeers(port int) {
	conn := peer.connections[port]
	if conn != nil {
		msg := Message{Action: "askForPeers", SendingPort: peer.Port}
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

func (p *Peer) signDraw(slot []byte) []byte {
	data := "lottery" + strconv.Itoa(p.Blockchain.Seed) + string(slot)

	return rsa.SignMessage([]byte(data), p.Account.Sk)
}

func (p *Peer) HashTicket(slotNum []byte, pk rsa.PublicKey, signature []byte) []byte {

	data := strconv.Itoa(p.Blockchain.Seed) + string(slotNum) + string(rsa.EncodePublicKey(pk)) + string(signature)
	//hash the data
	hash := sha256.New()
	hash.Write([]byte(data))
	hashedMessage := hash.Sum(nil)

	return hashedMessage
}

// GenerateLotterySignature creates a signature for a slot
func (p *Peer) GenerateLotterySignature() []byte {
	slot := p.CurrenetSlotNum()
	Sk := p.Account.Sk
	return rsa.SignMessage([]byte(strconv.Itoa(slot)), Sk)
}

// CurrenetSlotNum returns the current slot number as a byte array
func (peer *Peer) CurrenetSlotNum() int {
	GenTime := peer.Blockchain.StartTime
	elapsed := time.Since(GenTime).Seconds()
	return int(elapsed) / SLOTLENGTH
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
	go peer.lottery()
	peer.open = true
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
	msg1 := Message{Action: "peers", Ports: peer.getPorts(), SendingPort: peer.Port}

	b, err := json.Marshal(msg1)
	if err != nil {
		fmt.Println("Error marshalling ports:", err)
	}
	conn.Write(b)
}

func (peer *Peer) SendBlockToPeers(lotteryBlock block.Lottery) {
	msg := Message{Action: "block", SendingPort: peer.Port, LotteryBlock: lotteryBlock}
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}
	for _, conn := range peer.connections {
		if conn != nil {
			_, err = conn.Write(b)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				continue
			}
		}
	}
	fmt.Println("Block sent to peers")
}

func (peer *Peer) SendLottery(lotteryBlock block.Lottery) {
	msg := Message{Action: "lottery", SendingPort: peer.Port, LotteryBlock: lotteryBlock}
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}
	for _, conn := range peer.connections {
		if conn != nil {
			_, err = conn.Write(b)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				continue
			}
		}
	}
	fmt.Println("Lottery sent to peers")
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
