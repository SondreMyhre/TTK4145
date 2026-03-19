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
