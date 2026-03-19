## ElevatorController - Local Elevator FSM

### Overall Responsibility

Runs the **finite state machine (FSM)** for a single elevator and handles all local control logic.

**Key principle**: This module is **purely reactive** - it does not make decisions about the global system. It only executes orders that have been assigned to it by OrderSync.

**Single responsibility**: Control ONE elevator based on assigned orders.

---

### How It Works

The elevator cycles through three behavioral states:

**Idle**: No requests to handle
- Motor stopped
- Waiting for assigned orders
- When a new order arrives, calls `chooseDirection()` to decide where to go

**Moving**: Traveling toward target floor(s)
- Motor running in chosen direction
- Monitors floor sensor for arrival
- When arriving at floor, checks `shouldStop()`
- If should stop: transitions to DoorOpen
- If should not stop: continues moving

**DoorOpen**: At target floor, doors open
- Door lamp on, motor off
- Clears all orders at this floor via `clearAtCurrentFloor()`
- Waits for door timer timeout (or obstruction clears)
- Calls `chooseDirection()` to decide next behavior
- Returns to Idle, Moving, or stays DoorOpen

---

### State Management - Local Only

This module **owns and maintains** all state related to the local elevator:

```
elevator struct {
  state    ElevatorState           ← ALL state is local, never shared/mutated externally
  requests [N_FLOORS][N_BUTTONS]bool  ← Only this module modifies these
}

ElevatorState {
  Floor       int                  ← Current floor (readonly from floorChan)
  Direction   Direction            ← Chosen direction (set by FSM)
  Behaviour   ElevatorBehaviour    ← Current state: Idle/Moving/DoorOpen
  Obstructed  bool                 ← Door obstruction status
  MotorStuck  bool                 ← Motor timeout flag
}
```

**Critical design**: State is NOT shared. Parameters and return values are the only way to pass data.

**State ownership clarity**:
- Only `elevatorcontroller.Run()` owns the elevator struct
- Other modules cannot read/write state directly
- Changes communicated via channels only
- Local state published on `localStateChan` for others to read (immutably)

---

### Information Flow - Input to Output

**INPUTS (receive-only channels):**
1. `assignedRequestsChan` - New request matrix from OrderSync
   - Called when: OrderSync gets new assignment from HRA
   - Updates: `elevator.requests` with orders this elevator should take
   
2. `floorChan` - Floor sensor event from ElevIO
   - Called when: Elevator passes a floor
   - Updates: `elevator.state.Floor`
   - Triggers: `onFloorArrival()` which may clear orders and change state
   
3. `obstructionChan` - Obstruction switch from ElevIO
   - Called when: Door obstruction detected/cleared
   - Resets: Door timer if obstruction while door is open

**OUTPUTS (send-only channels):**
1. `driverCommandChan` - Commands to execute (to ElevIO)
   - Motor direction (Up/Down/Stop)
   - Door lamp (on/off)
   - Floor indicator (lighting the correct floor lamp)
   
2. `clearedOrdersChan` - Orders completed at current floor
   - When: After `clearAtCurrentFloor()` removes orders
   - Content: List of `Order{Floor, Button}` that were cleared
   - Purpose: Notify OrderSync to broadcast cleared orders to all peers
   
3. `localStateChan` - Current state of this elevator
   - When: State changes (behavior, direction, floor)
   - Content: `ElevatorState` snapshot
   - Purpose: For OrderSync to broadcast and use in HRA algorithm

---

### Core FSM Functions - Pure Functions

All decision functions are **pure**: they depend only on inputs, have no side effects, and produce deterministic outputs.

