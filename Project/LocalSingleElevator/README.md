## LocalSingleElevator

### Responsibility
- Owns and runs the **local elevator** logic for one node.
- Executes the **FSM** (Idle / DoorOpen / Moving) and local request handling.
- Converts button, floor and obstruction events into **DriverCmd** outputs (motor/lamps/indicator) via channels.
- Outputs to OrderSync via channels:
  - `ElevatorState` (for heartbeats / OrderSync)
  - `ClearedOrders` (when orders are served/cleared)
- Receives `Orders (OrderMatrix?)` from `OrderSync` via channel, describing which hall orders this elevator should serve, and also all cab orders.

### Owns (mutable state)
This module is the **only** writer of:
- `LocalSingleElevator` which holds:
  - `ElevatorState` (floor, direction, behaviour)
  - Local request matrix
- Door timer resource (timer lives in the **shell**)

No other module is allowed to mutate these structures directly, only via channels

### Run()

#### Inputs (receive-only channels)
- `buttonCh <-chan elevio.ButtonEvent`
  Button events from OrderSync
- `floorCh <-chan int`  
  Floor sensor events from ElevIO.
- `obstructionCh <-chan bool`
  Obstruction switch events from ElevIO

#### Outputs (send-only channels)
- `driverCmdch chan<- DriverCmd`  
  Commands to be executed by the ElevIO boundary driver.
- `clearedOrdersCh chan<- ClearedOrders`  
  Notifies OrderSync about which orders were cleared via channel.
- `stateOutCh chan<- ElevatorState`  
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
