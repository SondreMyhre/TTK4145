package ordersync

import (
	elevatorcontroller "project/elevatorcontroller"
)

func RunAssigner(
	myID ElevID,

	worldviewChan <-chan WorldviewMsg,

	assignedRequestsChan chan<- elevatorcontroller.RequestMatrix,
) {
	var prevAssignments [N_FLOORS][N_BUTTONS]bool

	
	for worldview := range worldviewChan {

		combined := elevatorcontroller.RequestMatrix{}
		for floor := 0; floor < N_FLOORS; floor++ {
			combined[floor][elevatorcontroller.BtnHallUp] = worldview.HallRequests[floor][elevatorcontroller.BtnHallUp]
			combined[floor][elevatorcontroller.BtnHallDown] = worldview.HallRequests[floor][elevatorcontroller.BtnHallDown]
			combined[floor][elevatorcontroller.BtnCab] = worldview.CabRequests[myID][floor]
		}

		hasNew := false
		for floor := 0; floor < N_FLOORS; floor++ {
			for button := 0; button < N_BUTTONS; button++ {
				if combined[floor][button] && !prevAssignments[floor][button] {
					hasNew = true
					break
				}
			}
			if hasNew {
				break
			}
		}

		if !hasNew {
			continue
		}


		hraResult, err := AssignRequests(worldview, myID)
		if err != nil {
			continue
		}

		newAssignments := hraResult[string(myID)]
		for floor := range N_FLOORS {
			newAssignments[floor][BT_CAB] = worldview.CabRequests[myID][floor]
		}

		hraResult[string(myID)] = newAssignments

		if newAssignments != prevAssignments {
			prevAssignments = newAssignments
			assignedRequestsChan <- newAssignments
		}
	}
}
