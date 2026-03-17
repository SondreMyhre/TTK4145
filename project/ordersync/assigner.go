package ordersync

import (
	"context"
	elevatorcontroller "project/elevatorcontroller"
)

func RunAssigner(
	ctx context.Context,
	myID ElevID,

	worldviewChan <-chan WorldviewMsg,

	requestMatrixChan chan<- elevatorcontroller.RequestMatrix,
) error {
	var prevRequests [N_FLOORS][N_BUTTONS]bool

	for {
		select {
		case <-ctx.Done():
			return nil

		case worldview := <-worldviewChan:
			assigned, err := AssignRequests(worldview, myID)
			if err != nil {
				continue
			}

			if assigned != prevRequests {
				prevRequests = assigned
				select {
				case requestMatrixChan <- assigned:
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}
