package elevio

type DriverCommandType int

const (
	CommandSetMotorDirection DriverCommandType = iota
	CommandSetButtonLamp
	CommandSetFloorIndicator
	CommandSetDoorLamp
)

type DriverCommand struct {
	Kind           DriverCommandType
	MotorDirection MotorDirection
	Button         ButtonType
	Floor          int
	Value          bool
}

func RunDriver(driverCommandChan <-chan DriverCommand) {
	for command := range driverCommandChan {
		switch command.Kind {
		case CommandSetMotorDirection:
			SetMotorDirection(command.MotorDirection)
		case CommandSetButtonLamp:
			SetButtonLamp(command.Button, command.Floor, command.Value)
		case CommandSetFloorIndicator:
			SetFloorIndicator(command.Floor)
		case CommandSetDoorLamp:
			SetDoorOpenLamp(command.Value)
		}
	}
}
