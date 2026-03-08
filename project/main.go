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
	buttonChan := make(chan elevio.ButtonEvent, 10)
	floorChan := make(chan int, 1)
	obstructionChan := make(chan bool, 1)
	driverCommandChan := make(chan elevio.DriverCommand, 10)

	// localsingle channels
	requestMatrixChan := make(chan localsingle.RequestMatrix, 1)
	clearedOrdersChan := make(chan []localsingle.Order, 10)
	localStateChan := make(chan localsingle.ElevatorState, 1)

	// networking channels
	orderSyncTx := make(chan ordersync.NetMsg, 10)
	orderSyncRx := make(chan ordersync.NetMsg, 10)
	peerMonitorTx := make(chan peermonitor.HeartBeat, 10)
	peerMonitorRx := make(chan peermonitor.HeartBeat, 10)

	// Channel between ordersync and peermonitor
	peerEventChan := make(chan []ordersync.PeerUpdate, 10)

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
