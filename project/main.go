package main

import (
	"flag"
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
	networking "project/networking"
	"time"
)

func main() {
	peerID := flag.String("peerID", "0", "peerID of the elevator to be created")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	elevio.Init(*peerID, *serverAddr, 4)

	driverCommandChan := make(chan elevio.DriverCommand) // Kan vurdere to separate channels inn??? Vet ikke hva som er best praksis
	buttonChan := make(chan elevio.ButtonEvent)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)
	clearedOrdersChan := make(chan []localsingle.Order)
	localStateChan := make(chan localsingle.ElevatorState)
	peerEventChan := make(chan []ordersync.Peer)
	localOrderChan := make(chan elevio.ButtonEvent)

	// Channels and goroutines for networking
	OrderSyncTx := make(chan ordersync.NetMsg)
	OrderSyncRx := make(chan ordersync.NetMsg)
	PeerMonitorTx := make(chan peermonitor.HeartBeat)
	PeerMonitorRx := make(chan peermonitor.HeartBeat)

	go networking.Run(OrderSyncTx, OrderSyncRx, PeerMonitorTx, PeerMonitorRx)

	go elevio.RunDriver(driverCommandChan)
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	peermonitorConfig := peermonitor.PeerConfig{Timeout: 10 * time.Second, TickInterval: 50 * time.Millisecond} // Vurdere endring? Kanskje unødvendig med egen struct PeerConfig

	go peermonitor.Run(*peerID, peermonitorConfig, PeerMonitorRx, PeerMonitorTx, peerEventChan)

	go ordersync.Run(ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, OrderSyncRx, peerEventChan, localOrderChan, OrderSyncTx, driverCommandChan)
	go localsingle.Run(localOrderChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan)

	select {}
}
