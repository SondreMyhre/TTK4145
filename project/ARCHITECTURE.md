# Elevator System Architecture

Complete architectural overview of the distributed elevator control system.

---

## What Problem Does This Solve?

**Goal**: Control multiple elevators in a building such that hall orders (buttons in hallway) are efficiently assigned to elevators.

**Constraints**:
- Multiple independent elevator nodes (potentially 4+)
- No central coordinator (each node is peer)
- Network messages can be lost
- Peers can fail (network partition, crash)
- Need optimal order assignment in real-time

**Solution**: Each elevator node runs the complete system logic independently and reaches the same assignments via deterministic HRA algorithm.

---

## Core Insight: Peer-to-Peer Consensus

The key architectural insight is that **all elevators compute the same order assignments independently**:

```
All elevators know the same:
  - Hall order matrix (what orders are pending)
  - Current state of all elevators (floor, direction, behavior)
  - Set of alive peers (via heartbeats)
  
All elevators have the same algorithm:
  - HRA (Hall Request Assignment)
  
Therefore:
  - All elevators compute identical assignment: "Order at floor 3 up → Elevator 2"
  - No central authority needed
  - No network round-trips to ask "should I take this order?"
  - Survives temporary network partitions
```

This is the fundamental pattern that makes the system decentralized and resilient.

---

## System Layers

```
┌─────────────────────────────────────────────────────────────────┐
│  APPLICATION LAYER: Control Logic                              │
│                                                                 │
│  ┌─────────────────────┐          ┌──────────────────────────┐ │
│  │ ElevatorController  │          │ OrderSync (two parts)    │ │
│  │ - Local FSM         │          │ ├─ Worldview            │ │
│  │ - Request clearing  │◄────────►│ └─ Assigner (HRA)       │ │
│  │ - Motor control     │          │                          │ │
│  └─────────────────────┘          └──────────────────────────┘ │
│           ↑                                ↑                    │
└───────────┼────────────────────────────────┼───────────────────┘
            │                                │ wor     
┌───────────┼────────────────────────────────┼───────────────────┐
│  COORDINATION LAYER: State Management                          │
│           │                                │                   │
│           │          ┌──────────────────────┴───┐              │
│           │          │                          │              │
│           │     ┌─────────────────┐      ┌─────────────────┐  │
│           │     │  PeerMonitor    │      │  Networking     │  │
│           │     │  - Heartbeats   │      │  - UDP Broadcast│  │
│           │     │  - Failure det. │      │  - Encode/decode│  │
│           │     └────────┬────────┘      └────────┬────────┘  │
│           │              │                        │            │
└───────────┼──────────────┼────────────────────────┼────────────┘
            │              │                        │
        Command        Heartbeat                 Network
         (Motor,    (I'm alive +           (UDP broadcast)
         Door,       State snapshot)
         Lamps)      
            │              │                        │
            ▼              ▼                        ▼
┌───────────┬──────────────┬────────────────────────────────────┐
│  HARDWARE LAYER: I/O                                          │
│                                                              │
│          ┌──────────────────────────────────────┐           │
│          │  ElevIO                              │           │
│          │  - Simulator interface               │           │
│          │  - Poll buttons, floors, obstruction│           │
│          │  - Execute motor, door, lamps       │           │
│          │  - Hardware communication           │           │
│          └──────────────────────────────────────┘           │
│                     ↓                    ↑                    │
│                 (Events)            (Commands)               │
│                                                              │
└──────────────────────────────────────────────────────────────┘
                        ▼
                 Elevator Simulator
```

---

## Module Responsibilities

### Layer 1: Application (Domain Logic)

#### ElevatorController (Local FSM)
- **Owns**: Local elevator state, FSM behavior
- **Inputs**: Assigned orders, floor events, obstruction events
- **Outputs**: Driver commands, cleared orders, local state broadcast
- **Decision**: "Should I stop here?" "Which orders to clear?"
- **Never asks**: "Should I take this order globally?" (OrderSync decides that)

