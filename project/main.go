package main

import (
	"flag"
	"log"
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
	ordersync "project/ordersync"
	// peermonitor "project/peermonitor"
	transportUDP "project/transportudp"
	"strconv"
)

func main() {
	peerID := flag.String("peerID", "0", "peerID of the elevator to be created")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()
	if *peerID == "0" {
		log.Fatal("Not valid peerID.")
	}

	peerIDInt, _ := strconv.Atoi(*peerID)

	elevio.Init(*serverAddr, 4)

	driverCommandChan := make(chan elevio.DriverCommand)
	buttonChan := make(chan elevio.ButtonEvent)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)
	clearedOrdersChan := make(chan []localsingle.Order)
	localStateChan := make(chan localsingle.LocalSingleElevator)
	peerEventChan := make(chan []ordersync.Peer)
	localOrderChan := make(chan elevio.ButtonEvent)

	// Channels and goroutines for networking
	// PeerMonitorTx := make(chan peermonitor.RecoveryMsg)
	// PeerMonitorRecMsgRx := make(chan peermonitor.RecoveryMsg)
	// PeerMonitorNetMsgRx := make(chan ordersync.NetMsg)

	OrderSyncTx := make(chan ordersync.NetMsg)
	OrderSyncRx := make(chan ordersync.NetMsg)

	go transportUDP.Run(peerIDInt, OrderSyncTx, OrderSyncRx) // ElevID måtte være string

	go elevio.RunDriver(driverCommandChan)
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	go ordersync.Run(ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, OrderSyncRx, peerEventChan, localOrderChan, OrderSyncTx, driverCommandChan)
	go localsingle.Run(localOrderChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan)

	select {}
}
