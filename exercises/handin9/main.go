package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/util/wait"

	. "./account"
	"./aesrsa"
	bt "./blocktree"
	. "./peers"
	serv "./services"
)

var localKeys *aesrsa.RSAKeyPair

func main() {

	var (
		keys = kingpin.Flag("keys", "Use predefined keys.").Short('k').String()
		pw   = kingpin.Flag("password", "Password for the keys").Short('x').String()
		dir  = kingpin.Flag("dir", "Directory for the founders' keys. (Must already exist)").Short('d').Default("founders").String()

		server     = kingpin.Command("server", "Create your own network.")
		portServer = server.Flag("port", "Port of server.").Short('p').Default("4444").Int()

		peer = kingpin.Command("peer", "Connect to a peer in a pre-existing network.")
		ip   = peer.Arg("ip", "IP address of Peer.").Required().IP()
		port = peer.Arg("port", "Port of Peer.").Required().Int()
	)

	kingpin.CommandLine.HelpFlag.Short('h')

	cmd := kingpin.Parse()

	serv.InitNetwork()

	listenCh := make(chan SignedTransaction)
	blockCh := make(chan bt.SignedNode)

	switch cmd {
	case "server":
		if _, err := os.Stat(*dir); err != nil && os.IsNotExist(err) {
			os.Mkdir(*dir, 0755)
			GenerateFounders(10, *dir)
		}
		initKeys(*keys, *pw)
		serv.CreateNetwork(*portServer, listenCh, blockCh, localKeys.Public)
	case "peer":
		firstPeer := Peer{
			IP:   ip.String(),
			Port: *port}
		initKeys(*keys, *pw)
		serv.ConnectToNetwork(firstPeer, listenCh, blockCh, localKeys.Public)
	}

	InitBlockChain(*dir)
	startServices(listenCh, blockCh)
}

func startServices(listenCh chan SignedTransaction, blockCh chan bt.SignedNode) {
	sequencerCh := make(chan Transaction)
	quitCh := make(chan struct{})

	serv.Wg.Add(1)
	go serv.BeServer(listenCh, blockCh, quitCh)

	wait.PollInfinite(time.Second*5, wait.ConditionFunc(func() (bool, error) {
		return serv.PeerList.Length() > 1, nil
	}))

	serv.Wg.Add(3)
	go serv.ProcessTransactions(listenCh, sequencerCh, quitCh)
	go serv.ProcessNodes(sequencerCh, blockCh, localKeys, quitCh)
	go serv.Write(listenCh, quitCh)

	<-quitCh
	serv.Connect(&serv.LocalPeer)
	serv.Wg.Wait()
}

/////////// Init Functions ///////////

func initKeys(keys, pw string) {
	if keys != "" && pw != "" {
		localKeys = aesrsa.ReadKeyPair(keys, pw)
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
}

// GenerateFounders creates n founders' keys and and write the pairs and just the public in files
func GenerateFounders(n int, dir string) {
	for i := 0; i < n; i++ {
		keys, err := aesrsa.KeyGen(2048)
		if err != nil {
			panic(err)
		}

		privFile := fmt.Sprintf(dir+"/"+"founder-%d.keys", i)
		pw := fmt.Sprintf("password-%d", i)
		aesrsa.StoreKeyPair(keys, privFile, pw)

		pubFile := fmt.Sprintf(dir+"/"+"founder-%d.cert", i)
		pw2 := "nopassword"
		aesrsa.StoreKey(keys.Public, pubFile, pw2)
	}
}

// InitBlockChain make the necessary preparetions for the blockchain
func InitBlockChain(dir string) {
	founders := ReadPublicKeys(10, dir)
	tl := InitTransactions(founders)
	serv.Tree = bt.NewTree(tl)
}

// ReadPublicKeys returns the list of founders' public keys
func ReadPublicKeys(n int, dir string) []string {
	var founders = []string{}

	for i := 0; i < n; i++ {

		pubFile := fmt.Sprintf(dir+"/"+"founder-%d.cert", i)
		key := aesrsa.ReadKey(pubFile, "nopassword")

		founders = append(founders, aesrsa.KeyToString(*key))
	}

	return founders
}

// InitTransactions returns the list of genesis' transactions given a list of peers (pubKeys)
func InitTransactions(founders []string) []Transaction {
	var list = []Transaction{}

	for i, f := range founders {
		id := fmt.Sprintf("Genesis - %d", i)
		list = append(list, NewTransaction(id, "Genesis", f, 1e6))
	}

	return list
}
