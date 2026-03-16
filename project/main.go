package main

import (
	"context"
	"flag"
	"log"
	elevio "project/elevio"
	localsingle "project/elevatorcontroller"
	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
	"project/supervisor"
)

func main() {
	peerID := flag.String("peerID", "0", "peerID of the elevator to be created")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	if *peerID == "0" {
		log.Fatal("Not valid peerID.")
	}

	elevio.Init(*serverAddr, localsingle.N_FLOORS)

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

	// Channel between worldview and assigner in ordersync
	worldviewChan := make(chan ordersync.WorldviewMsg, 1)

	// elevio polling routines
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	// Define supervised children
	children := []supervisor.ChildSpec{
		{
			Name: "elevio",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				elevio.RunDriver(ctx, driverCommandChan)
				return nil
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "networking",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				networking.Run(ctx, orderSyncTx, orderSyncRx, peerMonitorTx, peerMonitorRx)
				return nil
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "peermonitor",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				return peermonitor.Run(ctx, *peerID, peerMonitorRx, peerEventChan, peerMonitorTx)
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "worldview",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				return ordersync.RunWorldView(ctx, ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, orderSyncRx, peerEventChan, orderSyncTx, driverCommandChan, worldviewChan)
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "assigner",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				return ordersync.RunAssigner(ctx, ordersync.ElevID(*peerID), worldviewChan, requestMatrixChan)
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "localsingle",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				return localsingle.Run(ctx, requestMatrixChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan)
			}),
			Restart: supervisor.Permanent,
		},
	}

	// Create and run supervisor
	sup := supervisor.NewSupervisor(children)

	log.Println("Starting elevator system with supervisor")
	if err := sup.Run(ctx); err != nil {
		log.Printf("Supervisor exited with error: %v", err)
	}
}
