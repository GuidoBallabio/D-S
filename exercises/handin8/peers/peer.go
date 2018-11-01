package peers

import (
	"fmt"
	"net"
	"strconv"
	"encoding/gob"
)

// Peer is an object representing peers connections
type Peer struct {
	IP     string
	Port   int
	PubKey string
	conn   net.Conn
	enc	   *gob.Encoder
}

// newPeer is the constructor of the Peer type
func newPeer(IP string, Port int, conn net.Conn) Peer {
	return Peer{
		IP:   IP,
		Port: Port,
		conn: conn}
}

// AddConn sets a conn to an existing peer
func (peer *Peer) AddConn(conn net.Conn) {
	peer.conn = conn
}

// AddPubKey sets a conn to an existing peer
func (peer *Peer) AddPubKey(key string) {
	peer.PubKey = key
}

// GetAddress return the address of the peer as IP:Port
func (peer *Peer) GetAddress() string {
	return peer.IP + ":" + peer.GetPort()
}

// GetPort return the address of the peer as IP:Port
func (peer *Peer) GetPort() string {
	return strconv.Itoa(peer.Port)
}

// GetConn return the connection to the peer if available
func (peer *Peer) GetConn() net.Conn {
	return peer.conn
}

// Less defines an order relationshIP for peers
func (peer *Peer) less(peer2 Peer) bool {
	switch {
	case peer.IP < peer2.IP:
		return true
	case peer.IP == peer2.IP:
		if peer.Port < peer2.Port {
			return true
		}
	}
	return false
}

func (peer Peer) String() string {
	return fmt.Sprintf("Peer: %s", peer.GetAddress())
}