#### OrderSync (Distributed Coordination)
Two parallel submodules running as separate goroutines:

**Worldview**:
- **Owns**: Hall order matrix, cab calls, peer list
- **Inputs**: Local buttons, network messages with peer state/orders, cleared orders, peer events
- **Outputs**: Broadcasts state, feeds worldview to Assigner
- **Decision**: "What is the global state right now?"

**Assigner**:
- **Owns**: Result of HRA algorithm  
- **Inputs**: Complete worldview snapshot
- **Outputs**: Assignments for this elevator to ElevatorController
- **Decision**: "Which orders should this elevator take?" (deterministic HRA)

### Layer 2: Coordination

#### PeerMonitor (Failure Detection)
- **Owns**: Peer liveness state, heartbeat timestamps
- **Inputs**: Heartbeat messages (embedded in network)
- **Outputs**: Peer status change notifications
- **Decision**: "Is peer alive or dead based on heartbeat timeout?"

#### Networking (Transport)
- **Owns**: UDP socket configuration
- **Inputs**: Messages to broadcast from OrderSync and PeerMonitor
- **Outputs**: Received messages routed to OrderSync and PeerMonitor
- **Decision**: None! Pure transport, no domain logic

### Layer 3: Hardware Interface

#### ElevIO (Hardware I/O)
- **Owns**: Simulator connection
- **Inputs**: Motor, door, lamp commands from driver code
- **Outputs**: Button events, floor sensor events, obstruction events
- **Decision**: None! Just routes hardware in/out

---

## Information Flow - Traceability

### Scenario 1: User Presses Hall Button at Floor 3, Up

```
Timeline:
─────────────────────────────────────────────────────────────────

T1: elevio.PollButtons() detects button press on floor 3 up
    Sends: elevio.ButtonEvent{3, BtnHallUp}
    Channel: buttonChan

T2: ordersync.RunWorldview() receives on buttonChan
    Action: HallOrderMatrix[3][UP] = true
    Action: Hall lamp for 3-up lights up (tells driver)
    
T3: Within worldview update, sends on orderSyncTx
    Sends: NetMsg{HallOrderMatrix: [...], ...}
    Channel: orderSyncTx

T4: networking.Run() receives on orderSyncTx
    Action: Encodes to JSON
    Action: Sends UDP broadcast to all peers
    
T5: (Milliseconds later) ordersync.RunAssigner() receives on worldviewChan
    Input: Complete worldview with HallOrderMatrix[3][UP] = true
    Action: Calls HRA algorithm ("which elevator should take floor 3 up?")
    HRA says: Elevator 2 (could be different if distances vary)
    Action: Sends RequestMatrix for my assignments
    Channel: assignedRequestsChan (only if this elevator won)

T6: (Nearly simultaneously, other elevators)
    Elevator 1's Assigner: "Floor 3 up → Elevator 2"
    Elevator 2's Assigner: "Floor 3 up → Elevator 2"
    Elevator 3's Assigner: "Floor 3 up → Elevator 2"
    (All computed SAME decision independently!)

T7: Elevator 2 only: elevatorcontroller.Run() receives on assignedRequestsChan
    New request at floor 3, direction up
    If currently idle: transitions to Moving
    Updates state
    Sends motor command to ElevIO
    Channel: driverCommandChan

T8: elevio.RunDriver() receives on driverCommandChan
    Command: setMotorDirection(DirUp)
    Action: Tells simulator to run motor up
    Elevator starts moving toward floor 3

T9-N: Floor events, obstruction events, etc.

TN+1: Elevator 2 arrives at floor 3
     floorChan receives: 3
     
TN+2: elevatorcontroller.onFloorArrival() processes arrival at floor 3
     Checks: shouldStop? → Yes (has request at 3 up)
     Action: clearAtCurrentFloor() removes request
     Removed orders: [{floor: 3, button: BtnHallUp}]
     Sends: clearedOrdersChan with cleared orders
     
TN+3: ordersync.RunWorldview() receives on clearedOrdersChan
     Action: HallOrderMatrix[3][UP] = false
     Action: Hall lamp 3-up turns off
     Action: Broadcasts cleared order
     Channel: orderSyncTx
     
TN+4: All other elevators' OrderSync receive cleared order
     Action: Update their copy of HallOrderMatrix[3][UP] = false
     
[ORDER COMPLETE]
```

