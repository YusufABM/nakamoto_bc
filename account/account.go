package account

import (
	"sync"
)

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

// Updates the ledger with the transaction
// Defer unlocks the mutex at the end of the function
func (l *Ledger) Transaction(t *Transaction) {
	if l.Accounts == nil {
		l.Accounts = make(map[string]int)
	}
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}
