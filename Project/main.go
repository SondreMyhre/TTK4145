package main

import (
	elevio "Project/ElevIO"
	localsingle "Project/LocalSingleElevator"
	transportUDP "Project/TransportUDP"
	peermonitor "Project/PeerMonitor" 
	ordersync "Project/OrderSync" 
	"flag"
	"log"
)

func main() {
	peerID := flag.Int("peerID", 0, "peerID of the elevator to be created")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()
	if *peerID == 0 {
		log.Fatal("Not valid peerID.")
	}

	elevio.Init(*serverAddr, 4)

	driverCommandChan := make(chan elevio.DriverCommand)
	buttonChan := make(chan elevio.ButtonEvent)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)
	clearedOrdersChan := make(chan []localsingle.Order)
	localStateChan := make(chan localsingle.LocalSingleElevator)

	// Channels and goroutines for networking
	PeerMonitorTx := make(chan peermonitor.RecoveryMsg)
	PeerMonitorRecMsgRx := make(chan peermonitor.RecoveryMsg)
	PeerMonitorNetMsgRx := make(chan ordersync.NetMsg)
	
	OrderSyncTx := make(chan ordersync.NetMsg)
	OrderSyncRx := make(chan ordersync.NetMsg)

	go transportUDP.Run(*peerID, PeerMonitorTx, PeerMonitorRecMsgRx, PeerMonitorNetMsgRx, OrderSyncTx, OrderSyncRx)

	go elevio.RunDriver(driverCommandChan)
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	go func() { for range clearedOrdersChan {} }()	// OrderSync ikke implementert
    go func() { for range localStateChan {} }()		// OrderSync ikke implementert

	go localsingle.Run(buttonChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan) // localOrderChan = buttonChan

	select {}
}
