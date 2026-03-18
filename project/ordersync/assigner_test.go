package ordersync

import (
	"testing"
	"time"

	elevatorcontroller "project/elevatorcontroller"
)

func TestAssigner(t *testing.T) {
    myID := ElevID("1")
    peer2 := ElevID("2")

    worldviewChan := make(chan WorldviewMsg, 1)
    assignedRequestsChan := make(chan elevatorcontroller.RequestMatrix, 1)

    go RunAssigner(myID, worldviewChan, assignedRequestsChan)

    var hall HallRequests
    // hall[2][elevatorcontroller.BtnHallUp] = true
	hall[2][elevatorcontroller.BtnHallDown] = true

    cab := make(map[ElevID][N_FLOORS]bool)
    cab[myID] = [N_FLOORS]bool{}
    cab[peer2] = [N_FLOORS]bool{}

    peerStates := make(map[ElevID]elevatorcontroller.ElevatorState)
    peerStates[myID] = elevatorcontroller.ElevatorState{Floor: 2, Behaviour: elevatorcontroller.BehaviourDoorOpen}
    peerStates[peer2] = elevatorcontroller.ElevatorState{Floor: 1, Behaviour: elevatorcontroller.BehaviourMoving}

    worldviewChan <- WorldviewMsg{
        HallRequests: hall,
        CabRequests:  cab,
        PeerStates:   peerStates,
        Peers: []Peer{
            {ID: myID, PeerStatus: StatusAlive},
            {ID: peer2, PeerStatus: StatusAlive},
        },
    }

    select {
	case assigned := <-assignedRequestsChan:
		t.Logf("Assigned matrix: %v", assigned)
	case <-time.After(10 * time.Second):
		t.Fatal("no assignment received (HRA binary missing?)")
	}
}