package TransportUDP

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// find out what information NetMsg should carry
// has to contain all information to be sent over the network
// needs the orderMatrix from orderSync and a peer-message from peerMonitor
type NetMsg struct {
	Message string
	Iter int
}

// this is what will run in main.go
func Run(tx <-chan NetMsg, rx chan<- NetMsg) {
	
	// Reads messages from tx chan and broadcast them over the network
	// go broadcastMessages(tx <- chan NetMsg)

	// Reads messages from the network, decodes them and, send over respective channels
	// go recieveMessages(rx chan<- NetMsg)
}