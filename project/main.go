package main

import (
	"context"
	"flag"
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
	"time"
)

func main() {
	peerID := flag.String("peerID", "0", "peerID of the elevator to be created")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	elevio.Init(*peerID, *serverAddr, 4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// elevio channels
	buttonChan := make(chan elevio.ButtonEvent)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)
	driverCommandChan := make(chan elevio.DriverCommand)

	// localsingle channels
	requestMatrixChan := make(chan localsingle.RequestMatrix)
	clearedOrdersChan := make(chan []localsingle.Order)
	localStateChan := make(chan localsingle.ElevatorState)

	// networking channels
	orderSyncTx := make(chan ordersync.NetMsg)
	orderSyncRx := make(chan ordersync.NetMsg)
	peerMonitorTx := make(chan peermonitor.HeartBeat)
	peerMonitorRx := make(chan peermonitor.HeartBeat)

	// Channel between ordersync and peermonitor
	peerEventChan := make(chan []ordersync.PeerUpdate)

	worldviewChan := make(chan ordersync.WorldviewMsg, 1) 


	// Routines for single elevator operations
	go elevio.RunDriver(driverCommandChan)
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	// Channels for distributed system
	peermonitorConfig := peermonitor.PeerConfig{Timeout: 10 * time.Second, TickInterval: 50 * time.Millisecond, HeartBeatTicker: 1 * time.Second} // Vurdere endring? Kanskje unødvendig med egen struct PeerConfig
	go peermonitor.Run(*peerID, ctx, peermonitorConfig, peerMonitorRx, peerMonitorTx, peerEventChan)
	go networking.Run(ctx, orderSyncTx, orderSyncRx, peerMonitorTx, peerMonitorRx)
	go ordersync.RunWorldView(ctx, ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, orderSyncRx, peerEventChan, orderSyncTx, driverCommandChan, worldviewChan)
	go ordersync.RunAssigner(ctx, ordersync.ElevID(*peerID), worldviewChan, requestMatrixChan)
	go localsingle.Run(ctx, requestMatrixChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan)

	select {}
}
