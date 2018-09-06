package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"

	"./lib"
)

const defaultPort string = "4444"

var listeningPort string

var connArray = lib.NewAtomicSlice()
var messages = lib.NewAtomicMap()

func main() {
	ip, port := askPeer()
	var writeCh = make(chan string)
	var listenCh = make(chan string)
	var quitCh = make(chan struct{})

	go connectToNetwork(ip, port, listenCh, quitCh)
	go printBroadcast(writeCh, listenCh, quitCh)
	go write(writeCh, quitCh)
	<-quitCh
}

func askPeer() (ip net.IP, port string) {
	var temp string

	fmt.Println("Enter IP address:")
	fmt.Scanln(&temp)

	ip = net.ParseIP(temp)

	fmt.Println("Enter port:")
	fmt.Scanln(&port)

	return ip, port
}

func connect(ip net.IP, port string) (net.Conn, error) {
	if ip == nil {
		return nil, errors.New("IP is not valid")
	}
	return net.Dial("tcp", ip.String()+":"+port)
}

func getLocalIP() net.IP {
	netInterfaceAddresses, err := net.InterfaceAddrs()

	if err != nil {
		return nil
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {

		networkIP, ok := netInterfaceAddress.(*net.IPNet)

		if ok && !networkIP.IP.IsLoopback() && networkIP.IP.To4() != nil {
			return networkIP.IP
		}
	}
	return nil
}

func connectToNetwork(ip net.IP, port string, listenCh chan string, quitCh chan struct{}) {
	conn1, err := connect(ip, port)
	if err == nil {
		fmt.Println("Connection to the network Succesfull")
		connArray.Append(conn1)
		go handleConn(conn1, listenCh)
		portInt, _ := strconv.Atoi(port)
		port = strconv.Itoa(portInt + 1)
	} else {
		fmt.Println(err.Error())
		port = defaultPort
		fmt.Printf("Initializing your own network on port %s\n", port)
	}
	listeningPort = port
	fmt.Println("Your IP is:", getLocalIP(), "with open port:", port)
	fmt.Println("You can start chatting")

	ln, err := net.Listen("tcp", ":"+port)
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
				break //Done
			}
		default:
			connArray.Append(conn)
			go handleConn(conn, listenCh)
		}
	}

}

func broadcast(msg string) {
	messages.Set(msg, true)

	for conn := range connArray.Iter() {
		fmt.Fprintf(conn, msg)
	}
}

func printBroadcast(writeCh chan string, listenCh chan string, quitCh chan struct{}) {
	for {
		select {
		case msg := <-writeCh:
			broadcast(msg)
		case msg := <-listenCh:
			if val, found := messages.Get(msg); !found || !val {
				fmt.Printf(msg)
				broadcast(msg)
			}
		case <-quitCh:
			connect(getLocalIP(), listeningPort)
			closeAllConn()
			break //Done
		}
	}
}

func closeAllConn() {
	for conn := range connArray.Iter() {
		conn.Close()
	}
}

func write(writeCh chan string, quitCh chan struct{}) {
	var msg string
	for {
		fmt.Scanln(&msg)
		if msg == "quit" {
			close(quitCh)
			break //Done
		}
		writeCh <- (msg + "\n")
	}
}

func handleConn(conn net.Conn, listenCh chan string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Closed connection to", conn.RemoteAddr())
			break //Done
		} else {
			listenCh <- msg
		}
	}
}
