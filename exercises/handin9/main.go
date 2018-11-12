package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/util/wait"

	. "./account"
	"./aesrsa"
	. "./peers"
	serv "./services"
)

var localKeys *aesrsa.RSAKeyPair

func main() {

	var (
		sk = kingpin.Flag("public-key", "Use predefined keys: private key file").Short('c').String()
		pk = kingpin.Flag("secret-key", "Use predefined keys: public key file").Short('s').String()

		server     = kingpin.Command("server", "Create your own network")
		portServer = server.Flag("port", "Port of server.").Short('p').Default("4444").Int()

		peer = kingpin.Command("peer", "Connect to a peer in a pre-existing network.")
		ip   = peer.Arg("ip", "IP address of Peer.").Required().IP()
		port = peer.Arg("port", "Port of Peer.").Required().Int()
	)

	kingpin.CommandLine.HelpFlag.Short('h')

	cmd := kingpin.Parse()

	if *sk != "" && *pk != "" {
		skey, _ := ioutil.ReadFile(*sk)
		pkey, _ := ioutil.ReadFile(*pk)

		localKeys = &aesrsa.RSAKeyPair{
			Public:  aesrsa.KeyFromString(string(pkey)),
			Private: aesrsa.KeyFromString(string(skey))}
	} else {
		var err error
		localKeys, err = aesrsa.KeyGen(2048)
		if err != nil {
			panic(err.Error())
		}
	}

	fmt.Println("Your secret key is:")
	fmt.Println(aesrsa.KeyToString(localKeys.Private))
	fmt.Println("Your public key is:")
	fmt.Println(aesrsa.KeyToString(localKeys.Public))

	serv.InitNetwork()
	listenCh := make(chan SignedTransaction)
	blockCh := make(chan SignedBlock)

	switch cmd {
	case "server":
		serv.CreateNetwork(*portServer, listenCh, blockCh, localKeys.Public)
		serv.BecomeSequencer()

	case "peer":
		firstPeer := Peer{
			IP:   ip.String(),
			Port: *port}
		serv.ConnectToNetwork(firstPeer, listenCh, blockCh, localKeys.Public)
	}

	startServices(listenCh, blockCh)
}

func startServices(listenCh chan SignedTransaction, blockCh chan SignedBlock) {
	sequencerCh := make(chan Transaction)
	quitCh := make(chan struct{})

	serv.Wg.Add(1)
	go serv.BeServer(listenCh, blockCh, quitCh)

	wait.PollInfinite(time.Second*10, wait.ConditionFunc(func() (bool, error) {
		return serv.PeerList.Length() > 1, nil
	}))

	serv.Wg.Add(3)
	go serv.ProcessTransactions(listenCh, sequencerCh, quitCh)
	go serv.ProcessBlocks(blockCh, quitCh)
	go serv.Write(listenCh, quitCh)

	if serv.CheckIfSequencer() {
		serv.Wg.Add(1)
		go serv.BeSequencer(sequencerCh, blockCh, quitCh)
	}

	<-quitCh
	serv.Connect(&serv.LocalPeer)
	serv.Wg.Wait()
}
