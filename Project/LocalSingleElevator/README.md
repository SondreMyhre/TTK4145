## LocalSingleElevator

### Responsibility
- Owns and runs the **local elevator** logic for one node.
- Executes the **FSM** (Idle / DoorOpen / Moving) and local request handling.
- Converts button, floor and obstruction events into **DriverCommand** outputs (motor/lamps/indicator) via channels.
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
- `localOrderChan <-chan elevio.ButtonEvent`
  Button events from OrderSync
- `floorChan <-chan int`  
  Floor sensor events from ElevIO.
- `obstructionChan <-chan bool`
  Obstruction switch events from ElevIO

#### Outputs (send-only channels)
- `driverCommandChan chan<- DriverCommand`  
  Commands to be executed by the ElevIO boundary driver.
- `clearedOrdersChan chan<- ClearedOrders`  
  Notifies OrderSync about which orders were cleared via channel.
- `localStateChan chan<- ElevatorState`  
  Used by OrderSync to build heartbeats and compute assignments.