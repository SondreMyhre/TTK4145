package main

import (
	"flag"
	"fmt"
	"log"
	config "project/config"
	elevatorcontroller "project/elevatorcontroller"
	elevio "project/elevio"
	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
)

func main() {
	peerID := flag.String("peerID", "0", "Unique identifier for this elevator node in the distributed system")
	serverAddr := flag.String("serverAddr", "localhost:15657", "Address of the elevator simulator server for this node")
	flag.Parse()

	if *peerID == "0" {
		log.Fatal("Not valid peerID.")
	}

	elevio.Init(*serverAddr)

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

	// ---- Initialize elevator node  ----
	go elevio.RunDriver(driverCommandChan)
	go networking.Run(orderSyncTx, peerMonitorTx, orderSyncRx, peerMonitorRx)
	go peermonitor.Run(*peerID, peerMonitorRx, peerEventChan, peerMonitorTx)
	go ordersync.RunWorldview(ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, orderSyncRx, peerEventChan, orderSyncTx, driverCommandChan, worldviewChan)
	go ordersync.RunAssigner(ordersync.ElevID(*peerID), worldviewChan, assignedRequestsChan)
	go elevatorcontroller.Run(assignedRequestsChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan)

	// ---- Print configuration ----
	fmt.Println("\n" + "=== Elevator System Started ===")
	fmt.Println("Elevator ID:              " + *peerID)
	fmt.Println("Server Address:           " + *serverAddr)
	fmt.Println("\n=== Configuration ===")
	fmt.Printf("N_FLOORS:                 %d\n", config.N_FLOORS)
	fmt.Printf("BROADCAST_PORT:           %d\n", config.BROADCAST_PORT)
	fmt.Printf("BROADCAST_ADDR:           %s\n", config.BROADCAST_ADDRESS)
	fmt.Printf("POLL_RATE:                %v\n", config.POLL_RATE)
	fmt.Printf("MOTOR_TIMEOUT:            %v\n", config.MOTORTIMEOUT)
	fmt.Printf("PEER_TIMEOUT:             %v\n", config.PEER_TIMEOUT)
	fmt.Printf("PEER_TICK_INTERVAL:       %v\n", config.PEER_TICK_INTERVAL)
	fmt.Printf("HEARTBEAT_TICK_INTERVAL:  %v\n", config.HEARTBEAT_TICK_INTERVAL)
	fmt.Printf("NETMSG_TICK_INTERVAL:     %v\n", config.NETMSG_TICK_INTERVAL)
	fmt.Println("===========================")

	select {}
}
