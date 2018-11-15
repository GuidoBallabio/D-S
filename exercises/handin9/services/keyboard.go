package services

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	. "../account"
	"../aesrsa"
)

// Write handles the input from keyboard
func Write(listenCh chan<- SignedTransaction, quitCh chan<- struct{}) {
	defer Wg.Done()

	fmt.Println("Insert a transaction as: FromWho ToWho HowMuch each on different lines, then the private key to sign it ")
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
		fmt.Println("Confirm with Secret Key")
		key := aesrsa.KeyFromString(scanKey(scanner))
		st := SignTransaction(t, key)
		listenCh <- st
		fmt.Println("Sent")
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

	tmp, err := strconv.Atoi(amount)
	intAmount := uint64(tmp)

	for err != nil {
		fmt.Println("not valid integer amount")
		scanner.Scan()
		amount := scanner.Text()

		if amount == "quit" {
			return Transaction{}, true
		}

		tmp, err = strconv.Atoi(amount)
		intAmount = uint64(tmp)
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
		if buf == "quit" {
			return buf
		}
		scanner.Scan()
		buf = scanner.Text()
	}

	key := buf + "\n"

	scanner.Scan()
	buf = scanner.Text()

	for buf != "-----END KEY-----" {
		if buf == "quit" {
			return buf
		}
		key += buf

		scanner.Scan()
		buf = scanner.Text()
	}

	key += "\n" + buf

	return key
}

func GatherKeys() []string {
	p := []string{}
	for peers := range PeerList.Iter() {
		p = append(p, peers.PubKey)
	}

	l := Tree.GetAccountNumbers()

	l = append(l, p)

	return l
}
