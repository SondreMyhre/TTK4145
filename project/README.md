# TTK4145 Distributed Elevator Control System

A distributed elevator control system built in Go that implements peer-to-peer coordination for multiple elevators.

## System Architecture

This elevator system is designed as a **peer-to-peer distributed system** where each elevator node operates independently but coordinates via UDP broadcast messages. The system uses the **Hall Request Assignment (HRA) algorithm** to optimally distribute orders among elevators.

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

**Key Decision Logic:**
- `shouldStop()` - Determines if elevator should stop at current floor
- `clearAtCurrentFloor()` - Determines which orders to clear at current floor
- `chooseDirection()` - Decides next direction when idle

---

### 3. **OrderSync** - Distributed Order Coordination
Maintains global order state and runs order assignment algorithm.

**Responsibility:**
- Maintains the global hall order matrix (which hall orders are pending)
- Maintains cab call state (each elevator's local requests)
- Receives hall button presses and updates global state
- Broadcasts state changes to all peers (via networking-module)
- Receives state updates from other elevators
- Detects when peers become unavailable (via PeerMonitor)
- Controls hall lamps (lights in hallway)

**State ownership:**
- `HallOrderMatrix` - Status of all hall orders globally
- `CabCalls` - Cab requests for all elevators
- `PeerList` - Known elevator peers and their status

**Key Submodules:**
- `RunWorldview()` - Maintains distributed state consensus
- `RunAssigner()` - Runs HRA algorithm to compute assignments

**Inputs (receive-only channels):**
- `buttonChan` - Local button presses (from ElevIO)
- `localStateChan` - My elevator's current state (from ElevatorController)
- `clearedOrdersChan` - Orders I just completed (from ElevatorController)
- `orderSyncRx` - Order state from other elevators (from Network)
- `peerEventChan` - Peer up/down events (from PeerMonitor)

**Outputs (send-only channels):**
- `assignedRequestsChan` - Orders assigned to this elevator (to ElevatorController)
- `orderSyncTx` - Broadcast of my state (to Network)
- `driverCommandChan` - Hall lamp commands (to ElevIO)

**Information Flow Example - Button Press:**
```
User presses button → ElevIO detects → buttonChan → OrderSync receives
  → HallOrderMatrix updated → Broadcast via orderSyncTx
  → Network sends to all peers → Other elevators' OrderSync receives
  → All run HRA algorithm independently → Same assignment computed
  → My elevator receives assignedRequestsChan → ElevatorController schedules
  → ElevatorController arrives at floor → clearedOrdersChan
  → OrderSync notifies other peers
```

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
