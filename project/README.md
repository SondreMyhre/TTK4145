# TTK4145 Distributed Elevator Control System

This project presents a distributed elevator control system, managing multiple elevators, built in Go and was part of the TTK4145 elevator lab assignment.

## System Architecture

This elevator system is designed as a **peer-to-peer distributed system** where each elevator node operates independently but coordinates via UDP broadcast messages. The system uses a **Hall Request Assignment (HRA) algorithm**, created by [klasbo](https://github.com/klasbo) to optimally distribute orders among elevators.

## System Modules

### 1. **ElevIO** - Hardware Interface
Provides the low-level interface to the elevator hardware.

**Responsibility:**
- Polls hardware sensors (floor sensor, buttons, obstruction switch)
- Executes motor commands and lamp controls

**Inputs (receive-only channels):**
- `driverCommandChan` - Commands to motor, door, and floor indicator

**Outputs (send-only channels):**
- `buttonChan` - Button press events (hall and cab)
- `floorChan` - Current floor updates
- `obstructionChan` - Obstruction switch events

---

### 2. **ElevatorController** - Local FSM (Finite State Machine)
Runs the local elevator's state machine and handles local request execution.

**Responsibility:**
- Maintains local elevator state (floor, direction, behavior)
- Executes FSM logic (Idle → Moving → DoorOpen)
- Manages door timers and motor timeouts
- Decides when to stop and open doors
- Clears orders when arriving at destination floors
- Never makes global decisions; only executes assigned orders

**State ownership:**
- `ElevatorState`: Current state of this elevator (floor, direction, behavior, obstruction, motor_stuck)
- `RequestMatrix`: Which orders THIS elevator should execute (received from OrderSync)
- `DoorTimer`: Time remaining until door closes

**Inputs (receive-only channels):**
- `assignedRequestsChan` - Orders this elevator should execute (from OrderSync via HRA)
- `floorChan` - Floor sensor events (from ElevIO)
- `obstructionChan` - Obstruction events (from ElevIO)

**Outputs (send-only channels):**
- `driverCommandChan` - Motor, door lamp, and floor indicator commands
- `clearedOrdersChan` - Orders that were completed at this floor
- `localStateChan` - Current state for broadcast to other elevators

---

### 3. **OrderSync** - Distributed Order Coordination
Manages distributed consensus on all elevator orders and computes HRA assignments. OrderSync is split into two independent submodules: Worldview and Assigner. This separation ensures that state consensus is kept distinct from assignment decisions.

#### 3.1 **Worldview** - Distributed State Consensus
Maintains the global hall order matrix and state synchronization across all elevators.

**Responsibility:**
- Maintains the distributed hall order matrix with version tracking
- Merges incoming state updates from peer elevators using versions for conflict resolution
- Tracks cab call requests from all elevators
- Publishes the current local state to all peers periodically
- Clears orders when peers become unavailable
- Updates hall order lamps to reflect current system state

**Inputs (receive-only channels):**
- `buttonChan` - Local button press events (from ElevIO): floor and button type
- `localStateChan` - Current state of this elevator (from ElevatorController): floor, direction, behavior
- `clearedOrdersChan` - Orders this elevator just completed (from ElevatorController): list of floors cleared
- `netRx` - Network messages from peer elevators (from Networking): HallOrderMatrix, CabCalls, ElevatorState
- `peerEventChan` - Peer alive/dead events (from PeerMonitor): peer ID and status

**Outputs (send-only channels):**
- `netTx` - Broadcast state to all peers (to Networking): HallOrderMatrix, CabCalls, this elevator's state
- `lightCommandChan` - Hall lamp control commands (to ElevIO): which hall buttons to illuminate
- `worldviewChan` - Current global state snapshot (to RunAssigner): HallRequests, CabRequests, PeerStates

#### 3.2 **Assigner** - HRA Order Assignment Algorithm
Computes which orders THIS elevator should execute using deterministic HRA algorithm.

**Responsibility:**
- Receives the current global state from Worldview
- Runs the HRA algorithm to compute optimal order assignments
- Only sends new assignments when they change from previous state
- Minimizes communication overhead by filtering duplicate assignments

**Inputs (receive-only channels):**
- `worldviewChan` - Current global state snapshot (from RunWorldview): hall requests, cab requests, peer states

**Outputs (send-only channels):**
- `assignedRequestsChan` - New orders for this elevator (to ElevatorController): which floors to serve

---

### 4. **PeerMonitor** - Failure Detection
Tracks availability of peer elevator nodes.

**Responsibility:**
- Monitors heartbeat messages from other peers
- Detects dead peers via timeout
- Notifies OrderSync when peers become unavailable
- Sends heartbeats that include this elevator's peerID

**Inputs (receive-only channels):**
- `peerMonitorRx` - Heartbeat messages from other elevators (from Network)

**Outputs (send-only channels):**
- `peerEventChan` - Peer status changes: alive or dead (to OrderSync)
- `peerMonitorTx` - This elevator's heartbeat (to Network)

**Failure Detection Logic:**
- Missing heartbeat for `PEER_TIMEOUT` → Mark peer as dead
- Resuming heartbeats → Mark peer as alive again

---

### 5. **Networking** - UDP Broadcast
Handles all network communication between peers.

**Responsibility:**
- Sends broadcast UDP messages on `BROADCAST_ADDRESS`
- Receives UDP messages and routes to correct modules
- Encodes/decodes network messages
- Contains NO domain logic (no assignments, no scheduling)

**Inputs (receive-only channels):**
- `orderSyncTx` - Order state updates to broadcast (from OrderSync)
- `peerMonitorTx` - Heartbeat messages to broadcast (from PeerMonitor)

**Outputs (send-only channels):**
- `orderSyncRx` - Order state from peer elevators (to OrderSync)
- `peerMonitorRx` - Heartbeat messages from peer elevators (to PeerMonitor)

**Data Flow:**
- OrderSync sends via `orderSyncTx` → UDP broadcast to all peers
- PeerMonitor sends via `peerMonitorTx` → UDP broadcast to all peers
- All incoming UDP messages → Routed to both `orderSyncRx` and `peerMonitorRx`

**Key Design:**
- Pure transport layer: just encode/decode and broadcast
- Broadcast-only: no directed messages, all peers receive same messages
- No caching or message ordering: domain modules handle that
- Resilient to message loss: periodic rebroadcast by OrderSync and PeerMonitor
---

## Configuration

All parameters in `config/config.go`:
- `N_FLOORS`: Number of floors in the building
- `BROADCAST_PORT`: UDP port for peer communication
- `BROADCAST_ADDRESS`: UDP broadcast address
- `PEER_TIMEOUT`: How long to wait before declaring peer dead
- `MOTOR_TIMEOUT`: How long before declaring motor stuck
- ...

Run multiple elevators:
```bash
./SimElevatorServer --port 15657 &
./SimElevatorServer --port 15658 &

go run main.go -peerID 1 -serverAddr localhost:15657 &
go run main.go -peerID 2 -serverAddr localhost:15658 &
```
