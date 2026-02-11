package elevio

type DriverCmdType int

const (
	CmdSetMotor DriverCmdType = iota
	CmdSetButtonLamp
	CmdSetFloorIndicator
	CmdSetDoorLamp
)

type DriverCmd struct {
	_type DriverCmdType
	MotorDir MotorDirection
	Button ButtonType
	Floor int
	Value bool
}

// type DriverCmd struct {
// 	_type DriverCmdType
// 	value any
// }

func RunDriver(cmdCh <-chan DriverCmd) {
	for cmd := range cmdCh {
		switch cmd._type {
		case CmdSetMotor:
			SetMotorDirection(cmd.MotorDir)
		case CmdSetButtonLamp:
			SetButtonLamp(cmd.Button, cmd.Floor, cmd.Value)
		case CmdSetFloorIndicator:
			SetFloorIndicator(cmd.Floor)
		case CmdSetDoorLamp:
			SetDoorOpenLamp(cmd.Value)
		}
}
}