**Key observation**: Information flowed in ONE DIRECTION:
```
Button → OrderSync → HRA → ElevatorController → Motor
                      ↓
                   Broadcast
                      ↓
                  Other elevators
```

No callbacks, no back-and-forth between modules.

---

### Scenario 2: Elevator Fails (Network Partition)

```
T0:   Elevator 2 operational, sends heartbeats
      All peers healthy

T5:   Network partition occurs (Elevator 2 isolated)
      Elevator 2 unaware, keeps trying to broadcast
      Others no longer receive heartbeats from Elev2

T10:  PeerMonitor on Elev 1 detects timeout (6 seconds no heartbeat)
      Sends: PeerUpdate{Elev2, StatusDead}
      Channel: peerEventChan

T11:  OrderSync Worldview receives PeerUpdate
      Action: Update PeerList, mark Elev2 dead
      Action: Triggers reassignment
      
T12:  OrderSync Assigner receives updated worldview
      Input: Any orders assigned to Elev2 now have no responsible elevator
      Action: HRA recomputes with Elev2 removed
      New assignments: Elevator 2's orders → redistributed among Elev 1, 3, 4
      
T13:  (Simultaneously) Elevators 1, 3, 4 all independently compute:
      "Elevator 2 is dead, reassign its orders"
      All reach SAME new assignments
      
T14:  Each elevator sends updated assignments to ElevatorController
      Elevators 1, 3, 4 update their request matrices
      Elevator 2's orders start getting handled by others

T30:  Network heals, Elevator 2 sends heartbeat again

T31:  PeerMonitor on Elev 1 detects heartbeat resumed
      Sends: PeerUpdate{Elev2, StatusAlive}
      Channel: peerEventChan

T32:  OrderSync updates PeerList, Elev2 marked alive again
      HRA recomputes with Elev2 back in game
      Orders may get redistributed again

[SYSTEM RECOVERED]
```

**Resilience property**: No central authority means:
- Detection doesn't require asking a master
- Reassignment doesn't require central permission
- Each elevator independently decides what to do
- System converges to new consensus automatically

---

## Data Structures Flow

```
┌──────────────────┐
│   ButtonEvent    │ User action: which button pressed
│ {Floor, Button}  │ 
└────────┬─────────┘
         │
         ▼
┌──────────────────────────────────────┐
│   HallOrderMatrix / CabCalls         │ Global state in OrderSync
│ [Floor][Button] → OrderStatus        │
└────────┬─────────────────────────────┘
         │
         ├─→ ┌────────────────────┐
         │   │ HRA Algorithm      │
         │   │ (external C++ exe) │ Deterministic assignment
         │   └────────┬───────────┘
         │            │
         ▼            ▼
┌────────────────────────────────────────────┐
│   RequestMatrix (per elevator)             │ What THIS elevator should execute
│   [Floor][Button] → bool (assigned to me?) │
└────────┬─────────────────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│   ElevatorState              │ Current state of elevator
│ {Floor, Direction, Behavior} │
└────────┬─────────────────────┘
         │
         ├─→ ┌──────────────────────────┐
         │   │  shouldStop() decision   │ Do I stop at this floor?
         │   │  clearAtCurrentFloor()   │ What orders to clear?
         │   └────────┬─────────────────┘
         │            │
         ▼            ▼
┌────────────────────────────────────┐
│   DriverCommand                    │ What to do (motor, door, lamps)
│ {Type, Value}                      │
└────────┬──────────────────────────┘
         │
         ▼
┌────────────────────────────────────┐
│   Elevator Simulator               │ Hardware effects
└────────────────────────────────────┘
```

