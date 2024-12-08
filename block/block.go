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
	Hash         string
	Transactions []account.SignedTransaction
	Height       int
	timeStamp    time.Time
}

// Blockchain is a struct that contains the blocks and the ledger
type Blockchain struct {
	Blocks        map[string]Block
	Ledger        *account.Ledger
	genesisLedger *account.Ledger
	seed          int
	ChainHead     Block
}

type Lottery struct {
	Block     *Block
	Slot      int
	Pk        rsa.PublicKey
	Draw      int
	signature []byte
}

// NewBlockchain creates a new instance of Blockchain with a genesis block where 10 accounts are created with 1000000 in each
func NewBlockchain(ledger *account.Ledger) *Blockchain {
	blockchain := new(Blockchain)
	blockchain.Ledger = ledger
	blockchain.genesisLedger = ledger.DeepCopy()
	blockchain.seed = 42
	blockchain.Blocks = make(map[string]Block)
	return blockchain
}

func NewBlock(parent *Block, transactions []account.SignedTransaction, pk rsa.PublicKey) *Block {
	block := new(Block)

	parentHash := ""
	block.Transactions = transactions
	block.timeStamp = time.Now()
	if parent != nil {
		block.Height = parent.Height + 1
		parentHash = parent.Hash
	} else {
		block.Height = 1
	}
	block.PrevHash = parentHash
	hash := block.HashBlock(pk)
	block.Hash = hash
	return block
}

// adds a block to the blockchain and makes block with biggest height chainhead
// Calls switchToChain if the new chain is longer
func (blockchain *Blockchain) AddBlock(block Block) {
	currentHeight := blockchain.ChainHead.Height
	newBlockHeight := block.Height
	if blockchain.ChainHead.Hash == block.PrevHash {
		blockchain.ChainHead = block
		blockchain.addTransactionsToLedger(block.Transactions)
	} else if newBlockHeight > currentHeight {
		blockchain.ChainHead = block
		blockchain.Ledger = blockchain.genesisLedger
		blockchain.switchToChain(blockchain.ChainHead)
	}
	blockchain.Blocks[block.Hash] = block
}

func (blockchain *Blockchain) switchToChain(block Block) {
	if block.PrevHash == "" {
		blockchain.addTransactionsToLedger(block.Transactions)
		return
	} else {
		blockchain.switchToChain(blockchain.Blocks[block.PrevHash])
		blockchain.addTransactionsToLedger(block.Transactions)
		return
	}
}

func (blockchain *Blockchain) addTransactionsToLedger(signedTransaction []account.SignedTransaction) {
	for _, signedTransaction := range signedTransaction {
		blockchain.Ledger.ProcessSignedTransaction(&signedTransaction)
	}
}

// add signed transaction to a block
func (block *Block) AddTransaction(st *account.SignedTransaction) {
	block.Transactions = append(block.Transactions, *st)
}

// hashes a block
func (block *Block) HashBlock(key rsa.PublicKey) string {
	blockData := rsa.EncodePublicKey(key) + block.PrevHash + fmt.Sprintf("%d", (block.timeStamp.Second()/SLOTLENGTH)+1) + account.EncodeTransactions(block.Transactions) + block.timeStamp.String()
	hash := sha256.New()
	hash.Write([]byte(blockData))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// signs a block
func (block *Block) SignBlock(key account.Account) []byte {
	blockData := rsa.EncodePublicKey(key.Pk) + block.PrevHash + fmt.Sprintf("%d", (block.timeStamp.Second()/SLOTLENGTH)+1) + account.EncodeTransactions(block.Transactions) + block.timeStamp.String() + block.Hash
	signedMessage := rsa.SignMessage([]byte(blockData), key.Sk)
	return signedMessage
}

// verifies a block
func (block *Block) VerifyBlock(key rsa.PublicKey, signature []byte) bool {
	blockData := rsa.EncodePublicKey(key) + block.PrevHash + fmt.Sprintf("%d", (block.timeStamp.Second()/SLOTLENGTH)+1) + account.EncodeTransactions(block.Transactions) + block.timeStamp.String() + block.Hash
	verifiedSignature := rsa.VerifySignature([]byte(blockData), signature, key)
	validBlockTransactions := block.VerifyBlockTransactions()
	return verifiedSignature && validBlockTransactions
}

// verify all transactions in a block
func (block *Block) VerifyBlockTransactions() bool {
	for _, signedTransaction := range block.Transactions {
		pk := rsa.DecodePublicKey(signedTransaction.From)
		if !account.VerifySignedTransaction(pk, &signedTransaction) {
			return false
		}
	}
	return true
}

// processes a block taken as a parameter
func (blockchain *Blockchain) ProcessLotteryBlock(lottery Lottery) {
	verified := lottery.Block.VerifyBlock(lottery.Pk, lottery.signature)
	if verified {
		blockchain.AddBlock(*lottery.Block)
		fmt.Println("Block verified")
	} else {
		fmt.Println("Block not verified")
	}
}
