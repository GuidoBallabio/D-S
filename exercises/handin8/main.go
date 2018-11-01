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
	"./aesrsa"
	. "./peers"
)

const defaultPort int = 4444

var localPeer Peer
var localKeys *aesrsa.RSAKeyPair

var sequencer aesrsa.RSAKey
var sequencerSecret aesrsa.RSAKey

var peersList = NewList()
var ledger = NewLedger()
var past = make(map[string]bool, 1)
var wg sync.WaitGroup

func main() {
	firstPeer := askPeer()
	var kbCh = make(chan SignedTransaction)
	var listenCh = make(chan SignedTransaction)
	var quitCh = make(chan struct{})

	createKeys()
	connectToNetwork(firstPeer, listenCh)

	wg.Add(3)
	go beServer(listenCh, quitCh)
	go processTransactions(kbCh, listenCh, quitCh)
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

func createKeys() {
	var err error

	localKeys, err = aesrsa.KeyGen(2048)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Your secret key is:")
	fmt.Println(aesrsa.KeyToString(localKeys.Private))
	fmt.Println("Your public key is:")
	fmt.Println(aesrsa.KeyToString(localKeys.Public))
}

func connectToNetwork(peer Peer, listenCh chan<- SignedTransaction) {
	conn1, err := connect(peer)
	if err == nil {
		fmt.Println("Connection to the network Succesfull")
		localPeer = GetLocalPeer(peer.Port+1, aesrsa.KeyToString(localKeys.Public))
		peersList.SortedInsert(localPeer)
		handleFirstConn(conn1, listenCh)
	} else {
		fmt.Println(err.Error())
		localPeer = GetLocalPeer(defaultPort, aesrsa.KeyToString(localKeys.Public))
		peersList.SortedInsert(localPeer)
		go beSequencer()
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

func handleFirstConn(conn net.Conn, listenCh chan<- SignedTransaction) {

	// asking for list of peers
	signalAsk(conn)

	getSequencer(conn)

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

// getSequencer receive the sequencer's public key
func getSequencer(conn net.Conn) {
	dec := gob.NewDecoder(conn)
	key := aesrsa.RSAKey{}
	err := dec.Decode(key)

	sequencer = key
}

func beServer(listenCh chan<- SignedTransaction, quitCh <-chan struct{}) {
	defer wg.Done()
	defer fmt.Println("server closed")

	ln, err := net.Listen("tcp", ":"+localPeer.GetPort())
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
				closeAllConn()
				return //Done
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

			enc.Encode(sequencer)

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

func handleConn(peer Peer, listenCh chan<- SignedTransaction) {
	fmt.Println("Connected to", peer)
	defer wg.Done()
	defer peer.GetConn().Close()

	for {
		dec := gob.NewDecoder(peer.GetConn())
		st := SignedTransaction{}
		err := dec.Decode(&st)

		if err != nil {
			fmt.Println(err)
			fmt.Println("Closed connection to", peer)
			peersList.Remove(peer)
			break //Done
		} else {
			listenCh <- st
		}
	}
}

func processTransactions(kbCh <-chan SignedTransaction, listenCh <-chan SignedTransaction, quitCh <-chan struct{}) {
	defer wg.Done()

	for {
		select {
		case st := <-kbCh:
			fmt.Println("Processing transaction")
			if t := st.ExtractTransaction(); !isOld(st) && isVerified(st) {
				updateLedger(t)
				fmt.Println("Sent:\n", t)
				fmt.Println(ledger)
				broadcast(st)
			}
		case st := <-listenCh:
			if t := st.ExtractTransaction(); !isOld(st) && isVerified(st) {
				updateLedger(t)
				fmt.Println("Received:\n", t)
				fmt.Println(ledger)
				broadcast(st)
			}
		case <-quitCh:
			connect(localPeer)
			return //Done
		}
	}
}

func isOld(st SignedTransaction) bool {
	if val, found := past[st.ID]; found && val {
		return true
	}
	return false
}

func isVerified(st SignedTransaction) bool {
	return st.VerifyTransaction() && st.Amount > 0
}

func signTransaction(t Transaction, k aesrsa.RSAKey) SignedTransaction {
	return SignTransaction(t, k)
}

func attachNextID(t Transaction) Transaction {
	t.ID = fmt.Sprintf("%d-%s", ledger.GetClock(), localPeer.GetAddress())
	return t
}

func updateLedger(t Transaction) {
	ledger.Transaction(t)
}

func broadcast(st SignedTransaction) {
	past[st.ID] = true

	for conn := range peersList.IterConn() {
		enc := gob.NewEncoder(conn)
		enc.Encode(st)
	}
}

func write(kbCh chan<- SignedTransaction, quitCh chan<- struct{}) {
	defer wg.Done()

	fmt.Println("Insert a transaction as: FromWhom ToWhom HowMuch each on different lines, then the private key to sign it ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	for {
		t, quit := askTransaction(scanner)
		if quit {
			fmt.Println("quitting...")
			close(quitCh)
			break //Done
		}
		t = attachNextID(t)
		key := aesrsa.KeyFromString(scanKey(scanner))
		st := signTransaction(t, key)
		kbCh <- st
	}
}

func askTransaction(scanner *bufio.Scanner) (Transaction, bool) {

	from := scanKey(scanner)

	if from == "quit" {
		return Transaction{}, true
	}

	to := scanKey(scanner)

	if to == "quit" {
		return Transaction{}, true
	}

	scanner.Scan()
	amount := scanner.Text()

	if amount == "quit" {
		return Transaction{}, true
	}

	intAmount, err := strconv.Atoi(amount)

	for err != nil {
		fmt.Println("not valid integer amount")
		scanner.Scan()
		amount := scanner.Text()

		if amount == "quit" {
			return Transaction{}, true
		}

		intAmount, err = strconv.Atoi(amount)
	}

	return Transaction{
		From:   from,
		To:     to,
		Amount: intAmount}, false
}

func scanKey(scanner *bufio.Scanner) string {
	scanner.Scan()
	buf := scanner.Text()

	for buf != "-----BEGIN KEY-----" {
		scanner.Scan()
		buf = scanner.Text()
	}

	key := buf + "\n"

	scanner.Scan()
	buf = scanner.Text()

	for buf != "-----END KEY-----" {
		key += buf

		scanner.Scan()
		buf = scanner.Text()
	}

	key += "\n" + buf

	return key
}