package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	. "./account"
	. "./peers"
)

const defaultPort int = 4444

var localPeer Peer

var peersList = NewList()
var ledger = NewLedger()
var wg sync.WaitGroup

func main() {
	firstPeer := askPeer()
	var kbCh = make(chan Transaction)
	var listenCh = make(chan Transaction)
	var quitCh = make(chan struct{})

	connectToNetwork(firstPeer, listenCh)

	wg.Add(3)

	go beServer(listenCh, quitCh)
	go printBroadcast(kbCh, listenCh, quitCh)
	go write(kbCh, quitCh)
	<-quitCh
	wg.Wait()
}

func askPeer() Peer {
	var temp string

	fmt.Println("Enter IP address:")
	fmt.Scanln(&temp)

	ip := net.ParseIP(temp)

	fmt.Println("Enter port:")
	fmt.Scanln(&temp)
	port, _ := strconv.Atoi(temp)

	return Peer{
		IP:   ip.String(),
		Port: port}
}

func connect(peer Peer) (net.Conn, error) {
	if peer.IP == "<nil>" {
		return nil, errors.New("IP is not valid")
	}
	return net.Dial("tcp", peer.GetAddress())
}

func connectToNetwork(peer Peer, listenCh chan<- Transaction) {
	conn1, err := connect(peer)
	if err == nil {
		fmt.Println("Connection to the network Succesfull")
		localPeer = GetLocalPeer(peer.Port + 1)
		peersList.SortedInsert(localPeer)
		handleFirstConn(conn1, listenCh) //remember to add to the list of peers
	} else {
		fmt.Println(err.Error())
		localPeer = GetLocalPeer(defaultPort)
		peersList.SortedInsert(localPeer)
		fmt.Println("Initializing your own network")
	}

	fmt.Println("Your IP is:", peer.IP, "with open port:", peer.GetPort())
	fmt.Println("You can start chatting")

}

func beServer(listenCh chan<- Transaction, quitCh <-chan struct{}) {
	defer wg.Done()

	ln, err := net.Listen("tcp", ":"+localPeer.GetPort())
	if err != nil {
		fmt.Println("Fatal server error")
		panic(-1)
	}
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		select {
		case _, done := <-quitCh:
			if !done {
				closeAllConn()
				break //Done
			}
		default:
			checkAsk(conn)
			p := peersList.AddPeerFromConn(conn)
			wg.Add(1)
			go handleConn(p, listenCh)
		}
	}

}

func closeAllConn() {
	for conn := range peersList.IterConn() {
		conn.Close()
	}
}

func broadcast(t Transaction) {
	for conn := range peersList.IterConn() {
		enc := gob.NewEncoder(conn)
		enc.Encode(t)
	}
}

func printBroadcast(kbCh <-chan Transaction, listenCh <-chan Transaction, quitCh <-chan struct{}) {
	defer wg.Done()

	for {
		select {
		case t := <-kbCh:
			t = attachNextID(t)
			applyTransaction(t)
			broadcast(t)
		case t := <-listenCh:
			if checkIfAcceptable(t) {
				applyTransaction(t)
				fmt.Println("Received ", t)
				broadcast(t)
			}
		case <-quitCh:
			connect(localPeer)
			break //Done
		}
	}
}

func write(kbCh chan<- Transaction, quitCh chan<- struct{}) {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)

	for {
		t, quit := askTransaction()
		if quit {
			close(quitCh)
			break //Done
		}
		kbCh <- t
	}
}

func handleConn(peer Peer, listenCh chan<- Transaction) {
	defer wg.Done()
	defer peer.GetConn().Close()

	dec := gob.NewDecoder(peer.GetConn())

	for {
		t := Transaction{}
		err := dec.Decode(&t)

		if err != nil {
			fmt.Println("Closed connection to ", peer)
			break //Done
		} else {
			listenCh <- t
		}
	}
}

func handleFirstConn(conn net.Conn, listenCh chan<- Transaction) {
	defer conn.Close()

	signalAsk(conn)

	dec := gob.NewDecoder(conn)
	p := &Peer{}
	for p.Port == -1 {
		p := &Peer{}
		err := dec.Decode(p)
		if err != nil {
			peersList.SortedInsert(*p)
		}
	}

	i := 0
	for p := range peersList.IterWrap(localPeer) {
		if i >= 10 {
			break
		}
		conn1, err := connect(*p)
		if err != nil {
			p.AddConn(conn1)
		}
		signalNoAsk(conn1)
		i++
	}

	// useless broadcast of presence
	// for conn2 := range peersList.IterConn() {
	// 	enc := gob.NewEncoder(conn2)
	// 	enc.Encode(localPeer)
	// }

	wg.Add(1)
	peer := peersList.AddPeerFromConn(conn)
	go handleConn(peer, listenCh)
}

// ask for list of peers
func signalAsk(conn net.Conn) {
	enc := gob.NewEncoder(conn)
	enc.Encode(Peer{IP: "", Port: -2})
}

// signal not asking for list of peers
func signalNoAsk(conn net.Conn) {
	enc := gob.NewEncoder(conn)
	enc.Encode(Peer{IP: "", Port: -1})
}

// check if the peer asks for list of peers
func checkAsk(conn net.Conn) {
	dec := gob.NewDecoder(conn)
	p := &Peer{}
	err := dec.Decode(p)
	if err != nil {
		if p.Port == -2 {
			enc := gob.NewEncoder(conn)
			for p := range peersList.Iter() {
				enc.Encode(p)
			}
		}
	}
}
