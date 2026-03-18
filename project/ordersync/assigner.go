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
		newAssignments, err := AssignRequests(worldview, myID)
		if err != nil {
			continue
		}

		if newAssignments != prevAssignments {
			prevAssignments = newAssignments
			assignedRequestsChan <- newAssignments
		}
	}
}
