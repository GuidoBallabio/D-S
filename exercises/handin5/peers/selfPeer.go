package peers

import (
	"net"
)

// GetLocalPeer returns the peer of the local machine
func GetLocalPeer(listeningPort int, pubKey string) Peer {
	p := newPeer(getLocalIP().String(), listeningPort, nil)
	p.AddPubKey(pubKey)
	return p
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
