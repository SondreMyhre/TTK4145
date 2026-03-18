package main

import (
	"context"
	"flag"
	"log"
	elevatorcontroller "project/elevatorcontroller"
	elevio "project/elevio"
	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
	supervisor "project/supervisor"
)

func main() {
	peerID := flag.String("peerID", "0", "peerID of the elevator to be created")
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	if *peerID == "0" {
		log.Fatal("Not valid peerID.")
	}

	elevio.Init(*serverAddr, elevatorcontroller.N_FLOORS)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
				networking.Run(ctx, orderSyncTx, peerMonitorTx, orderSyncRx, peerMonitorRx)
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
				return ordersync.RunWorldview(ctx, ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, orderSyncRx, peerEventChan, orderSyncTx, driverCommandChan, worldviewChan)
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "assigner",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				return ordersync.RunAssigner(ctx, ordersync.ElevID(*peerID), worldviewChan, assignedRequestsChan)
			}),
			Restart: supervisor.Permanent,
		},
		{
			Name: "elevatorcontroller",
			Worker: supervisor.WorkerFunc(func(ctx context.Context) error {
				return elevatorcontroller.Run(ctx, assignedRequestsChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan)
			}),
			Restart: supervisor.Permanent,
		},
	}

	// ---- Run supervisor ----
	sup := supervisor.NewSupervisor(children)

	log.Println("Starting elevator system with supervisor")
	if err := sup.Run(ctx); err != nil {
		log.Printf("Supervisor exited with error: %v", err)
	}
}
