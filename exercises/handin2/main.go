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
	go processTransaction(kbCh, listenCh, quitCh)
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

	fmt.Println("Your IP is:", localPeer.IP, "with open port:", localPeer.GetPort())
}

func connect(peer Peer) (net.Conn, error) {
	if peer.IP == "<nil>" {
		return nil, errors.New("IP is not valid")
	}
	return net.Dial("tcp", peer.GetAddress())
}

func handleFirstConn(conn net.Conn, listenCh chan<- Transaction) {
	defer conn.Close()

	// asking for list of peers
	signalAsk(conn)
	dec := gob.NewDecoder(conn)
	p := Peer{}
	err := dec.Decode(p)
	for p.Port != -1 {
		if err == nil {
			peersList.SortedInsert(p)
		}
		err = dec.Decode(&p)
	}
	conn.Close()

	// broadcasting ourselves
	i := 0
	for p := range peersList.IterWrap(localPeer) {
		if p != localPeer {
			if i >= 10 {
				break
			}
			conn1, err := connect(p)
			if err == nil {
				peersList.AddConn(p, conn1)
				p.AddConn(conn1)
				signalNoAsk(conn1)
				wg.Add(1)
				go handleConn(p, listenCh)
			}
			i++
		}
	}

}

// ask for list of peers
func signalAsk(conn net.Conn) {
	enc := gob.NewEncoder(conn)
	enc.Encode(Peer{IP: "", Port: -1})
}

// signal not asking for list of peers
func signalNoAsk(conn net.Conn) {
	enc := gob.NewEncoder(conn)
	enc.Encode(localPeer)
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
			if p, firstConn := checkAsk(conn); !firstConn {
				wg.Add(1)
				go handleConn(p, listenCh)
			}
		}
	}

}

func closeAllConn() {
	for conn := range peersList.IterConn() {
		conn.Close()
	}
}

// check if the peer asks for list of peers
func checkAsk(conn net.Conn) (Peer, bool) {
	dec := gob.NewDecoder(conn)
	p := &Peer{}
	err := dec.Decode(p)
	if err == nil {
		if p.Port == -1 {
			enc := gob.NewEncoder(conn)
			for p := range peersList.Iter() {
				enc.Encode(p)
			}

			enc.Encode(Peer{Port: -1})
			return Peer{}, true
		}
		p.AddConn(conn)
		peersList.SortedInsert(*p)
		return *p, false
	}
	return Peer{}, true
}

func handleConn(peer Peer, listenCh chan<- Transaction) {
	fmt.Println("Connected to", peer)
	defer wg.Done()
	defer peer.GetConn().Close()

	dec := gob.NewDecoder(peer.GetConn())

	for {
		t := Transaction{}
		err := dec.Decode(&t)

		if err != nil {
			fmt.Println(err)
			fmt.Println("Closed connection to", peer)
			break //Done
		} else {
			listenCh <- t
		}
	}
}

func processTransaction(kbCh <-chan Transaction, listenCh <-chan Transaction, quitCh <-chan struct{}) {
	defer wg.Done()

	for {
		select {
		case t := <-kbCh:
			fmt.Println("processing") //remove
			t = attachNextID(t)
			applyTransaction(t)
			broadcast(t)
		case t := <-listenCh:
			if checkIfAcceptable(t) {
				applyTransaction(t)
				fmt.Println("Received", t)
				broadcast(t)
			}
		case <-quitCh:
			connect(localPeer)
			break //Done
		}
	}
}

func checkIfAcceptable(t Transaction) bool {
	return true
}

func attachNextID(t Transaction) Transaction {
	return t
}

func applyTransaction(t Transaction) {

}

func broadcast(t Transaction) {
	fmt.Println(peersList)
	for conn := range peersList.IterConn() {
		fmt.Println(conn) //remove
		enc := gob.NewEncoder(conn)
		enc.Encode(t)
	}
}

func write(kbCh chan<- Transaction, quitCh chan<- struct{}) {
	defer wg.Done()

	fmt.Println("Insert a transaction as FromWho ToWhom HowMuch")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)

	for {
		t, quit := askTransaction(scanner)
		if quit {
			fmt.Println("quitting")
			close(quitCh)
			break //Done
		}
		kbCh <- t
		fmt.Println("Sent ", t)
	}
}

func askTransaction(scanner *bufio.Scanner) (Transaction, bool) {

	scanner.Scan()
	from := scanner.Text()

	if from == "quit" {
		return Transaction{}, true
	}

	scanner.Scan()
	to := scanner.Text()

	if from == "quit" {
		return Transaction{}, true
	}

	scanner.Scan()
	amount := scanner.Text()

	if from == "quit" {
		return Transaction{}, true
	}

	intAmount, err := strconv.Atoi(amount)
	for err != nil {
		fmt.Println("not valid integer amount")
		scanner.Scan()
		amount := scanner.Text()

		if from == "quit" {
			return Transaction{}, true
		}

		intAmount, err = strconv.Atoi(amount)
	}

	return Transaction{
		From:   from,
		To:     to,
		Amount: intAmount}, false
}
