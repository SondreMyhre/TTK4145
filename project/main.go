package main

import (
	"flag"
	"log"
	elevatorcontroller "project/elevatorcontroller"
	elevio "project/elevio"
	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
)

func main() {
	peerID := flag.String("peerID", "0", "peerID of the elevator to be created")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	if *peerID == "0" {
		log.Fatal("Not valid peerID.")
	}

	elevio.Init(*serverAddr, elevatorcontroller.N_FLOORS)

	// ---- Initialize channels ---
	// Hardware communication
	buttonChan := make(chan elevio.ButtonEvent, 4)
	floorChan := make(chan int, 1)
	obstructionChan := make(chan bool, 1)
	driverCommandChan := make(chan elevio.DriverCommand, 10)

	// Networking
	orderSyncTx := make(chan ordersync.NetMsg, 1)
	orderSyncRx := make(chan ordersync.NetMsg, 2)
	peerMonitorTx := make(chan peermonitor.HeartBeat, 1)
	peerMonitorRx := make(chan peermonitor.HeartBeat, 2)

	// Elevator control system
	assignedRequestsChan := make(chan elevatorcontroller.RequestMatrix, 1)
	clearedOrdersChan := make(chan []elevatorcontroller.Order, 1)
	localStateChan := make(chan elevatorcontroller.ElevatorState, 1)
	worldviewChan := make(chan ordersync.WorldviewMsg, 1)
	peerEventChan := make(chan []ordersync.PeerUpdate)

	// ---- Initialize polling routines ----
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	// ---- Initialize supervised elevator node  ----
	go elevio.RunDriver(driverCommandChan)
	go networking.Run(orderSyncTx, peerMonitorTx, orderSyncRx, peerMonitorRx)
	go peermonitor.Run(*peerID, peerMonitorRx, peerEventChan, peerMonitorTx)
	go ordersync.RunWorldview(ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, orderSyncRx, peerEventChan, orderSyncTx, driverCommandChan, worldviewChan)
	go ordersync.RunAssigner(ordersync.ElevID(*peerID), worldviewChan, assignedRequestsChan)
	go elevatorcontroller.Run(assignedRequestsChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan)

	select {}
}
