package localsingle2

import (
	elevio "TTK4145/ElevIO"
	"fmt"
	"time"
)

// Default configuration values.
const (
	DefaultServerAddr       = "localhost:15657"
	DefaultDoorOpenDuration = 3.0
)

// Run starts the main elevator control loop using ElevIO.
// This is the entry point for running the single elevator FSM.
func Run() {
	fmt.Println("Started!")
	fmt.Println()

	// Initialize hardware via ElevIO.
	elevio.Init(DefaultServerAddr, NumFloors)

	// Create FSM.
	fsm := NewFSM()
	fsm.SetDoorOpenDuration(DefaultDoorOpenDuration)

	// Initialize if starting between floors.
	if elevio.GetFloor() == -1 {
		fsm.OnInitBetweenFloors()
	}

	// Print initial state.
	fsm.PrintInitialState()

	// Channels for events from ElevIO polling.
	buttonCh := make(chan elevio.ButtonEvent)
	floorCh := make(chan int)

	// Start ElevIO polling goroutines.
	go elevio.PollButtons(buttonCh)
	go elevio.PollFloorSensor(floorCh)

	// Main event loop.
	for {
		select {
		case evt := <-buttonCh:
			fsm.OnRequestButtonPress(evt.Floor, ButtonType(evt.Button))

		case floor := <-floorCh:
			fsm.OnFloorArrival(floor)

		default:
			// Check timer.
			if TimerTimedOut() {
				TimerStop()
				fsm.OnDoorTimeout()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
