package localsingle

import (
	elevio "TTK4145/ElevIO"
	"fmt"
	"time"
)

const (
	DefaultServerAddr       = "localhost:15657" // TODO: gjøre dette mulig å sette med flags
	DefaultDoorOpenDuration = 3.0
)

func Run() {
	fmt.Println("Started!")
	elevio.Init(DefaultServerAddr, 4)

	elevator := MakeUninitializedElevator()

	if elevio.GetFloor() == -1 {
		elevator.FSM_OnInitBetweenFloors()
	}

	buttonCh := make(chan elevio.ButtonEvent)
	floorCh := make(chan int)

	go elevio.PollButtons(buttonCh)	// Her vil vi Polle fra OrderSync også
	go elevio.PollFloorSensor(floorCh)

	for {
		select {
		case evt := <-buttonCh:
			elevator.FSM_OnRequestButtonPress(evt.Floor, evt.Button)	// Her vil vi sjekke if Cab

		case floor := <-floorCh:
			elevator.FSM_OnFloorArrival(floor)

		case <-elevator.doorTimer.C:
            if elevator.doorTimer != nil {
                elevator.FSM_OnDoorTimeout()
            }
            
        default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
