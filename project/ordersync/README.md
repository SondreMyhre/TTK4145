## OrderSync - Distributed Order State & Assignment

### Overall Responsibility

Maintains the **global distributed state** of all orders in the system and computes optimal order assignments using the HRA (Hall Request Assignment) algorithm.

**Systems**: Two main submodules run in parallel (via two separate goroutines):

1. **Worldview** - Maintains distributed consensus on order state
2. **Assigner** - Runs HRA algorithm to compute order assignments

**Key design principle**: OrderSync does NOT execute orders. It assigns which orders THIS elevator should execute, and ElevatorController carries them out.

---

### Worldview Submodule - The Consensus Component

**Responsibility**: Maintain the global "worldview" - what we know about all current orders and elevator states.

**State ownership - Worldview maintains:**
- `HallOrderMatrix` - Status of each hall button (pending/confirmed/inactive)
- `CabCalls` - Cab request state for each elevator
- `PeerStates` - Current state (floor, direction, behavior) of all known peers
- `PeerList` - Which peers are alive vs dead

**Key operation: Building the Worldview**:
```go
type WorldviewMsg struct {
  HallRequests  [N_FLOORS][N_HALL]bool              // Active orders in hallway
  CabRequests   map[ElevID][N_FLOORS]bool          // Cab calls for all elevators
  PeerStates    map[ElevID]ElevatorState           // Current state of each peer
  Peers         []Peer                              // Peer list + status
}
```

When Worldview sends on worldviewChan to Assigner, it's sending a complete snapshot of what we know about the system. The Assigner reads this and computes assignments.

---

### Assigner Submodule - The Optimization Component

**Responsibility**: Read the worldview and decide which elevator should handle each order.

**How it works**:
1. Receive `WorldviewMsg` from Worldview
2. Call HRA algorithm with current worldview
3. HRA returns: for each elevator, which orders it should take
4. Extract "my" assignments (this elevator's ID)
5. Send via `assignedRequestsChan` to ElevatorController

**The HRA algorithm** (not implemented here, handout executable):
- Input: All hall requests + all elevator states
- Output: For each elevator, which orders to take
- Properties: 
  - Deterministic: same input → same output
  - Optimal: minimizes sum of travel times
  - Fair: no elevator gets stuck with too many orders
  - Symmetric: all nodes compute independently, same result

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

### Configuration

From `config/config.go`:
- `NETMSG_TICK_INTERVAL`: How often OrderSync broadcasts (e.g., 100ms)
