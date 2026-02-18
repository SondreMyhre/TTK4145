package main

import (
	elevio "Project/ElevIO"
	localsingle "Project/LocalSingleElevator"
	transportUDP "Project/TransportUDP"
	"flag"
)

func main() {
	// TO-DO: add port int and id string as flags
	// maybe only id string is necessary, then port id can be calculated with the id
	// peerID := flag.Int("peerID", "")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	elevio.Init(*serverAddr, 4)

	driverCommandChan := make(chan elevio.DriverCommand)
	buttonChan := make(chan elevio.ButtonEvent)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)
	clearedOrdersChan := make(chan []localsingle.Order)
	localStateChan := make(chan localsingle.LocalSingleElevator)

	// Channels and goroutines for networking
	PeerMonitorTx := make(chan transportUDP.PeerMonitorMsg)
	PeerMonitorRx := make(chan transportUDP.PeerMonitorMsg)

	OrderSyncTx := make(chan transportUDP.OrderSyncMsg)
	OrderSyncRx := make(chan transportUDP.OrderSyncMsg)
		// TO-DO: maybe run transportUDP.init() function to set port-number
	go transportUDP.Run(PeerMonitorTx, OrderSyncTx, PeerMonitorRx, OrderSyncRx, port)


	go elevio.RunDriver(driverCommandChan)
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	go func() { for range clearedOrdersChan {} }()	// OrderSync ikke implementert
    go func() { for range localStateChan {} }()		// OrderSync ikke implementert

	go localsingle.Run(buttonChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan) // localOrderChan = buttonChan

	select {}
}
