package peers

import (
	"encoding/gob"
	"fmt"
	"net"
	"strconv"
)

// Peer is an object representing peers connections
type Peer struct {
	IP     string
	Port   int
	PubKey string
	conn   net.Conn
	enc    *gob.Encoder
	dec    *gob.Decoder
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

// AddEnc sets an encoder to an existing peer
func (peer *Peer) AddEnc(enc *gob.Encoder) {
	peer.enc = enc
}

// AddDec sets an decoder to an existing peer
func (peer *Peer) AddDec(dec *gob.Decoder) {
	peer.dec = dec
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

// GetDec return the decoder to the peer if available
func (peer *Peer) GetDec() *gob.Decoder {
	return peer.dec
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
