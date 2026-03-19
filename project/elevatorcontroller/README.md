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

