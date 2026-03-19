## ElevIO - Hardware Interface

### Overall Responsibility

Provides the **low-level interface to the elevator hardware**.

---

### How It Works

ElevIO provides three main polling routines and one command executor:

**Polling (continuously monitoring hardware)**:
- `PollButtons()` - Detects when user presses buttons (hall/cab)
- `PollFloorSensor()` - Detects when elevator passes a floor
- `PollObstructionSwitch()` - Detects when door is obstructed

**Command Execution**:
- `RunDriver()` - Listens for commands and executes them (motor, door lamp, floor indicator)

**Socket Connection**:
- Maintains TCP connection to elevator simulator
- Encodes/decodes command and sensor data

---

### Inputs (receive-only channels)

- `driverCommandChan <-chan elevio.DriverCommand` (from ElevatorController)
  - Commands to execute on hardware
  - Types:
    - `setMotorDirection` - Motor: up/down/stop
    - `setDoorOpenLamp` - Door lamp: on/off
    - `setFloorIndicator` - Floor lamp: which floor
    - `setButtonLamp` - Hall/cab button lamp: on/off

### Outputs (send-only channels)

- `buttonChan chan<- elevio.ButtonEvent` (to OrderSync)
  - User button presses (hall up/down, cab)
  - Format: `{Floor int, Button ButtonType}`
  - Emitted when: user presses button

- `floorChan chan<- int` (to ElevatorController)
  - Current floor of elevator
  - Format: floor number (0 to N_FLOORS-1)
  - Emitted when: elevator passes a floor

- `obstructionChan chan<- bool` (to ElevatorController)
  - Door obstruction status
  - Format: true = obstructed, false = clear
  - Emitted when: obstruction state changes
