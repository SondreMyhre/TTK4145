package elevio

type DriverCmdType int

const (
	CmdSetMotorDirection DriverCmdType = iota
	CmdSetButtonLamp
	CmdSetFloorIndicator
	CmdSetDoorLamp
)

type DriverCmd struct {
	Type    DriverCmdType
	MotorDirection MotorDirection
	Button   ButtonType
	Floor    int
	Value    bool
}

func RunDriver(cmdCh <-chan DriverCmd) {
	for cmd := range cmdCh {
		switch cmd.Type {
		case CmdSetMotorDirection:
			SetMotorDirection(cmd.MotorDirection)
		case CmdSetButtonLamp:
			SetButtonLamp(cmd.Button, cmd.Floor, cmd.Value)
		case CmdSetFloorIndicator:
			SetFloorIndicator(cmd.Floor)
		case CmdSetDoorLamp:
			SetDoorOpenLamp(cmd.Value)
		}
	}
}
