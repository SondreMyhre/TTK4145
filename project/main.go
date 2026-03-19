// Package main is the entry point for the distributed elevator control system.
//
// The system is structured as a collection of independent modules that communicate via channels.
// This follows a peer-to-peer architecture where each elevator node is responsible for:
//   - Running its own FSM (Finite State Machine) to control the elevator locally
//   - Sharing its state with other peers via broadcast
//   - Coordinating order distribution via a distributed assignment algorithm (HRA)
//
// Architecture Overview:
//
// HARDWARE LAYER (elevio):
//
//	└─ Polls physical hardware (buttons, floor sensors, obstruction switches)
//	   └─ Sends events to elevatorcontroller and ordersync
//
// LOCAL ELEVATOR CONTROLLER (elevatorcontroller):
//
//	└─ Runs the local elevator FSM (Idle/DoorOpen/Moving)
//	├─ Receives: assigned orders, floor events, obstruction events
//	├─ Maintains: local elevator state, local request matrix
//	└─ Sends: driver commands to hardware, cleared orders, local state
//
// DISTRIBUTED ORDER COORDINATION (ordersync):
//
//	├─ Maintains the global view of all hall orders
//	├─ Receives: local buttons, elevator states, network messages, peer events
//	├─ Computes: which elevator should serve which orders (via HRA algorithm)
//	└─ Sends: assigned orders to local controller, network messages to peers
//
// PEER MONITORING (peermonitor):
//
//	├─ Tracks heartbeats from other peer elevators
//	├─ Detects timeouts and failures
//	└─ Notifies ordersync when peers go down
//
// NETWORK TRANSPORT (networking):
//
//	└─ UDP broadcast communication between all peers
//	   ├─ Receives messages from ordersync and peermonitor
//	   └─ Forwards network messages to ordersync and peermonitor
//
// Information Flow (example: user presses button):
//  1. Button press detected by elevio.PollButtons
//  2. Button event sent to ordersync (via local button handler)
//  3. ordersync updates hall order state
//  4. ordersync broadcasts new state via networking
//  5. HRA algorithm assigns orders to elevators
//  6. My assigned orders sent to elevatorcontroller
//  7. elevatorcontroller executes FSM, updates state
//  8. When order cleared, notification sent back to ordersync
//  9. ordersync broadcasts cleared order, updates global state
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

	// ---- Initialize all communication channels ----
	// These channels form the backbone of the system, connecting all modules.
	// Each channel is unidirectional to enforce a clear information flow direction.

	// Hardware I/O Channels: elevio -> elevatorcontroller & ordersync
	// Capacities are sized to handle burst events without blocking
	buttonChan := make(chan elevio.ButtonEvent, 4)           // User button presses (cab & hall)
	floorChan := make(chan int, 1)                           // Floor sensor updates
	obstructionChan := make(chan bool, 1)                    // Door obstruction detection
	driverCommandChan := make(chan elevio.DriverCommand, 10) // Motor, door lamp, floor indicator

	// Network Transport Channels: networking <-> ordersync & peermonitor
	// orderSyncRx and peerMonitorRx receive from the same networking module
	orderSyncTx := make(chan ordersync.NetMsg, 1)        // Order state broadcasts to network
	orderSyncRx := make(chan ordersync.NetMsg, 2)        // Receive order state from other elevators
	peerMonitorTx := make(chan peermonitor.HeartBeat, 1) // Heartbeats from peermonitor
	peerMonitorRx := make(chan peermonitor.HeartBeat, 2) // Heartbeats from other elevators

	// Elevator Control System Channels: elevatorcontroller <-> ordersync
	// This channel set forms the core control loop
	assignedRequestsChan := make(chan elevatorcontroller.RequestMatrix, 1) // HRA assignments: ordersync -> controller
	clearedOrdersChan := make(chan []elevatorcontroller.Order, 1)          // Completed orders: controller -> ordersync
	localStateChan := make(chan elevatorcontroller.ElevatorState, 1)       // Local state: controller -> ordersync
	worldviewChan := make(chan ordersync.WorldviewMsg, 1)                  // Global state: ordersync internal
	peerEventChan := make(chan []ordersync.PeerUpdate)

	// ---- Initialize polling routines ----
	// These goroutines continuously monitor hardware and emit events to channels
	// Critical: These must run before the system processes any orders
	go elevio.PollButtons(buttonChan)                // Continuous button polling
	go elevio.PollFloorSensor(floorChan)             // Continuous floor detection
	go elevio.PollObstructionSwitch(obstructionChan) // Continuous obstruction monitoring

	// ---- Initialize core system modules ----
	// Each module runs independently as a goroutine and communicates via channels.
	// Order of initialization matters: transport must start before coordinator modules.

	// 1. HARDWARE DRIVER: Executes motor, door, and lamp commands
	//    Reads from: driverCommandChan (from elevatorcontroller)
	go elevio.RunDriver(driverCommandChan)

	// 2. NETWORK TRANSPORT: UDP broadcast communication
	//    Reads from: orderSyncTx (order state), peerMonitorTx (heartbeats)
	//    Writes to: orderSyncRx (new orders from peers), peerMonitorRx (peer heartbeats)
	go networking.Run(orderSyncTx, peerMonitorTx, orderSyncRx, peerMonitorRx)

	// 3. PEER MONITOR: Detects peer failures via heartbeat timeouts
	//    Reads from: peerMonitorRx (heartbeats from other elevators)
	//    Writes to: peerEventChan (dead/alive peer notifications)
	go peermonitor.Run(*peerID, peerMonitorRx, peerEventChan, peerMonitorTx)

	// 4. WORLDVIEW COORDINATOR: Maintains global distributed order state
	//    Reads from: ordersync.ReceiveOrderSync (network), localStateChan, clearedOrdersChan, peerEventChan
	//    Writes to: worldviewChan, orderSyncTx (broadcast to peers), driverCommandChan (for lamps)
	go ordersync.RunWorldview(ordersync.ElevID(*peerID), buttonChan, localStateChan, clearedOrdersChan, orderSyncRx, peerEventChan, orderSyncTx, driverCommandChan, worldviewChan)

	// 5. ORDER ASSIGNER: Runs HRA algorithm to assign orders to elevators
	//    Reads from: worldviewChan (global state)
	//    Writes to: assignedRequestsChan (my assigned requests for this elevator)
	go ordersync.RunAssigner(ordersync.ElevID(*peerID), worldviewChan, assignedRequestsChan)

	// 6. LOCAL ELEVATOR CONTROLLER: Runs the FSM for this elevator
	//    Reads from: assignedRequestsChan, floorChan, obstructionChan
	//    Writes to: driverCommandChan, clearedOrdersChan, localStateChan
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
