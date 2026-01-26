// Package localsingle2 provides a single-elevator FSM implementation.
package localsingle2

import (
	"fmt"
	"strings"
)

// Constants for elevator configuration.
const (
	NumFloors  = 4
	NumButtons = 3
)

// Direction represents elevator movement direction.
type Direction int

const (
	DirDown Direction = -1
	DirStop Direction = 0
	DirUp   Direction = 1
)

// String returns a string representation of the direction.
func (d Direction) String() string {
	switch d {
	case DirUp:
		return "DirUp"
	case DirDown:
		return "DirDown"
	case DirStop:
		return "DirStop"
	default:
		return "DirUndefined"
	}
}

// ButtonType represents the type of elevator button.
type ButtonType int

const (
	ButtonHallUp ButtonType = iota
	ButtonHallDown
	ButtonCab
)

// String returns a string representation of the button type.
func (b ButtonType) String() string {
	switch b {
	case ButtonHallUp:
		return "ButtonHallUp"
	case ButtonHallDown:
		return "ButtonHallDown"
	case ButtonCab:
		return "ButtonCab"
	default:
		return "ButtonUndefined"
	}
}

// Behaviour represents the current state of the elevator.
type Behaviour int

const (
	BehaviourIdle Behaviour = iota
	BehaviourDoorOpen
	BehaviourMoving
)

// String returns a string representation of the behaviour.
func (b Behaviour) String() string {
	switch b {
	case BehaviourIdle:
		return "Idle"
	case BehaviourDoorOpen:
		return "DoorOpen"
	case BehaviourMoving:
		return "Moving"
	default:
		return "Undefined"
	}
}

// Config holds elevator configuration parameters.
type Config struct {
	DoorOpenDuration float64 // Duration in seconds the door stays open.
}

// Elevator represents the state of a single elevator.
type Elevator struct {
	Floor     int
	Direction Direction
	Requests  [NumFloors][NumButtons]bool
	Behaviour Behaviour
	Config    Config
}

// NewElevator creates and returns an uninitialized elevator with default configuration.
func NewElevator() Elevator {
	return Elevator{
		Floor:     -1,
		Direction: DirStop,
		Behaviour: BehaviourIdle,
		Config: Config{
			DoorOpenDuration: 3.0,
		},
	}
}

// Print outputs the current elevator state to stdout.
func (e Elevator) Print() {
	fmt.Println("  +--------------------+")
	fmt.Printf("  |floor = %-2d          |\n", e.Floor)
	fmt.Printf("  |dirn  = %-12s|\n", e.Direction)
	fmt.Printf("  |behav = %-12s|\n", e.Behaviour)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")

	for f := NumFloors - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < NumButtons; btn++ {
			// Skip invalid buttons (no up on top floor, no down on ground floor).
			if (f == NumFloors-1 && btn == int(ButtonHallUp)) ||
				(f == 0 && btn == int(ButtonHallDown)) {
				fmt.Print("|     ")
			} else if e.Requests[f][btn] {
				fmt.Print("|  #  ")
			} else {
				fmt.Print("|  -  ")
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

// PrintInPlace outputs the elevator state and updates in place using ANSI escape codes.
// Call this repeatedly to update the display without scrolling.
func (e Elevator) PrintInPlace() {
	// Number of lines in the output (header + 4 state lines + floor header + NumFloors + bottom border)
	numLines := 7 + NumFloors

	var sb strings.Builder

	// Move cursor up to overwrite previous output
	sb.WriteString(fmt.Sprintf("\033[%dA", numLines))

	// Clear from cursor to end of screen
	sb.WriteString("\033[J")

	// Print state
	sb.WriteString("  +--------------------+\n")
	sb.WriteString(fmt.Sprintf("  |floor = %-2d          |\n", e.Floor))
	sb.WriteString(fmt.Sprintf("  |dirn  = %-12s|\n", e.Direction))
	sb.WriteString(fmt.Sprintf("  |behav = %-12s|\n", e.Behaviour))
	sb.WriteString("  +--------------------+\n")
	sb.WriteString("  |  | up  | dn  | cab |\n")

	for f := NumFloors - 1; f >= 0; f-- {
		sb.WriteString(fmt.Sprintf("  | %d", f))
		for btn := 0; btn < NumButtons; btn++ {
			if (f == NumFloors-1 && btn == int(ButtonHallUp)) ||
				(f == 0 && btn == int(ButtonHallDown)) {
				sb.WriteString("|     ")
			} else if e.Requests[f][btn] {
				sb.WriteString("|  #  ")
			} else {
				sb.WriteString("|  -  ")
			}
		}
		sb.WriteString("|\n")
	}
	sb.WriteString("  +--------------------+\n")

	fmt.Print(sb.String())
}

// PrintInitial prints the elevator state for the first time (no cursor movement).
func (e Elevator) PrintInitial() {
	fmt.Println("  +--------------------+")
	fmt.Printf("  |floor = %-2d          |\n", e.Floor)
	fmt.Printf("  |dirn  = %-12s|\n", e.Direction)
	fmt.Printf("  |behav = %-12s|\n", e.Behaviour)
	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")

	for f := NumFloors - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < NumButtons; btn++ {
			if (f == NumFloors-1 && btn == int(ButtonHallUp)) ||
				(f == 0 && btn == int(ButtonHallDown)) {
				fmt.Print("|     ")
			} else if e.Requests[f][btn] {
				fmt.Print("|  #  ")
			} else {
				fmt.Print("|  -  ")
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}
