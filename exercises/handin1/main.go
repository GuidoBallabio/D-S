package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"

	"./lib"
)

const DEFAULT_PORT int = 4444

var connArray = lib.NewAtomicSlice()
var messages = lib.NewAtomicMap()

func main() {
	ip, port := askPeer()
	var writeCh = make(chan string)
	var listenCh = make(chan string)
	var quitCh = make(chan struct{})

	go connectToNetwork(ip, port, listenCh, quitCh)
	go print(writeCh, listenCh, quitCh)
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

	if false { //remove
		fmt.Printf("Not valid port, using default port %d\n", DEFAULT_PORT)
		//port = DEFAULT_PORT
	}

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
	} else {
		fmt.Println(err.Error())
		fmt.Printf("Initializing your own network on port %s\n", port)
	}
	fmt.Println("Your IP is: ", getLocalIP())

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Fatal server error")
		return
	}
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		if _, done := <-quitCh; !done {
			break //Done
		}
		connArray.Append(conn)
		go handleConn(conn, listenCh)
	}

}

func broadcast(msg string) {
	messages.Set(msg, true)
	for conn := range connArray.Iter() {
		fmt.Fprintf(conn, msg)
	}
}

func print(writeCh chan string, listenCh chan string, quitCh chan struct{}) {
	for {
		select {
		case msg := <-writeCh:
			broadcast(msg)
		case msg := <-listenCh:
			if val, found := messages.Get(msg); !found || !val {
				fmt.Println(msg)
				broadcast(msg)
			}
		case <-quitCh:
			break //Done
		}
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
		writeCh <- msg
	}
}

func handleConn(conn net.Conn, listenCh chan string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: " + err.Error())
			break //Done
		} else {
			listenCh <- msg
		}
	}
}
