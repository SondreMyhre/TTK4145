package elevatorcontroller

import (
	"fmt"
)

func requestsAbove(elevator elevator) bool {
	for floor := elevator.state.Floor + 1; floor < N_FLOORS; floor++ {
		for button := range N_BUTTONS {
			if elevator.requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(elevator elevator) bool {
	for floor := range elevator.state.Floor {
		for button := range N_BUTTONS {
			if elevator.requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func requestsHere(elevator elevator) bool {
	for button := range N_BUTTONS {
		if elevator.requests[elevator.state.Floor][button] {
			return true
		}
	}
	return false
}

// func mergeRequests(localRequests RequestMatrix, remoteRequests RequestMatrix) RequestMatrix {
// 	result := remoteRequests

// 	for floor := range N_FLOORS {
// 		for button := range N_BUTTONS {
// 			result[floor][button] = localRequests[floor][button] || remoteRequests[floor][button]
// 		} 
// 	}
// 	return result
// }

func chooseDirection(elevator elevator) directionBehaviourPair {
	switch elevator.state.Direction {
	case DirUp:
		if requestsAbove(elevator) {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else if requestsHere(elevator) {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if requestsBelow(elevator) {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirDown:
		if requestsBelow(elevator) {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else if requestsHere(elevator) {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if requestsAbove(elevator) {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirStop:
		if requestsHere(elevator) {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if requestsAbove(elevator) {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else if requestsBelow(elevator) {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	default:
		return directionBehaviourPair{DirStop, BehaviourIdle}
	}
}

func shouldStop(elevator elevator) bool {
	switch elevator.state.Direction {
	case DirDown:
		return elevator.requests[elevator.state.Floor][BtnHallDown] ||
			elevator.requests[elevator.state.Floor][BtnCab] ||
			!requestsBelow(elevator)
	case DirUp:
		return elevator.requests[elevator.state.Floor][BtnHallUp] ||
			elevator.requests[elevator.state.Floor][BtnCab] ||
			!requestsAbove(elevator)
	default:
		return true
	}
}

func clearAtCurrentFloor(elevator elevator) (elevator, []Order) {
	var clearedOrders []Order

	if elevator.requests[elevator.state.Floor][BtnCab] {
		elevator.requests[elevator.state.Floor][BtnCab] = false
		clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnCab})
	}

	switch elevator.state.Direction {
	case DirUp:
		fmt.Println("case DirUp")
		c1 := !requestsAbove(elevator)
		c2 := elevator.state.Behaviour != BehaviourDoorOpen
		c3 := elevator.requests[elevator.state.Floor][BtnHallDown]
		all := c1  && c2 &&c3
		fmt.Printf("%t %t %t = %t\n", c1, c2, c3, all)

		if elevator.requests[elevator.state.Floor][BtnHallUp] {
			elevator.requests[elevator.state.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallUp})
		} else if !requestsAbove(elevator) && elevator.state.Behaviour != BehaviourDoorOpen && elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
		}

	case DirDown:
		if elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
		} else if !requestsBelow(elevator) && elevator.state.Behaviour != BehaviourDoorOpen && elevator.requests[elevator.state.Floor][BtnHallUp] {
			elevator.requests[elevator.state.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallUp})
		}

	default:
		if elevator.requests[elevator.state.Floor][BtnHallUp] {
			elevator.requests[elevator.state.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallUp})
		}

		if elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
			fmt.Println("clear HALL_DOWN case default")
		}
	}

	return elevator, clearedOrders
}
