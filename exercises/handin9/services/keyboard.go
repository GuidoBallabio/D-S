package services

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	. "../account"
	"../aesrsa"
)

var abbreviations map[string]string

// Write handles the input from keyboard
func Write(listenCh chan<- SignedTransaction, quitCh chan<- struct{}) {
	defer Wg.Done()

	fmt.Println("Insert a transaction as: FromWho ToWho HowMuch each on different lines, then the private key to sign it ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	abbreviations = map[string]string{}

	for {
		t, quit := askTransaction(scanner)
		if quit {
			fmt.Println("quitting...")
			close(quitCh)
			break //Done
		}
		t = attachNextID(t)
		fmt.Println("Confirm with Secret Key")
		key := aesrsa.KeyFromString(scanPrivKey(scanner))
		st := SignTransaction(t, key)
		listenCh <- st
		fmt.Println("Sent")
	}
}

func askTransaction(scanner *bufio.Scanner) (Transaction, bool) {

	printKeys()

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
	for {
		scanner.Scan()
		buf := scanner.Text()

		if buf == "quit" {
			return buf
		}

		val, found := abbreviations[buf]

		if found {
			return val
		}

		fmt.Println("Invalid key input! Please write a valide value")
	}
}

func printKeys() {
	populateAbbreviation()

	l := len(abbreviations)

	for i := 0; i < l; i++ {
		fmt.Printf("Input: " + strconv.Itoa(i) + "\t| Account: " + abbreviations[strconv.Itoa(i)][30:39] + "\n")
	}
}

// PopulateAbbreviation inserts the keys and their abbreviation in the abbreviation map
func populateAbbreviation() {
	p := gatherKeys()

	for i, c := range p {
		abbreviations[strconv.Itoa(i)] = c
	}
}

// GatherKeys returns all the pubkeys of the clients
func gatherKeys() []string {
	l := Tree.GetAccountNumbers()

	for p := range PeerList.Iter() {

		found := false

		for _, p1 := range l {
			if p1 == p.PubKey {
				found = true
				break
			}
		}

		if !found {
			l = append(l, p.PubKey)
		}
	}

	return l
}

func scanPrivKey(scanner *bufio.Scanner) string {
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
