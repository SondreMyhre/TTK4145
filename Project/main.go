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
	
	go elevio.RunDriver(cmdCh)

	
	buttonCh := make(chan elevio.ButtonEvent)
	floorCh := make(chan int)

	go elevio.PollButtons(buttonCh)	// Her vil vi Polle fra OrderSync også
	go elevio.PollFloorSensor(floorCh)

	go localsingle.Run(buttonCh, floorCh)

	select {}
}
