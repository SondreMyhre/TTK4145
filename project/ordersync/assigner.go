package ordersync

import (
	"context"
	localsingle "project/elevatorcontroller"
)

func RunAssigner(
	ctx context.Context,
	myID ElevID,

	worldviewChan <-chan WorldviewMsg,

	assignedRequestsChan chan<- localsingle.RequestMatrix,
) error {
	var prevAssignments [N_FLOORS][N_BUTTONS]bool

	for {
		select {
		case <-ctx.Done():
			return nil

		case worldview := <-worldviewChan:
			newAssignments, err := AssignRequests(worldview, myID)
			if err != nil {
				continue
			}

			if newAssignments != prevAssignments {
				prevAssignments = newAssignments
				select {
				case assignedRequestsChan <- newAssignments:
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}
