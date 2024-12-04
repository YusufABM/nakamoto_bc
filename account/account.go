package account

import (
	"HAND_IN_2/rsa"
	"encoding/base64"
	"fmt"
	"sync"
)

type Account struct {
	Pk rsa.PublicKey
	Sk rsa.SecretKey
}

// From assignment description
type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

// Ledger is a map of account names to balances
// sync.Mutex = It is safe for concurrent use
type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
}

// Creates a new instance of Ledger with an empty map
// Returns a pointer to the new Ledger
func MakeLedger() *Ledger {
	Ledger := new(Ledger)
	Ledger.Accounts = make(map[string]int)
	return Ledger
}

// creates 10 accounts with 1000000 in each
func CreateGenesisBlocks() {
	for i := 0; i < 10; i++ {
		// Create genesis block
		// Create a new ledger
		ledger := MakeLedger()
		// Create a new account
		account := MakeAccount()
		// Add the account to the ledger
		ledger.Accounts[rsa.EncodePublicKey(account.Pk)] = 1000000
	}
}

func MakeAccount() *Account {
	Account := new(Account)
	pk, sk, err := rsa.Keygen(1024)
	if err != nil {
		fmt.Println("Error generating keys")
	}
	Account.Pk = pk
	Account.Sk = sk
	return Account
}

type SignedTransaction struct {
	ID        string // Any string
	From      string // A verifaction key coded as a string
	To        string // A verifaction key coded as a string
	Amount    int    // Amount to transfer
	Signature string // Potential signature coded as string
}

func (l *Ledger) ProcessSignedTransaction(st *SignedTransaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	/* We verify that the t.Signature is a valid RSA
	 * signature on the rest of the fields in t under
	 * the public key t.From
	 */

	pk := rsa.DecodePublicKey(st.From)
	//fmt.Println("decoded pk: ", pk)
	validSignature := VerifySignedTransaction(pk, st)
	if validSignature {
		l.Accounts[st.From] -= st.Amount
		l.Accounts[st.To] += st.Amount - 1
	} else {
		fmt.Println("Invalid signature")
		//fmt.Println(st.Signature)
	}
}

// Updates the ledger with the transaction
// Defer unlocks the mutex at the end of the function
func (l *Ledger) ProcessTransaction(t *Transaction) {
	if l.Accounts == nil {
		l.Accounts = make(map[string]int)
	}
	l.lock.Lock()
	defer l.lock.Unlock()

	//transfers money
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

func SignTransaction(sk rsa.SecretKey, t *Transaction) SignedTransaction {
	message := t.ID + t.From + t.To + string(t.Amount)
	signature := rsa.SignMessage([]byte(message), sk)
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	return SignedTransaction{ID: t.ID, From: t.From, To: t.To, Amount: t.Amount, Signature: encodedSignature}
}

func VerifySignedTransaction(pk rsa.PublicKey, st *SignedTransaction) bool {
	message := st.ID + st.From + st.To + string(st.Amount)
	if st.Amount%1 != 0 {
		fmt.Println("Amount is not an integer")
		return false
	}
	decodedSignature, err := base64.StdEncoding.DecodeString(st.Signature)
	if err != nil {
		fmt.Println("Error decoding signature:", err)
		return false
	}
	return rsa.VerifySignature([]byte(message), decodedSignature, pk)
}

// encode signed transaction to string
func EncodeSignedTransaction(st SignedTransaction) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(st.ID + st.From + st.To + string(st.Amount) + st.Signature))
	return encoded
}

// encodes a list of transactions to a string
func EncodeTransactions(transactions []SignedTransaction) string {
	encoded := ""
	for _, t := range transactions {
		encoded += EncodeSignedTransaction(t)
	}
	return encoded
}
