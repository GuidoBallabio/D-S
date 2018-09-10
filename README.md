# Distributed Systems and Security

## Exercise 2.3 
Implement a toy peer-to-peer network.

This exercise asks you to program in Go a toy example of a peer-to-peer flooding network for send-
ing strings around. The peer-to-peer network should then be used to build a
distributed chat room. The chat room client should work as follows:

1. It runs as a command line program.
2. When it starts up it asks for the IP address and port number of an existing peer on the network. If the IP address or port is invalid or no peer is found at the address, the client starts its own new network with only itself as member.
3. Then the client prints its own IP address and the port on which it waits for connections.
4. Then it will iteratively prompt the user for text strings.
5. When the user types a text string at any connected client, then it will eventually be printed at all other clients.
6. Only the text string should be printed, no information about who sent it.

The system should be implemented as follows:

1. When a client connects to an existing peer, it will keep a TCP connection to that peer.
2. Then the client opens its own port where it waits for incoming TCP connections.
3. All the connections will be treated the same, they will be used for both sending and receiving strings.
4. It keeps a set of messages that it already sent. In Go you can make a set as a map var MessagesSent map[string]bool. You just map the strings that were sent to true. Initially all of them are set to false, so the set is initially empty, as it should be.
5. When a string is typed by the user or a string arrives on any of its connections, the client checks if it is already sent. If so, it does nothing. Otherwise it adds it to MessagesSent and then sends it on all its connections. (Remember concurrency control. Probably several go-routines will access the set at the same time. Make sure that does not give problems.)
6. Whenever a message is added to MessagesSent, also print it for the user to see.

Add this to your report:

1. Test you system and describe how you tested it.
2. Argue that you system has eventual consistency in the sense that if all clients stop typing, then eventually all clients will print the same set of strings.

## Exercise 4.5
Implement a simple peer-to-peer ledger

Modify your code from Exercise 2.3 to add the following features:
1. The system now no longer broadcasts strings and prints them. Instead it implements a distributed ledger. Each client keeps a Ledger.
2. Each client can make Transactions. When they do all other peers eventually update their ledger with the transaction.
3. The system should ensure eventual consistency, i.e., if all clients stop sending transactions, then all ledgers will eventually be in the same correct state.

```go
package account
type Transaction struct {
    ID string
    From string
    To string
    Amount int
}

func (l *Ledger) Transaction(t *Transaction) {
    l.lock.Lock() ; defer l.lock.Unlock()
    l.Accounts[t.From] -= t.Amount
    l.Accounts[t.To] += t.Amount
}
```

4. Your system only has to work if there are two phases: first all the peers connect,
then they make transactions. But if you want to accommodate for later comers
a way to do it is to let each client keep a list of all the transactions it saw and
then forward them to clients that log in late.

Implement as follows:
1. Keep a sorted list of peers.
2. When connecting to a peer, ask for its list of peers.
3. Then add yourself to your own list.
4. Then connect to the ten peers after you on the list (with wrap around).
5. Then broadcast your own presence.
6. When a new presence is broadcast, add it to your list of peers.
7. When a transaction is made, broadcast the Transaction object.
8. When a transaction is received, update the local Ledger object.

Add this to your report:
1. Test you system and describe how you tested it.
2. Discuss whether connection to the next ten peers is a good strategy with respect to connectivity. In particular, if the network has 1000 peers, how many connections need to break to partition the network?
3. Argue that your system has eventual consistency if all processes are correct and the system is run in two-phase mode.
4. Assume we made the following change to the system: When a transaction arrives, it is rejected if the receiving account goes below 0. Does your system
still have eventual consistency? Why or why not?