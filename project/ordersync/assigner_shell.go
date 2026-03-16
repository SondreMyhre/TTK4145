package ordersync

import (
	"context"
	localsingle "project/localsingle"
)

func RunAssigner(
	ctx context.Context,
	myID ElevID,

	worldviewChan <-chan WorldviewMsg,

	requestMatrixChan chan<- localsingle.RequestMatrix,
) error {
	var prevRequests [N_FLOORS][N_BUTTONS]bool

	for {
		select {
		case <-ctx.Done():
			return nil

		case worldview := <-worldviewChan:
			assigned, err := AssignRequests(worldview, myID)
			if err != nil {
				// fmt.Printf("assigner: %v\n", err)
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
