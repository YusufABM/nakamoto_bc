package peer

import (
	"HAND_IN_2/account"
	"net"
)

// Peer is a struct that contains the IP address of the peer and the ledger
type Peer struct {
	Port   int
	Ledger *account.Ledger
	conn   net.Conn
}

// NewPeer creates a new instance of Peer
func NewPeer(port int, ledger *account.Ledger) *Peer {
	peer := new(Peer)
	peer.Port = port
	peer.Ledger = ledger
	return peer
}