---

## Key Design Decisions

### 1. Peer-to-Peer, Not Client-Server
- **Why**: No single point of failure
- **Trade-off**: Harder to debug (distributed state)
- **Mitigated by**: Consistent algorithm and frequent broadcasts

### 2. Deterministic Assignment (HRA Algorithm)
- **Why**: All peers reach same conclusion without voting
- **How**: All have same input (worldview) + same algorithm
- **Benefit**: Works despite message loss and reordering

### 3. Separate Worldview from Assignment
- **Why**: Clean separation of concerns
- **Worldview**: Facts (what orders exist)
- **Assigner**: Decisions (who takes what)
- **Benefit**: Easy to swap different assignment algorithms

### 4. State Communicated via Snapshots, Not Deltas
- **Sent**: ElevatorState snapshot (not "moved up one floor")
- **Why**: Robust to lost messages
- **Lost message**: Next snapshot will have correct state anyway

### 5. Broadcast-Only Network, No Directed Messages
- **Why**: Simple, no addressing/routing needed
- **Trade-off**: All peers see all messages (even ones not for them)
- **Benefit**: Resilient to topology changes, works with firewalls

### 6. Pure Functions for Decisions
- **Why**: Testable, no side effects, easy to reason about
- **How**: Decision functions input state, return actions (effects)
- **Benefit**: Can be simulated without running full system

---

## Failure Modes and Recovery

### Scenario: Message Lost
```
T0: OrderSync broadcasts "Order at floor 5"
T1: MESSAGE LOST in network
T2: Other elevators don't know about floor 5 order
T3-5: ...waiting...
T6: OrderSync periodically rebroadcasts (NETMSG_TICK_INTERVAL)
    "Order at floor 5" broadcast again
T7: Other elevators receive, now they know
    [RECOVERED - No human intervention needed]
```

### Scenario: Peer Crashes
```
T0: Elevator 2 running, sends heartbeats
T5: Elevator 2 crashes (power loss, software fault)
T10: PeerMonitor timeout reached (6 seconds no heartbeat)
    All peers independently declare Elevation 2 dead
    All recompute assignments without Elevator 2
    (System continues with 3 elevators instead of 4)
T6000: Someone restarts Elevator 2
    Elevator 2 sends heartbeat
    All peers see it, mark it alive
    All recompute assignments including Elevator 2
    [RECOVERED - Order gets reassigned properly]
```

### Scenario: Network Partition
```
T0:   [Elev 1] ═══ NETWORK ═══ [Elev 2, Elev 3]
      Both groups operational but can't communicate

T5:   Each group computes assignments for orders
      Group A (Elev 1): "Order 5 → I'll take it"
      Group B (Elev 2): "Order 5 → I'll take it"
      BUG: Order 5 might be started by both!

BUT: This scenario's harm is limited because:
- Elevators can't physically go to same floor simultaneously
- One will reach floor 5, serve it, clear in their group
- When network heals, cleared status propagates
- Other group sees order is cleared, stops trying to serve it
- "At least one of you will try" beats "nobody tries"
```

---

## Tracing a Request - The Complete Path

**Question**: How does a button press become an elevator moving?

**Answer**: Follow the channels:

1. **Hardware reads button** → `buttonChan`
2. **OrderSync processes** → Updates HallOrderMatrix
3. **OrderSync broadcasts** → `orderSyncTx`
4. **Network delivers** → All peers' `orderSyncRx`
5. **Assigner processes** → Runs HRA
6. **Assigner outputs** → `assignedRequestsChan`
7. **ElevatorController receives** → Updates its request matrix
8. **FSM transitions** → Changes behaviors, direction
9. **Controller outputs** → `driverCommandChan`
10. **ElevIO driver executes** → Motor, door, lamp commands
11. **Simulator responds** → Elevator moves
12. **Floor sensor fires** → `floorChan`
13. **Controller processes arrival** → Clears order, transitions state
14. **Controller outputs cleared** → `clearedOrdersChan`
15. **OrderSync receives cleared** → Updates HallOrderMatrix
16. **OrderSync broadcasts** → All peers see order cleared
17. **All halt attempts to serve it** → State converged

