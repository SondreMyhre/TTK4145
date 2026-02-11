package main

import (
	localsingle "Project/LocalSingleElevator"
	elevio "Project/ElevIO"
	"flag"
)

func main() {

	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the server")
	flag.Parse()

	elevio.Init(*serverAddr, 4)

	cmdCh := make(chan elevio.DriverCmd)
	buttonCh := make(chan elevio.ButtonEvent, 16)
    floorCh := make(chan int, 16)
    obstructionCh := make(chan bool, 16)
    clearedCh := make(chan []localsingle.Order, 16)
    stateOutCh := make(chan localsingle.LocalSingleElevator, 1)
	
	go elevio.RunDriver(cmdCh)
	go elevio.PollButtons(buttonCh)
    go elevio.PollFloorSensor(floorCh)
    go elevio.PollObstructionSwitch(obstructionCh)

	go elevio.PollButtons(buttonCh)	// Her vil vi Polle fra OrderSync også
	go elevio.PollFloorSensor(floorCh)

	go func() { for range clearedCh {} }()
    go func() { for range stateOutCh {} }()

	localsingle.Run(buttonCh, floorCh, obstructionCh, cmdCh, clearedCh, stateOutCh)

	// select {}
}
