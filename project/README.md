# TTK4145 Distributed Elevator Control System

A highly distributed elevator control system built in Go that implements peer-to-peer coordination for multi-elevator environments.

## System Architecture

This elevator system is designed as a **peer-to-peer distributed system** where each elevator node operates independently but coordinates via broadcast messages. The system uses the **Hall Request Assignment (HRA) algorithm** to optimally distribute orders among elevators.

### Design Pattern: Peer-to-Peer Coordination
- **No central authority**: Each node runs the same algorithm and makes independent decisions
- **Consensus-based**: Order assignment is deterministic and the same on all nodes
- **Fault-tolerant**: System continues if peers fail; others detect via timeouts

## System Modules

### 1. **ElevIO** - Hardware Interface
Provides the low-level interface to the elevator simulator hardware.

**Responsibility:**
- Polls hardware sensors (floor sensor, buttons, obstruction switch)
- Executes motor commands and lamp controls
- No business logic; purely a wrapper around hardware

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
- Broadcasts state changes to all peers via UDP
- Receives state updates from other elevators
- Detects when peers become unavailable (via PeerMonitor)
- Triggers order reassignment when topology changes
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
- Detects peer failures via timeout
- Notifies OrderSync when peers become unavailable
- Sends heartbeats that include this elevator's state

**Inputs (receive-only channels):**
- `peerMonitorRx` - Heartbeat messages from other elevators (from Network)

**Outputs (send-only channels):**
- `peerEventChan` - Peer status changes: alive or dead (to OrderSync)
- `peerMonitorTx` - This elevator's heartbeat (to Network)

**Failure Detection Logic:**
- Missing heartbeat for `PEER_TIMEOUT` → Mark peer as dead
- Resuming heartbeats → Mark peer as alive again

---

### 5. **Networking** - UDP Broadcast Transport
Handles all network communication between peers.

**Responsibility:**
- Sends broadcast UDP messages on `BROADCAST_ADDRESS`
- Receives UDP messages and routes to correct module
- Encodes/decodes network messages
- Has NO business logic (no assignments, no scheduling)

**Data Flow:**
- OrderSync sends via `orderSyncTx` → Broadcast to all peers
- PeerMonitor sends via `peerMonitorTx` → Broadcast to all peers
- All incoming messages → Route to `orderSyncRx` and `peerMonitorRx`

**Key Design Decisions:**
- Broadcast addresses all nodes simultaneously
- Messages can be lost; modules handle retransmission via timers
- No reliability layer; ordered delivery not guaranteed

---

## System Startup and Dependencies

The modules must be initialized in this order (see `main.go`):

1. **Hardware Polling** - Start sensor polling routines immediately
2. **Network Transport** - Must be ready before other modules try to broadcast
3. **PeerMonitor** - Listens to network for heartbeats
4. **OrderSync/Worldview** - Maintains distributed state
5. **OrderSync/Assigner** - Receives worldview and computes assignments
6. **ElevatorController** - Listens for assignments and controls hardware
7. **Driver** - Executes commands

This order ensures that when a module tries to send/receive, the channel it depends on is ready.

---

## Information Flow During Operation

### Scenario: User presses hall button at floor 3
```
F(t=0):    elevio.PollButtons() detects button
F(t=1):    buttonChan receives event
F(t=2):    ordersync.RunWorldview() processes it
F(t=3):    HallOrderMatrix[3][UP] = true
F(t=4):    orderSyncTx sends state to network
F(t=5):    networking.Run() broadcasts to all peers
F(t=6):    Other elevators receive on orderSyncRx
F(t=7):    ordersync.RunAssigner() runs HRA algorithm
F(t=8):    assignedRequestsChan updated with my assignments
F(t=9):    elevatorcontroller.Run() changes its motor/behavior
F(t=10):   Elevator starts moving toward floor 3
...
F(t=N):    elevatorcontroller arrives at floor 3
F(t=N+1): clearAtCurrentFloor() removes order from my request matrix
F(t=N+2): clearedOrdersChan notifies ordersync
F(t=N+3): ordersync broadcasts cleared order
F(t=N+4): All peers remove order from their HallOrderMatrix
```

**Key Property**: All peers independently compute the exact same assignments because:
- All peers receive the same world state (HallOrderMatrix + PeerStates)
- HRA algorithm is deterministic
- No central coordinator needed

---

## Configuration

All parameters in `config/config.go`:
- `N_FLOORS`: Number of floors in the building
- `BROADCAST_PORT`: UDP port for peer communication
- `BROADCAST_ADDRESS`: UDP broadcast address
- `PEER_TIMEOUT`: How long to wait before declaring peer dead
- `MOTOR_TIMEOUT`: How long before declaring motor stuck

Run multiple elevators:
```bash
./SimElevatorServer --port 15657 &
./SimElevatorServer --port 15658 &

go run main.go -peerID 1 -serverAddr localhost:15657 &
go run main.go -peerID 2 -serverAddr localhost:15658 &
```

---

## Key Design Decisions

1. **Consensus-based assignment**: Every node independently computes the same assignments
   - No central bottleneck
   - Works if network partitions occur
   - Robust to message loss

2. **Separate assignment from execution**: OrderSync computes, ElevatorController executes
   - Clean separation of concerns
   - Controller only knows about its assigned orders
   - Easy to test each component independently

3. **Peer-to-peer with broadcast**: No master/slave relationship
   - Symmetric - any elevator can serve any order
   - Resilient - doesn't fail if one elevator dies
   - Scales to many elevators

4. **Lean network protocol**: Only essential state is broadcast
   - Reduce bandwidth and latency
   - Easier to detect message loss
   - Less computation at each node
