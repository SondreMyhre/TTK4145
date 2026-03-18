package ordersync

import (
	"encoding/json"
	"os/exec"
	elevatorcontroller "project/elevatorcontroller"
	"runtime"
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

func callHRA(input HRAInput) map[string][N_FLOORS][N_BUTTONS]bool {
	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		// fmt.Println("json.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command("./"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		// fmt.Println("exec.Command error: ", err)
		// fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][N_FLOORS][N_BUTTONS]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		// fmt.Println("json.Unmarshal error: ", err)
		return nil
	}

	return *output
}

func AssignRequests(worldview WorldviewMsg, myID ElevID) (map[string][N_FLOORS][N_BUTTONS]bool, error) {
	states := make(map[string]HRAElevState)
	states[string(myID)] = localStateToHRA(worldview.PeerStates[myID], worldview.CabRequests[myID])
	for _, peer := range worldview.Peers {
		if peer.PeerStatus == StatusAlive {
			peerState := worldview.PeerStates[peer.ID]
			if peerState.Obstructed || peerState.MotorStuck {
				continue
			}
			states[string(peer.ID)] = localStateToHRA(peerState, worldview.CabRequests[peer.ID])
		}
	}

	input := HRAInput{
		HallRequests: worldview.HallRequests,
		States:       states,
	}

	hraResult := callHRA(input)
	return hraResult, nil
}

func localStateToHRA(localState elevatorcontroller.ElevatorState, cabRequests [N_FLOORS]bool) HRAElevState {
	return HRAElevState{
		Behavior:    behaviourToString(localState.Behaviour),
		Floor:       localState.Floor,
		Direction:   directionToString(localState.Direction),
		CabRequests: cabRequests,
	}
}

func directionToString(direction elevatorcontroller.Direction) string {
	switch direction {
	case elevatorcontroller.DirUp:
		return "up"
	case elevatorcontroller.DirDown:
		return "down"
	default:
		return "stop"
	}
}

func behaviourToString(behaviour elevatorcontroller.ElevatorBehaviour) string {
	switch behaviour {
	case elevatorcontroller.BehaviourMoving:
		return "moving"
	case elevatorcontroller.BehaviourDoorOpen:
		return "doorOpen"
	default:
		return "idle"
	}
}
