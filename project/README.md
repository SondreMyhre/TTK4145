# TTK4145 Elevator Project

This project implements the elevator project in TTK4145

## Overview

The project consists of several main modules:

- **OrderSync**  
  Responsible for distributed synchronization and assignment of hall orders between all elevators. Maintains the global state of hall orders, determines which elevator should take which order, and controls the hall lamps.

- **LocalSingleElevator**  
  Runs the local elevator finite state machine, which controls motor, door, and floor indicator via driverCommands to elvio over channel. Receives assigned orders from OrderSync and reports back when orders are completed.

- **PeerMonitor**  
  Detects peer failures and restarts based on heartbeat timeouts, and notifies ordersync.

- **TransportUDP**  
  Handles UDP broadcast communication between all nodes for order and state synchronization.

- **Elevio**  
  Provides the interface to the elevator simulator hardware.

## Running the System

1. Start a SimElevatorServer for each elevator, for example:
    ```
    ./SimElevatorServer --port 15657
    ./SimElevatorServer --port 15658
    ```

2. Start an elevator process for each elevator:
    ```
    go run main.go -peerID 1 -serverAddr localhost:15657
    go run main.go -peerID 2 -serverAddr localhost:15658
    ```