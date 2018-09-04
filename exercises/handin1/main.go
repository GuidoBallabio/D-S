package main

import (
	"fmt"
	"net"
)

var connArray = []net.Conn
var messages = map[string]bool

func main() {
	ip, port := askPeer()
	var writeCh = make(chan string)
	var listenCh = make(chan string)
	go connectToNetwork(ip, port, listenCh)
	go print(writech, listenCh)
	go write(writeCh)
}

func askPeer() (ip net.IP, port int) {
	var temp string

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

func connectToNetwork(ip net.IP, port int, listenCh chan string) {
	conn1, err := connect(ip, port)
	if err != nil {
		fmt.Println("Connection to the network Succesfull")
	} else {
		fmt.Println(err)
		fmt.Println("Initializing your iwn network on port %d", port)
	}

	go handleConn(conn1, handleConn)

	ln, _ := net.Listen("tcp", ":"+string(port))
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		go handleConn(conn, listenCh)
	}

}

func broadcast(msg string){
	for _, conn := range connArray {
		fmt.Fprintf(conn, msg)
	}
}

func print(writeCh chan string, listenCh chan string) {
	for {
		select {
		case msg := <-writeCh:
			broadcast(msg)
		case msg := <-listenCh:
			if val, found := messages[msg]; val && found {
				fmt.Println(msg)
				broadcast(msg)
			}
		}
	
		messages[msg] = true
	}
}

func write(writeCh chan string){
	var msg string
	for{
		fmt.Scanln(&msg)
		writeCh<-msg
	}
}

func handleConn(conn net.Conn, listenCh chan string){
	defer conn.Close()
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if (err != nil) {
			fmt.Println("Error: " + err.Error())
			return
		} else {
			listenCh<-msg
		}
	}
}
