package main

import (
	elevio "Project/ElevIO"
	localsingle "Project/LocalSingleElevator"
	"flag"
)

func main() {
	serverAddr := flag.String("serverAddr", "localhost:15657", "IP-address of the elevatorserver or simulatorserver")
	flag.Parse()

	elevio.Init(*serverAddr, 4)

	driverCommandChan := make(chan elevio.DriverCommand)
	buttonChan := make(chan elevio.ButtonEvent)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)
	clearedOrdersChan := make(chan []localsingle.Order)
	localStateChan := make(chan localsingle.LocalSingleElevator)

	go elevio.RunDriver(driverCommandChan)
	go elevio.PollButtons(buttonChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	go func() { for range clearedOrdersChan {} }()	// OrderSync ikke implementert
    go func() { for range localStateChan {} }()		// OrderSync ikke implementert

	go localsingle.Run(buttonChan, floorChan, obstructionChan, driverCommandChan, clearedOrdersChan, localStateChan) // localOrderChan = buttonChan

	select {}
}
