package ordersync

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	elevio "project/elevio"
	localsingle "project/localsingleelevator"
)

type HRAElevState struct {
	Behavior    string         `json:"behaviour"`
	Floor       int            `json:"floor"`
	Direction   string         `json:"direction"`
	CabRequests [N_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [N_FLOORS][2]bool       `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func directionToString(direction localsingle.Direction) string {
	switch direction {
	case localsingle.DirUp:
		return "up"
	case localsingle.DirDown:
		return "down"
	default:
		return "stop"
	}
}

func behaviourToString(behaviour localsingle.ElevatorBehaviour) string {
	switch behaviour {
	case localsingle.BehaviourMoving:
		return "moving"
	case localsingle.BehaviourDoorOpen:
		return "doorOpen"
	default:
		return "idle"
	}
}

func localStateToHRA(id ElevID, state localsingle.ElevatorState, cabCalls CabCallsMap) HRAElevState {
	return HRAElevState{
		Behavior:    behaviourToString(state.Behaviour),
		Floor:       state.Floor,
		Direction:   directionToString(state.Direction),
		CabRequests: cabCalls[id],
	}
}

func callHRA(hallOrderMatrix HallOrderMatrix, myID ElevID, localState localsingle.ElevatorState, cabCalls CabCallsMap, peerList []Peer) map[string][N_FLOORS][2]bool {
	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	var hallRequests [N_FLOORS][2]bool
	for floor := range N_FLOORS {
		for btn := range N_HALL {
			if hallOrderMatrix[floor][btn].Status == Confirmed {
				hallRequests[floor][btn] = true
			}
		}
	}

	states := make(map[string]HRAElevState)
	states[string(myID)] = localStateToHRA(myID, localState, cabCalls)
	for _, peer := range peerList {
		if peer.PeerStatus == StatusAlive {
			states[string(peer.ID)] = localStateToHRA(peer.ID, peer.state, cabCalls)
		}
	}

	input := HRAInput{
		HallRequests: hallRequests,
		States:       states,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command("./"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][N_FLOORS][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return *output
}

func detertmineMyOrders(hallOrderMatrix HallOrderMatrix, myID ElevID, localState localsingle.ElevatorState, cabCalls CabCallsMap, peerList []Peer) []OrderLocation {
	hasConfirmed := false
	for floor := range N_FLOORS {
		for btn := range N_HALL {
			if hallOrderMatrix[floor][btn].Status == Confirmed {
				hasConfirmed = true
				break
			}
		}
		if hasConfirmed {
			break
		}
	}
	if !hasConfirmed {
		return []OrderLocation{}
	}

	hraResult := callHRA(hallOrderMatrix, myID, localState, cabCalls, peerList)
	myAssignments := hraResult[string(myID)]
	var orders []OrderLocation

	for floor := range N_FLOORS {
		for btn := range N_HALL {
			if myAssignments[floor][btn] {
				orders = append(orders, OrderLocation{
					Floor:  floor,
					Button: elevio.ButtonType(btn),
					Entry:  hallOrderMatrix[floor][btn],
				})
			}
		}
	}

	return orders
}
