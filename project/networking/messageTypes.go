package networking

type NetMsg struct {
	SenderID        ElevID
	HallOrderMatrix HallOrderMatrix
	CabCalls        CabCallsMap
	SenderState     localsingle.ElevatorState
}

