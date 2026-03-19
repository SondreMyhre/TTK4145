## OrderSync - Distributed Order State & Assignment

### Overall Responsibility

Maintains the **global distributed state** of all orders in the system and computes optimal order assignments using the HRA (Hall Request Assignment) algorithm.

**Systems**: Two main submodules run in parallel (via two separate goroutines):

1. **Worldview** - Maintains distributed consensus on order state
2. **Assigner** - Runs HRA algorithm to compute order assignments

**Key design principle**: OrderSync does NOT execute orders. It assigns which orders THIS elevator should execute, and ElevatorController carries them out.

---

### Module Coherence - Two Concerns, Two Goroutines

The module cleanly separates two distinct concerns:

```
RunWorldview()     →  Maintains global order state
    ↓
worldviewChan  →  (communication between Worldview and Assigner)
    ↓
RunAssigner()  →  Computes assignments and sends to ElevatorController
```

This separation ensures:
- Worldview focuses on consensus (collecting state from peers, broadcasting changes)
- Assigner focuses on optimization (runs HRA algorithm deterministically)
- Clean interface: Assigner depends only on worldviewChan, not network details

---

### Worldview Submodule - The Consensus Component

**Responsibility**: Maintain the global "worldview" - what we know about all current orders and elevator states.

**State ownership - Worldview maintains:**
- `HallOrderMatrix` - Status of each hall button (pending/confirmed/inactive)
- `CabCalls` - Cab request state for each elevator
- `PeerStates` - Current state (floor, direction, behavior) of all known peers
- `PeerList` - Which peers are alive vs dead

**Information sources (inputs):**
```
Local buttons (in hall/cab)
    ↓ buttonChan
    ├→ Update hall/cab state
    └→ Broadcast to peers
    
Other elevators' states
    ├ orderSyncRx (from network)
    └→ Merge into worldview
    
Other elevators' cleared orders
    ├ orderSyncRx (from network)
    └→ Remove from HallOrderMatrix
    
Local elevator state  
    ├ localStateChan (from ElevatorController)
    └→ Broadcast to peers
    
Local cleared orders
    ├ clearedOrdersChan (from ElevatorController)
    └→ Update HallOrderMatrix, broadcast
    
Peer liveness changes
    ├ peerEventChan (from PeerMonitor)
    └→ Update PeerList, trigger reassignment
```

**Key operation: Building the Worldview**:
```go
type WorldviewMsg struct {
  HallRequests  [N_FLOORS][N_HALL]bool              // Active orders in hallway
  CabRequests   map[ElevID][N_FLOORS]bool          // Cab calls for each elevator
  PeerStates    map[ElevID]ElevatorState           // Current state of each peer
  Peers         []Peer                              // Peer list + status
}
```

When Worldview sends worldviewChan, it's sending a complete snapshot of what we know about the system. The Assigner reads this and computes assignments.

**State consistency property**: All nodes should have the SAME worldview (or nearly the same) because:
- All nodes receive the same network messages (via broadcast)
- Nodes process messages in potentially different order, BUT
- HRA algorithm produces deterministic output for identical worlds
- Small timing differences don't break the system (assigner runs periodically)

---

### Assigner Submodule - The Optimization Component

**Responsibility**: Read the worldview and decide which elevator should handle each order.

