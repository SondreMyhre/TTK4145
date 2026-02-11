## ElevIO (Boundary)

### Responsibility
- Boundary module that **owns all hardware IO** to the physical (or simulated) elevator.
- Produces sensor events:
  - button presses
  - floor sensor
  - stop button
  - obstruction switch
- Executes actuator commands:
  - motor direction
  - door lamp
  - button lamps
  - floor indicator
  - stop lamp

> **Rule:** No other module may call `elevio.Set*` / raw driver functions directly.

---

### Owns (mutable state)
- TCP connection / driver state
- polling goroutines (buttons/floor/stop/obstruction)
- internal buffers required for IO

---

### Run() interface

#### Inputs
- `Cmd <-chan DriverCmd`  
  Commands produced by LocalSingleElevator (and possibly OrderSync).

#### Outputs
- `Buttons chan<- ButtonEvent`
- `Floor chan<- int`
- `Stop chan<- bool`
- `Obstruction chan<- bool`

Optional:
- `Err chan<- error` (recommended, to avoid panics propagating)

---

### Functional core vs Imperative shell
- **Shell-only in practice.**
- Optional pure helpers:
  - mapping from `DriverCmd` to concrete `elevio.SetX(...)` calls

---

### Suggested files
