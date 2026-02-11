## ElevIO (Boundary)

### Responsibility
- Boundary module that **owns all hardware IO** to the physical (or simulated) elevator.
- Produces sensor events (based on physical sensors and sends them on the respected channels):
  - button presses
  - floor sensor
  - stop button
  - obstruction switch
- Executes actuator commands (listens on cmd channel for actuator commands):
  - motor direction
  - door lamp
  - button lamps
  - floor indicator
  - stop lamp

### Owns (mutable state)
- TCP connection to elevator or simulator / driver state
- polling goroutines (buttons/floor/stop/obstruction)
- internal buffers required for IO

#### Inputs
- `Cmd <-chan DriverCmd`  
  Commands produced by LocalSingleElevator (and possibly OrderSync).

#### Outputs
- `Buttons chan<- ButtonEvent`
- `Floor chan<- int`
- `Obstruction chan<- bool`

### Functional core vs Imperative shell
- **Shell-only in practice.**
- Optional pure helpers:
  - mapping from `DriverCmd` to concrete `elevio.SetX(...)` calls

