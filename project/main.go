package main

import (
	"context"
	"flag"
	"log"
	elevio "project/elevio"
	localsingle "project/localsingle"
	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
	"project/supervisor"
	"time"
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

	// Configuration
	peermonitorConfig := peermonitor.PeerConfig{
		Timeout:         5 * time.Second,
		TickInterval:    50 * time.Millisecond,
		HeartBeatTicker: 1 * time.Second,
	}

	// Define supervised children
	children := []supervisor.ChildSpec{
		// Elevio routines
		{
			Name: "driver",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				go elevio.RunDriver(driverCommandChan)
				<-ctx.Done()
				return nil
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "poll-buttons",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				go elevio.PollButtons(buttonChan)
				<-ctx.Done()
				return nil
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "poll-floor",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				go elevio.PollFloorSensor(floorChan)
				<-ctx.Done()
				return nil
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "poll-obstruction",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				go elevio.PollObstructionSwitch(obstructionChan)
				<-ctx.Done()
				return nil
			}),
			Restart: supervisor.Permanent,
		},
		// Core system components
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
				return peermonitor.Run(ctx, *peerID, peermonitorConfig, peerMonitorRx, peerMonitorTx, peerEventChan)
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
	sup := supervisor.NewSupervisor(children, supervisor.SupervisorConfig{
		MaxRestarts:  5,
		MaxTime:      30 * time.Second,
		RestartDelay: 200 * time.Millisecond,
	})

	log.Println("Starting elevator system with supervisor")
	if err := sup.Run(ctx); err != nil {
		log.Printf("Supervisor exited with error: %v", err)
	}
}
