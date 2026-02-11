package localsingle

import (
	elevio "Project/ElevIO"
	"fmt"
	"time"
)

func Run(buttonCh <-chan elevio.ButtonEvent, floorCh <-chan int) {
	fmt.Println("Started!")

	elevator := MakeUninitializedElevator()

	if elevio.GetFloor() == -1 {
		elevator.FSM_OnInitBetweenFloors()
	}


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