```go
// Pure helper functions for FSM logic
func shouldStop(elevator elevator) bool
  Input:  Current state and requests
  Output: true if elevator should stop at current floor
  Logic:  Checks direction + requests to stop
  
func clearAtCurrentFloor(elevator elevator) (elevator, []Order)
  Input:  Current state and request matrix
  Output: Updated elevator + list of cleared orders
  Logic:  Determines which requests to clear and returns both
  
func chooseDirection(elevator elevator) directionBehaviourPair
  Input:  Current state and requests
  Output: Next direction and behavior
  Logic:  Searches up/down for requests relative to direction
  
func requestsAbove/Below/Here(elevator elevator) bool
  Input:  Elevator state and requests
  Output: true if requests exist in that direction/floor
  
// Event handlers - transform state based on events
func onFloorArrival(elevator, newFloor) (elevator, []effect)
  Description: Elevator arrived at a floor
  
func onNewRequestMatrix(elevator, requests) (elevator, []effect)
  Description: New orders assigned to this elevator
  
func onDoorTimeout(elevator) (elevator, []effect)
  Description: Door timer expired
  
func onObstruction(elevator, obstructed) (elevator, []effect)
  Description: Obstruction sensor changed
```

Each function:
- Takes elevator state as input
- Returns updated elevator state
- Produces a list of side effects (things to do)
- Does NOT modify any external state
- Is easily testable with different input scenarios

---

### Side Effects - The Effects System

Pure functions describe WHAT should happen via an `effect` list. The main loop EXECUTES these effects:

```go
type effect struct {
  kind  effectType  // setMotorDirection, setDoorOpenLamp, etc.
  value any         // The value to set (e.g., DirUp)
}

// Examples of effects:
effect{kind: setMotorDirection, value: DirUp}
    → Tells ElevIO to run motor upward
    
effect{kind: setDoorOpenLamp, value: true}
    → Tells ElevIO to light the door lamp
    
effect{kind: publishLocalState, value: elevator.state}
    → Tells system to broadcast my state
```

**Separation of concerns**: Decision logic (pure functions) is separate from action logic (effect execution).

---

### Entry Point: The Main Run Loop

```go
func Run(
  assignedRequests <-chan RequestMatrix,
  floors <-chan int,
  obstructions <-chan bool,
  commands chan<- DriverCommand,
  cleared chan<- []Order,
  state chan<- ElevatorState,
) {
  /// Elevator holds its state - local, never leaves
  elevator := makeUninitializedElevator()
  
  for {
    select {
    case newRequests := <-assignedRequests:
      elevator, effects := onNewRequestMatrix(elevator, newRequests)
      executeEffects(effects)
      
    case newFloor := <-floors:
      elevator, effects := onFloorArrival(elevator, newFloor)
      executeEffects(effects)
      
    case obstructed := <-obstructions:
      elevator, effects := onObstruction(elevator, obstructed)
      executeEffects(effects)
      
    case <-doorTimer:
      elevator, effects := onDoorTimeout(elevator)
      executeEffects(effects)
    }
  }
}
```

**Control flow**:
1. Wait for event on one of 4 input channels
2. Call pure handler function with elevator state + event
3. Receive updated state + list of effects to execute
4. Execute effects by sending on output channels
5. Loop back to waiting for next event

**This design ensures**:
- State consistency: each event processed atomically
- Testability: pure functions easy to test
- Clarity: effects explicitly describe all actions
- No side effects in decision logic

---

### Real-Time Constraints

**Not a hard real-time system**, but optimizes for responsiveness:
- Channel buffer sizes chosen to prevent blocking
- Floor events handled with priority (small buffer to force handling)
- DoorTimer ensures doors don't stay open indefinitely
- Motor timeout detects mechanical failures

---

### Inputs: Receives Only Channels (receive-only, never sends)
- `assignedRequestsChan <-chan RequestMatrix`
- `floorChan <-chan int`
- `obstructionChan <-chan bool`

### Outputs: Send-Only Channels (sends only, never receives)
- `driverCommandChan chan<- DriverCommand`
- `clearedOrdersChan chan<- []Order`
- `localStateChan chan<- ElevatorState`

---

### Design Principles Summary

1. **Coherence**: Single responsibility - execute assigned orders, nothing else
2. **Completeness**: Handles all elevator behaviors (Idle/Moving/DoorOpen)
3. **State locality**: All data stays local; communicate via immutable messages
4. **Functional purity**: Decision logic has no side effects
5. **Understandability**: Clear FSM structure; easy to trace execution
6. **Naming precision**: Function names describe WHEN they're called
7. **Directional flow**: Input → Process → Output (no callbacks back to input sources)