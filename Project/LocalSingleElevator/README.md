## LocalSingleElevator

### Responsibility
- Owns and runs the **local elevator** logic for one node.
- Executes the **FSM** (Idle / DoorOpen / Moving) and local request handling.
- Converts local events into **DriverCmd** outputs (motor/lamps/indicator) via channels.
- Outputs to OrderSync via channels:
  - `ElevatorState` (for heartbeats / OrderSync)
  - `ClearedOrders` (when orders are served/cleared)
- Receives `HallAssignment (OrderMatrix?)` from `OrderSync` via channel, describing which hall orders this elevator should serve.

### Owns (mutable state)
This module is the **only** writer of:
- `ElevatorState` (floor, direction, behaviour)
<!-- - Cab orders: `cabOrders [N_FLOORS]bool` SKAL HÅNDTERES AV ORDERSYNC -->
- Local request matrix
- Door timer resource (timer lives in the **shell**)???

No other module is allowed to mutate these structures directly, only via channels

### Run() interface

#### Inputs (receive-only channels)
- `CabButton <-chan CabOrder`  
  Cab button presses routed from OrderSync.
- `AssignedHall <-chan HallAssignment`  
  Hall orders assigned to this elevator by OrderSync.
- `Floor <-chan int`  
  Floor sensor events from ElevIO.
- `Obstruction <-chan bool`
  Obstruction switch events.
- `DoorTimeout <-chan struct{}` *(optional)*  
  If door timer is hosted outside this module.

#### Outputs (send-only channels)
- `DriverCmd chan<- DriverCmd`  
  Commands to be executed by the ElevIO boundary driver.
- `Cleared chan<- ClearedOrders`  
  Notifies OrderSync about which orders were cleared via channel.
- `StateOut chan<- ElevatorState`  
  Used by OrderSync to build heartbeats and compute assignments.

### Functional core vs Imperative shell

#### Functional core (testable)
- Pure-ish transition function(s):
  - `Step(state, requests, event) -> (newState, newRequests, effects)` (effects is commands and other output)
- No IO, no `elevio.*`, no `timer.C`, no `time.Now()`.

Effects are value objects such as:
- `SetMotor(dir)`
- `SetDoorLamp(on)`
- `SetFloorIndicator(floor)`
- `SetButtonLamp(floor, btn, on)`
- `StartDoorTimer(duration)`
- `PublishState(state)`
- `EmitCleared(clearedOrders)`

#### Imperative shell (Run loop)
- `for { select { ... } }` event loop
- Owns door timer + converts timer expiry into `DoorTimeout` event
- Applies effects by sending:
  - DriverCmd to ElevIO
  - ClearedOrders + State to OrderSync
