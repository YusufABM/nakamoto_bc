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
const HARDNESS int = 913000000000
const debugging bool = true
const BLOCKSIZE int = 40

// Block is a struct that contains the previous hash, the hash, the nonce, the transactions and the timestamp
type Block struct {
	PrevHash     string
	Hash         string
	Transactions []account.SignedTransaction
	Height       int
}

// Blockchain is a struct that contains the blocks and the ledger
type Blockchain struct {
	Blocks        map[string]Block
	Ledger        *account.Ledger
	GenesisLedger *account.Ledger
	Seed          int
	ChainHead     Block
	StartTime     time.Time
	BlockSize     int
}

type Lottery struct {
	Block     *Block
	Slot      []byte
	Pk        rsa.PublicKey
	Draw      []byte
	Signature []byte
}

// NewBlockchain creates a new instance of Blockchain with a genesis block where 10 accounts are created with 1000000 in each
func NewBlockchain(ledger *account.Ledger, startTime time.Time) *Blockchain {
	blockchain := new(Blockchain)
	blockchain.Ledger = ledger.DeepCopy()
	blockchain.GenesisLedger = ledger.DeepCopy()
	blockchain.Seed = 42
	blockchain.StartTime = startTime
	blockchain.Blocks = make(map[string]Block)
	blockchain.BlockSize = BLOCKSIZE
	return blockchain
}

func NewLotteryBlock(block Block, pk rsa.PublicKey, sk rsa.SecretKey, slotNum []byte, draw []byte) *Lottery {
	lottery := new(Lottery)
	lottery.Block = &block
	lottery.Pk = pk
	//sets lottery.slot to the entire slotNum array
	lottery.Slot = slotNum
	lottery.Signature = block.SignBlock(pk, sk)
	lottery.Draw = draw
	return lottery
}

func NewBlock(parent *Block, transactions []account.SignedTransaction, pk rsa.PublicKey) *Block {
	block := new(Block)
	parentHash := ""
	block.Transactions = transactions
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
		blockchain.Ledger = blockchain.GenesisLedger.DeepCopy()
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

func (blockchain *Blockchain) addTransactionsToLedger(signedTransactions []account.SignedTransaction) {
	for _, signedTransaction := range signedTransactions {
		blockchain.Ledger.ProcessSignedTransaction(&signedTransaction)
	}
}

// add signed transaction to a block
func (block *Block) AddTransaction(st *account.SignedTransaction) {
	block.Transactions = append(block.Transactions, *st)
}

// hashes a block
func (block *Block) HashBlock(key rsa.PublicKey) string {
	blockData := rsa.EncodePublicKey(key) + block.PrevHash + account.EncodeTransactions(block.Transactions)
	hash := sha256.New()
	hash.Write([]byte(blockData))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// signs a block
func (block *Block) SignBlock(pk rsa.PublicKey, sk rsa.SecretKey) []byte {
	blockData := rsa.EncodePublicKey(pk) + block.PrevHash + account.EncodeTransactions(block.Transactions) + block.Hash
	signedMessage := rsa.SignMessage([]byte(blockData), sk)
	return signedMessage
}

// verifies a block
func (block *Block) VerifyBlock(key rsa.PublicKey, signature []byte) bool {
	blockData := rsa.EncodePublicKey(key) + block.PrevHash + account.EncodeTransactions(block.Transactions) + block.Hash
	verifiedSignature := rsa.VerifySignature([]byte(blockData), signature, key)
	validBlockTransactions := block.VerifyBlockTransactions()

	if debugging {
		if !verifiedSignature {
			fmt.Println("Signature not verified")
		} else if !validBlockTransactions {
			fmt.Println("Block transactions not verified")
		} else {
			fmt.Println("Block verified")
		}
	}

	return verifiedSignature && validBlockTransactions
}

// verify all transactions in a block
func (block *Block) VerifyBlockTransactions() bool {
	for _, signedTransaction := range block.Transactions {
		pk := rsa.DecodePublicKey(signedTransaction.From)
		isStValid := account.VerifySignedTransaction(pk, &signedTransaction)
		if !isStValid {
			fmt.Println("reveived a transaction with an invalid signature")
			return false

		}
	}
	return true
}

// processes a block taken as a parameter
func (blockchain *Blockchain) ProcessLotteryBlock(lottery Lottery) {
	verified := lottery.Block.VerifyBlock(lottery.Pk, lottery.Signature)
	if verified {
		blockchain.AddBlock(*lottery.Block)
	}
}

// add 10 AU plus 1 AU per transaction to the miner
func (blockchain *Blockchain) AddMinerReward(Lottery Lottery) {
	miner := Lottery.Pk
	transactions := len(Lottery.Block.Transactions)
	amount := transactions + 10
	blockchain.Ledger.Accounts[rsa.EncodePublicKey(miner)] += amount
}
