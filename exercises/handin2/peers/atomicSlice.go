package peers

import (
	"errors"
	"net"
	"sort"
	"sync"
)

// AtomicSortedSlice is a synchronized []Peer type
type AtomicSortedSlice struct {
	data   []Peer
	rwLock sync.RWMutex
}

// NewList is the constructor of the AtomicSortedSlice type, it just creates an empty synchronized slice
func NewList() *AtomicSortedSlice {
	var aslice AtomicSortedSlice
	aslice.data = make([]Peer, 10)
	return &aslice
}

// AddPeerFromConn correctly adds a peer to the data structure from a net.Conn
func (aslice *AtomicSortedSlice) AddPeerFromConn(conn net.Conn) Peer {
	addr := conn.RemoteAddr().(*net.TCPAddr)
	p := newPeer(addr.IP.String(), addr.Port, conn)
	aslice.SortedInsert(p)
	return p
}

// find search the slice for an element and return its index
func (aslice *AtomicSortedSlice) find(peer Peer) (int, error) {
	i := sort.Search(len(aslice.data), func(i int) bool { return aslice.data[i].less(peer) })
	if i < len(aslice.data) && aslice.data[i] == peer {
		return i, nil
	}

	return i, errors.New("Not found, but i is the index where it would be inserted")
}

// SortedInsert is the synchronized sorted append that returns the index where it was added
func (aslice *AtomicSortedSlice) SortedInsert(peer Peer) int {
	aslice.rwLock.Lock()
	defer aslice.rwLock.Unlock()

	l := len(aslice.data)
	if l == 0 {
		aslice.data[0] = peer
		return 0
	}

	i, err := aslice.find(peer)

	if err != nil { //	i==l not found = new value is the smallest
		aslice.data = append([]Peer{peer}, aslice.data...)
		return 0
	}

	if i == l-1 { // new value is the biggest
		aslice.data = append(aslice.data[0:i], peer)
		return i
	}

	aslice.data = append(aslice.data[0:i], peer)
	aslice.data = append(aslice.data, aslice.data[i+1:]...)
	return i
}

// Iter iterates over the items in the concurrent slice
// Each item is sent over a channel, so that
// we can iterate over the slice using the builin range keyword
func (aslice *AtomicSortedSlice) Iter() <-chan Peer {
	c := make(chan Peer)

	f := func() {
		aslice.rwLock.RLock()
		defer aslice.rwLock.RUnlock()
		for _, value := range aslice.data {
			c <- value
		}
		close(c)
	}
	go f()

	return c
}

// IterWrap accept a peer or an index and iterates from that item to itself wrapping around
func (aslice *AtomicSortedSlice) IterWrap(peer Peer) <-chan *Peer {

	c := make(chan *Peer)

	f := func() {
		aslice.rwLock.RLock()
		defer aslice.rwLock.RUnlock()

		i, err := aslice.find(peer)
		if err != nil {
			close(c)
			return
		}

		for _, value := range aslice.data[i+1:] {
			c <- &value
		}
		for _, value := range aslice.data[:i] {
			c <- &value
		}
		close(c)
	}
	go f()

	return c
}

// IterConn iterates over the available connections of the peers
func (aslice *AtomicSortedSlice) IterConn() <-chan net.Conn {
	c := make(chan net.Conn)

	f := func() {
		aslice.rwLock.RLock()
		defer aslice.rwLock.RUnlock()
		for _, value := range aslice.data {
			if value.conn != nil {
				c <- value.conn
			}
		}
		close(c)
	}
	go f()

	return c
}

// GetPeerByConn returns a peer given a net.Conn
func (aslice *AtomicSortedSlice) GetPeerByConn(conn net.Conn) Peer {
	aslice.rwLock.RLock()
	defer aslice.rwLock.RUnlock()

	for _, value := range aslice.data {
		if value.conn == conn {
			return value
		}
	}
	return Peer{}
}
