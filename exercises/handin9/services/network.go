package services

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net"

	. "../account"
	"../aesrsa"
	bt "../blocktree"
	. "../peers"
)

// LocalPeer is the ID of the local machine
var LocalPeer Peer

// PeerList os the list of peers known
var PeerList = NewList()

// InitNetwork preconfigures some basic properties of the network layer
func InitNetwork() {
	gob.Register(&SignedNode{})
	gob.Register(&SignedTransaction{})
}

// ConnectToNetwork connects the local machine to a pre-existing network
func ConnectToNetwork(peer Peer, listenCh chan<- SignedTransaction, blockCh chan<- bt.SignedNode, localPK aesrsa.RSAKey) {
	conn1, err := Connect(&peer)

	if err != nil {
		panic(err.Error())
	}

	LocalPeer = GetLocalPeer(peer.Port+1, aesrsa.KeyToString(localPK))
	fmt.Println("Connection to the network Succesfull")
	PeerList.SortedInsert(&LocalPeer)
	handleFirstConn(conn1, listenCh, blockCh)
	fmt.Println("Your IP is:", LocalPeer.IP, "with open port:", LocalPeer.GetPort())
}

// CreateNetwork let the local machine create a p2p network
func CreateNetwork(port int, listenCh chan<- SignedTransaction, blockCh chan<- bt.SignedNode, localPK aesrsa.RSAKey) {
	LocalPeer = GetLocalPeer(port, aesrsa.KeyToString(localPK))
	PeerList.SortedInsert(&LocalPeer)
	fmt.Println("Initializing your own network")
	fmt.Println("Your IP is:", LocalPeer.IP, "with open port:", LocalPeer.GetPort())
}

// Connect starts a tcp connection given a peer
func Connect(peer *Peer) (net.Conn, error) {
	if peer.IP == "<nil>" {
		return nil, errors.New("IP is not valid")
	}
	return net.Dial("tcp", peer.GetAddress())
}

func handleFirstConn(conn net.Conn, listenCh chan<- SignedTransaction, blockCh chan<- bt.SignedNode) {
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	// asking for list of peers
	signalAsk(enc)

	p := &Peer{}
	err := dec.Decode(p)
	for p.Port != -1 {
		if err == nil {
			PeerList.SortedInsert(p)
		}
		p = &Peer{}
		err = dec.Decode(p)
	}
	conn.Close()

	// broadcasting ourselves
	i := 0
	for p := range PeerList.IterWrap(&LocalPeer) {
		if *p != LocalPeer {
			if i >= 10 {
				break
			}
			conn, err := Connect(p)
			if err == nil {
				p.AddConn(conn)
				enc = p.GetEnc()
				signalNoAsk(enc)
				Wg.Add(1)
				go handleConn(p, listenCh, blockCh)
			}
			i++
		}
	}
}

// ask for list of peers
func signalAsk(enc *gob.Encoder) {
	enc.Encode(Peer{IP: "", Port: -1})
}

// signal not asking for list of peers
func signalNoAsk(enc *gob.Encoder) {
	enc.Encode(LocalPeer)
}

// BeServer let the local machine accept connections to the p2p network
func BeServer(listenCh chan<- SignedTransaction, blockCh chan<- bt.SignedNode, quitCh <-chan struct{}) {
	defer fmt.Println("server closed")
	defer Wg.Done()

	ln, err := net.Listen("tcp", ":"+LocalPeer.GetPort())
	if err != nil {
		fmt.Println("Fatal server error")
		panic(-1)
	}
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		select {
		case _, open := <-quitCh:
			if !open {
				ln.Close()
				closeAllConn()
				return //Done
			}
		default:
			if p, firstConn := checkAsk(conn, sequencer); !firstConn {
				Wg.Add(1)
				go handleConn(p, listenCh, blockCh)
			}
		}
	}

}

func closeAllConn() {
	for conn := range PeerList.IterConn() {
		conn.Close()
	}
}

// check if the peer asks for list of peers
func checkAsk(conn net.Conn, sequencer aesrsa.RSAKey) (*Peer, bool) {
	dec := gob.NewDecoder(conn)
	p := &Peer{}
	err := dec.Decode(p)
	if err == nil {
		if p.Port == -1 {
			enc := gob.NewEncoder(conn)

			for p := range PeerList.Iter() {
				enc.Encode(*p)
			}

			enc.Encode(Peer{Port: -1})
			return &Peer{}, true
		}
		p.AddConn(conn)
		p.AddDec(dec)
		PeerList.SortedInsert(p)
		return p, false
	}
	return &Peer{}, true
}

func handleConn(peer *Peer, listenCh chan<- SignedTransaction, blockCh chan<- bt.SignedNode) {
	defer Wg.Done()
	defer peer.Close()

	fmt.Println("Connected to", peer)

	dec := peer.GetDec()

	for {
		var obj WhatType
		err := dec.Decode(&obj)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Closed connection to", peer)
			PeerList.Remove(peer)
			break //Done
		} else {
			switch obj.WhatType() {
			case "SignedTransaction":
				listenCh <- *obj.(*SignedTransaction)
			case "SignedNode":
				blockCh <- *obj.(*bt.SignedNode)
			}
		}
	}
}
