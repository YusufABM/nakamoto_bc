package block

import (
	"HAND_IN_2/account"
	//"HAND_IN_2/rsa"
	//"encoding/json"
	//"fmt"
	"sync"
	//"time"
)

// Block is a struct that contains the previous hash, the hash, the nonce, the transactions and the timestamp
type Block struct {
	PrevHash     string
	Hash         string
	Transactions []account.SignedTransaction
	slot         int
	maxBlockSize int
}

// Blockchain is a struct that contains the blocks and the ledger
type Blockchain struct {
	Blocks        []Block
	Ledger        *account.Ledger
	genesisLedger *account.Ledger
	seed          int
	lock          sync.Mutex
}

// NewBlockchain creates a new instance of Blockchain with a genesis block where 10 accounts are created with 1000000 in each
func NewBlockchain(ledger *account.Ledger) *Blockchain {
	blockchain := new(Blockchain)
	blockchain.Ledger = ledger
	blockchain.genesisLedger = ledger
	blockchain.seed = 42
	blockchain.Blocks = make([]Block, 0)
	return blockchain
}

// add signed transaction to a block
func (block *Block) AddTransaction(st *account.SignedTransaction) {
	block.Transactions = append(block.Transactions, *st)
}
