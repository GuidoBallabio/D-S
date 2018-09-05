package lib

import (
	"net"
	"sync"
)

// AtomicSlice is a synchronized []net.Conn type
type AtomicSlice struct {
	data   []net.Conn
	rwLock sync.RWMutex
}

// NewAtomicSlice is the constructor of the AtomicSlice type, it just creates an empty synchronized slice
func NewAtomicSlice() *AtomicSlice {
	var aslice AtomicSlice
	aslice.data = make([]net.Conn, 0)
	return &aslice
}

// Append is the synchronized append
func (aslice *AtomicSlice) Append(conn net.Conn) {
	aslice.rwLock.Lock()
	defer aslice.rwLock.Unlock()
	aslice.data = append(aslice.data, conn)
}

// Iter iterates over the items in the concurrent slice
// Each item is sent over a channel, so that
// we can iterate over the slice using the builin range keyword
func (aslice *AtomicSlice) Iter() <-chan net.Conn {
	c := make(chan net.Conn)

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