Every step flows in one direction (mostly). No callbacks or backtracks needed (normally).

---

## Configuration Points

From `config/config.go` (system architect controls):
- `N_FLOORS`: Building size
- `BROADCAST_ADDR`, `BROADCAST_PORT`: Network topology
- `PEER_TIMEOUT`: Failure detection sensitivity
- `*_TICK_INTERVAL`: Responsiveness vs. load

From command line (operator controls):
- `-peerID`: Identity of this elevator
- `-serverAddr`: Which simulator instance

Design goal: Changing these shouldn't require code changes.

---

## Module Independence

Each module can be understood and tested in isolation:

| Module | Depends On | Is Used By |
|--------|-----------|-----------|
| ElevIO | Hardware (simulator) | ElevatorController |
| ElevatorController | ElevIO (hardware outputs) | OrderSync (reads state) |
| Networking | Network (UDP) | OrderSync, PeerMonitor |
| PeerMonitor | Networking (hearbeat input) | OrderSync |
| OrderSync WorldView | ElevIO, ElevatorController, Networking, PeerMonitor | OrderSync Assigner |
| OrderSync Assigner | Worldview | ElevatorController |

**Key property**: Each module exposes a simple interface (channels and types). Internals are hidden.

---

## Testing Strategy

| Component | Test Type | Notes |
|-----------|-----------|-------|
| ElevIO | HW integration | Requires simulator |
| Pure decision functions | Unit tests | `shouldStop`, `clearAtCurrentFloor`, HRA input |
| ElevatorController FSM | Isolated with mocks | Mock channels for state/commands |
| OrderSync | State machine simulation | Mock network, verify assignments |
| Networking | Encode/decode unit tests | No network needed |
| PeerMonitor | Timeout simulation | Mock heartbeat streams |
| Full system | Integration test | Multiple processes, real network |

**Insight**: Pure functions don't need network running. FSM logic can run completely offline.

---

## Performance Considerations

**Responsiveness**:
- Button press → Assignment: ~100ms (NETMSG_TICK_INTERVAL)
- Assignment → Motor: ~10ms (direct channel)
- Floor arrival → Clear: ~10ms (FSM reaction time)

**Throughput**:
- ~100 messages/second total (very low)
- Broadcast to 4 peers = 400 UDP packets/second
- All fit within 1Mbps link with room to spare

**Scalability**:
- More elevators = more network traffic (linear)
- More floors = larger HRA computation (polynomial)
- System designed for ~4-8 elevators, ~4 floors

---

## Debugging Tips

**To trace information flow**:
1. Look at main.go for channel topology
2. Find where channel is written: that's the source
3. Find where channel is read: that's the destination
4. Add logging at those points

**To understand FSM logic**:
1. Read `fsm_core.go` - event handlers
2. Each handler: input state + event → output state + effects
3. Understand `chooseDirection` and `shouldStop` first

**To debug message loss**:
1. Check `networking.Run()` - is broadcast working?
2. Check peer listening - are sockets on right port?
3. Verify firewall/broadcast address config

**To debug failed assignment**:
1. Check HRA executable - is it installed?
2. Check HRA input - is worldview being computed correctly?
3. Add logging in Assigner to see HRA input/output

---

## Summary: The Big Picture

A distributed system where:
- **Peers communicate via broadcast** - not client/server
- **All peers compute same assignments** - consensus without voting
- **Computation is deterministic** - same inputs → same outputs
- **Failures are tolerated** - system recovers via rediscovery
- **Information flows in one direction** - button → assignment → execution
- **Pure functions make it testable** - don't need network running
- **Clear module boundaries** - each module is independently understandable

This design trades some complexity (distributed consensus) for huge benefits (no single point of failure, resilient to message loss and delays).
