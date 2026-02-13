package main

import (
	localsingle "Project/LocalSingleElevator"
	elevio "Project/ElevIO"
	"flag"
)

func main() {

	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	elevio.Init(*serverAddr, 4)

	driverCmdChan := make(chan elevio.DriverCmd)
	buttonChan := make(chan elevio.ButtonEvent)
    floorChan := make(chan int)
    obstructionChan := make(chan bool)
    clearedOrdersChan := make(chan []localsingle.Order)
    stateOutChan := make(chan localsingle.LocalSingleElevator)
	
	go elevio.RunDriver(driverCmdChan)
	go elevio.PollButtons(buttonChan)
    go elevio.PollFloorSensor(floorChan)
    go elevio.PollObstructionSwitch(obstructionChan)

	go func() { for range clearedOrdersChan {} }()	// OrderSync not implemented
    go func() { for range stateOutChan {} }()			// OrderSync not implemented

	go localsingle.Run(buttonChan, floorChan, obstructionChan, driverCmdChan, clearedOrdersChan, stateOutChan)

	select {}
}
