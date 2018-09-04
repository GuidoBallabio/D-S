package main

import (
	"bufio"
	"fmt"
	"net"

	"./lib"
)

var connArray = make(map[net.Conn]bool)
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

func askPeer() (net.IP, int) {
	var temp string
	var ip net.IP
	var port int

	fmt.Println("Enter IP address:")
	fmt.Scanln(&temp)

	ip = net.ParseIP(temp)

	fmt.Println("Enter port:")
	fmt.Scanf("%d", &port)

	return ip, port
}

func connect(ip net.IP, port int) (net.Conn, error) {
	return net.Dial("tcp", ip.String()+":"+string(port))
}

func connectToNetwork(ip net.IP, port int, listenCh chan string, quitCh chan struct{}) {
	conn1, err := connect(ip, port)
	if err == nil {
		fmt.Println("Connection to the network Succesfull")
	} else {
		fmt.Println(err.Error())
		fmt.Printf("Initializing your own network on port %d\n", port)
	}

	go handleConn(conn1, listenCh)

	ln, _ := net.Listen("tcp", ":"+string(port))
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		if _, done := <-quitCh; !done {
			break //Done
		}
		go handleConn(conn, listenCh)
	}

}

func broadcast(msg string) {
	messages.Set(msg, true)
	for conn := range connArray {
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
