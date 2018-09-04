package main

import (
	"fmt"
	"net"
)

func main() {
	ip, port := askPeer()

	var texts = make(chan string)
	go connectToNetwork(ip, port, texts)
	go chat(texts)
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

func connectToNetwork(ip net.IP, port int, texts chan string) {
	conn1, err := connect(ip, port)
	if err != nil {
		fmt.Println("Connection to the network Succesfull")
	} else {
		fmt.Println(err)
		fmt.Println("Initializing your iwn network on port %d", port)
	}

	inCh := make(chan string)
	//go broadcaster/receiver with texts and one channel to which it will listen and all the channel to output

	ln, _ := net.Listen("tcp", ":"+string(port))
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		sendCh := createAndAddChannel()
		go handleConn(conn, sendCh, inCh)
	}

}

func chat(texts chan string) {

}
