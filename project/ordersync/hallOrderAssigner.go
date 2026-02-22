package ordersync

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    localsingle "project/localsingleelevator"
    "runtime"
)

type HRAElevState struct {
    Behavior    string `json:"behaviour"`
    Floor       int    `json:"floor"`
    Direction   string `json:"direction"`
    CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
    HallRequests [][2]bool               `json:"hallRequests"`
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

func localStateToHRA(id ElevID, state LocalState, cabCalls CabCallsMap) HRAElevState {
    cabs := cabCalls[id]
    cabSlice := make([]bool, N_FLOORS)
    for i, v := range cabs {
        cabSlice[i] = v
    }
    return HRAElevState{
        Behavior:    behaviourToString(state.Behaviour),
        Floor:       state.Floor,
        Direction:   directionToString(state.Direction),
        CabRequests: cabSlice,
    }
}

func getHRAExecutablePath() (string, error) {
    var name string
    switch runtime.GOOS {
    case "linux":
        name = "hall_request_assigner"
    case "windows":
        name = "hall_request_assigner.exe"
    default:
        return "", fmt.Errorf("OS not supported: %s", runtime.GOOS)
    }

    // Prøv working directory først
    if _, err := os.Stat(name); err == nil {
        return "./" + name, nil
    }

    // Prøv relativt til executable
    execPath, err := os.Executable()
    if err == nil {
        dir := filepath.Dir(execPath)
        candidate := filepath.Join(dir, name)
        if _, err := os.Stat(candidate); err == nil {
            return candidate, nil
        }
    }

    return "", fmt.Errorf("could not find %s in working directory or executable directory", name)
}

func callHRA(
    hallOrderMatrix HallOrderMatrix,
    myID ElevID,
    localState LocalState,
    cabCalls CabCallsMap,
    peerList []Peer,
) (map[string][][2]bool, error) {

    hraExecutable, err := getHRAExecutablePath()
    if err != nil {
        return nil, err
    }

    hallReqSlice := make([][2]bool, N_FLOORS)
    for floor := range N_FLOORS {
        for btn := range N_HALL {
            if hallOrderMatrix[floor][btn].Status == Confirmed {
                hallReqSlice[floor][btn] = true
            }
        }
    }

    states := make(map[string]HRAElevState)
    states[string(myID)] = localStateToHRA(myID, localState, cabCalls)

    for _, peer := range peerList {
        if peer.ID != myID && peer.Status == Alive {
            states[string(peer.ID)] = localStateToHRA(peer.ID, peer.State, cabCalls)
        }
    }

    input := HRAInput{
        HallRequests: hallReqSlice,
        States:       states,
    }

    jsonBytes, err := json.Marshal(input)
    if err != nil {
        return nil, fmt.Errorf("json.Marshal error: %v", err)
    }

    ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("exec.Command error: %v, output: %s", err, string(ret))
    }

    output := new(map[string][][2]bool)
    err = json.Unmarshal(ret, &output)
    if err != nil {
        return nil, fmt.Errorf("json.Unmarshal error: %v", err)
    }

    return *output, nil
}