**How it works**:
1. Receive `WorldviewMsg` from Worldview
2. Call HRA algorithm with current worldview
3. HRA returns: for each elevator, which orders it should take
4. Extract "my" assignments (this elevator's ID)
5. Send via `assignedRequestsChan` to ElevatorController

**The HRA algorithm** (not implemented here, external executable):
- Input: All hall requests + all elevator states
- Output: For each elevator, which orders to take
- Properties: 
  - Deterministic: same input → same output
  - Optimal: minimizes sum of travel times
  - Fair: no elevator gets stuck with too many orders
  - Symmetric: all nodes compute independently, same result

**Why external executable?**:
- HRA is complex C++ code provided by course
- Go calls it via `callHRA(input)` which marshals JSON to subprocess
- Safe: isolated from main elevator logic
- Easy to replace: could swap in different algorithm

---

### Information Flow - How Orders Get to Controllers

**Flow of a button press:**
```
User presses button at floor 3 up
    ↓
elevio.PollButtons detects
    ↓ buttonChan
ordersync.Worldview receives
    ↓
HallOrderMatrix[3][UP] = true
    ↓ orderSyncTx
Broadcast to all peers
    ↓
(Other elevators receive on orderSyncRx)
    ↓
All Worldviews updated
    ↓
Assigner (all in parallel) receives worldviewChan
    ↓  (all run same HRA algorithm on same inputs)
Assigner 1: floor 3 UP → assign to elevator 2
Assigner 2: floor 3 UP → assign to elevator 2       (SAME DECISION MADE INDEPENDENTLY)
Assigner 3: floor 3 UP → assign to elevator 2
    ↓
Elevator 2's assignedRequestsChan gets updated
    ↓ assignedRequestsChan
ElevatorController receives: "serve floor 3 up"
    ↓
Elevator 2 starts moving toward floor 3
    ↓ (when it arrives)
ElevatorController clears the order
    ↓ clearedOrdersChan
OrderSync receives cleared order
    ↓
HallOrderMatrix[3][UP] = false
    ↓ orderSyncTx
Broadcast to all peers
    ↓
All worldviews updated
```

**Critical insight**: Each elevator independently computes the same assignments because all inputs are the same. This is consensus without a central coordinator.

---

### State Management - Distributed Consensus

**Hard problem**: Maintaining consistency across distributed elevator nodes when messages can be lost or delayed.

**Solution used**: 
- Broadcast all state changes frequently
- Nodes eventually agree even if messages are lost
- System is "eventually consistent" not "strongly consistent"
- Good enough because:
  - Lost messages are retransmitted periodically
  - Old versions of orders don't break system
  - Worst case: temporary re-assignment of already-handled orders

**State ownership clarity**:
```
Worldview owns (updates):
  - HallOrderMatrix (from buttons + network messages about clears)
  - CabRequests (from buttons + network)
  - PeerList (from heartbeats via PeerMonitor)

Worldview reads but does NOT own:
  - Individual elevator positions (read from network, not persisted)
  - Button events from other elevators (used momentarily to update hall matrix)

Assigner owns (outputs):
  - Assignment decisions (deterministic function of worldview)

Assigner does NOT modify:
  - Worldview (read-only access)
  - Any other external state
```

---

### Design Principles

**1. Separation of concerns**: Worldview ≠ Assigner
- Worldview: consensus on facts
- Assigner: allocation of work
- Clean interface: worldviewChan

**2. Deterministic assignment**: All elevators compute independently and reach same conclusion
- No master/slave
- No voting needed
- Works despite network delays

**3. Localized decisions**: Each elevator only cares about its own assignments
- ElevatorController doesn't know about global state
- Only processes "my orders"
- Can't be overridden by other elevators

**4. Eventual consistency**: System recovers from temporary disagreements
- Old orders will be broadcast again if assignments change
- Periodic reassignment when topology changes
- Resilient to message loss

---

### Inputs (receive-only channels)

- `buttonChan <-chan elevio.ButtonEvent` (from ElevIO)
  - Hall buttons: share globally
  - Cab buttons: accumulate for my elevator

- `localStateChan <-chan ElevatorState` (from ElevatorController)
  - My current elevator state
  - Broadcast to peers for assignment logic

- `clearedOrdersChan <-chan []Order` (from ElevatorController)
  - Orders just completed
  - Remove from HallOrderMatrix, broadcast

- `orderSyncRx <-chan NetMsg` (from Network)
  - Order/state updates from peer elevators
  - Merge into worldview

- `peerEventChan <-chan []PeerUpdate` (from PeerMonitor)
  - Peer went down / came back up
  - Update PeerList, reconsider assignments

### Outputs (send-only channels)

- `assignedRequestsChan chan<- RequestMatrix` (to ElevatorController)
  - Orders this elevator should execute
  - Derived from HRA algorithm

- `orderSyncTx chan<- NetMsg` (to Network)
  - Broadcast my current state
  - Send cleared orders to peers

- `driverCommandChan chan<- DriverCommand` (to ElevIO)
  - Hall lamp commands (light/unlight buttons)
  - Indicates which orders are pending

---

### Completeness - What Does OrderSync Handle?

✓ Receives hall orders from any elevator
✓ Broadcasts orders to all peers
✓ Merges incoming orders from peers
✓ Removes orders when they're cleared
✓ Assigns orders to best elevator
✓ Handles peer failures (via PeerMonitor)
✓ Re-assigns orders when peers fail
✓ Controls hall lamps (on/off based on pending orders)

✗ Does NOT execute orders (ElevatorController does)
✗ Does NOT communicate with hardware directly
✗ Does NOT compute elevator movements (that's the FSM)
