package block

import (
	"HAND_IN_2/account"
	"HAND_IN_2/rsa"
	"crypto/sha256"
	"fmt"
	"time"
	//"encoding/json"
	//"fmt"
)

const SLOTLENGTH int = 1

// Block is a struct that contains the previous hash, the hash, the nonce, the transactions and the timestamp
type Block struct {
	PrevHash     string
	Transactions []account.SignedTransaction
	timeStamp    time.Time
}

// Blockchain is a struct that contains the blocks and the ledger
type Blockchain struct {
	Blocks        []Block
	Ledger        *account.Ledger
	genesisLedger *account.Ledger
	seed          int
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

// hashes a block
func (block *Block) HashBlock(vk account.Account) string {
	blockData := rsa.EncodePublicKey(vk.Pk) + block.PrevHash + fmt.Sprintf("%d", (block.timeStamp.Second()/SLOTLENGTH)+1) + account.EncodeTransactions(block.Transactions) + block.timeStamp.String()
	hash := sha256.New()
	hash.Write([]byte(blockData))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// signs a block
func (block *Block) SignBlock(key account.Account) []byte {
	blockData := rsa.EncodePublicKey(key.Pk) + block.PrevHash + fmt.Sprintf("%d", (block.timeStamp.Second()/SLOTLENGTH)+1) + account.EncodeTransactions(block.Transactions) + block.timeStamp.String() + block.HashBlock(key)
	signedMessage := rsa.SignMessage([]byte(blockData), key.Sk)
	return signedMessage
}

// verifies a block
func (block *Block) VerifyBlock(key account.Account, signature []byte) bool {
	blockData := rsa.EncodePublicKey(key.Pk) + block.PrevHash + account.EncodeTransactions(block.Transactions) + block.timeStamp.String() + block.HashBlock(key)
	return rsa.VerifySignature([]byte(blockData), signature, key.Pk)
}
