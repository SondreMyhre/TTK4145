package ordersync	// AI-generert

import (
    elevio "project/elevio"
    localsingle "project/localsingleelevator"
    "testing"
)

// ============================================================================
// Helpers
// ============================================================================

func makeLocalState(floor int, dir localsingle.Direction, behav localsingle.ElevatorBehaviour) localsingle.ElevatorState {
    return localsingle.ElevatorState{
            Floor:     floor,
            Direction: dir,
            Behaviour: behav,
		}
}

func drainCommands(cmds []command) (broadcasts int, lamps int, localOrders int) {
    for _, c := range cmds {
        switch c._type {
        case broadcastNetMessage:
            broadcasts++
        case setButtonLamp:
            lamps++
        case sendOrderToLocal:
            localOrders++
        }
    }
    return
}

func buildNetMsg(id ElevID, hom HallOrderMatrix, cabCalls CabCallsMap, state localsingle.ElevatorState) NetMsg {
    return NetMsg{
        SenderID:        id,
        HallOrderMatrix: hom,
        CabCalls:        cabCalls,
        SenderState:     state,
    }
}

// ============================================================================
// Test: To heiser synker en hall-ordre fra Pending til Confirmed
// ============================================================================

func TestTwoElevators_HallOrderBecomesConfirmed(t *testing.T) {
    var e1Hom HallOrderMatrix
    e1Hom, _ = onHallButtonEvent(e1Hom, elevio.ButtonEvent{Floor: 2, Button: elevio.BT_HallUp})

    if e1Hom[2][0].Status != Pending {
        t.Fatal("e1: ordre burde være Pending etter knappetrykk")
    }

    var e2Hom HallOrderMatrix
    cabCalls := make(CabCallsMap)
    var pending [N_FLOORS]bool

    msg := buildNetMsg("1", e1Hom, cabCalls, localsingle.ElevatorState{Floor: 2})
    e2Hom, _, _, cmds := onNetMsg(e2Hom, cabCalls, "2", pending, nil, msg)

    if e2Hom[2][0].Status != Confirmed {
        t.Fatalf("e2: forventet Confirmed, fikk %d", e2Hom[2][0].Status)
    }

    broadcasts, _, _ := drainCommands(cmds)
    if broadcasts == 0 {
        t.Fatal("e2 burde broadcaste etter bekreftelse")
    }

    msg2 := buildNetMsg("2", e2Hom, cabCalls, localsingle.ElevatorState{Floor: 0})
    e1Hom, _, _, _ = onNetMsg(e1Hom, cabCalls, "1", pending, nil, msg2)

    if e1Hom[2][0].Status != Confirmed {
        t.Fatalf("e1: forventet Confirmed etter synk, fikk %d", e1Hom[2][0].Status)
    }

    if e1Hom[2][0].Version != e2Hom[2][0].Version {
        t.Fatalf("versjonene burde matche: e1=%d, e2=%d", e1Hom[2][0].Version, e2Hom[2][0].Version)
    }
}

// ============================================================================
// Test: Full livssyklus — knappetrykk, bekreft, claim, clear
// ============================================================================

func TestFullOrderLifecycle(t *testing.T) {
    cabCalls := make(CabCallsMap)
    cabCalls["1"] = [N_FLOORS]bool{}
    var pending [N_FLOORS]bool

    var hom HallOrderMatrix
    hom, _ = onHallButtonEvent(hom, elevio.ButtonEvent{Floor: 1, Button: elevio.BT_HallUp})

    var e2Hom HallOrderMatrix
    msg := buildNetMsg("1", hom, cabCalls, localsingle.ElevatorState{Floor: 0})
    e2Hom, _, _, _ = onNetMsg(e2Hom, cabCalls, "2", pending, nil, msg)

    msg2 := buildNetMsg("2", e2Hom, cabCalls, localsingle.ElevatorState{Floor: 3})
    hom, _, _, _ = onNetMsg(hom, cabCalls, "1", pending, nil, msg2)

    if hom[1][0].Status != Confirmed {
        t.Fatalf("forventet Confirmed, fikk %d", hom[1][0].Status)
    }

    order := OrderLocation{Floor: 1, Button: elevio.BT_HallUp, Entry: hom[1][0]}
    hom, cmds := claimOrder(hom, "1", order)

    if hom[1][0].Status != Assigned || hom[1][0].AssignedElevator != "1" {
        t.Fatal("ordren burde være Assigned til 1")
    }

    _, _, localOrders := drainCommands(cmds)
    if localOrders == 0 {
        t.Fatal("burde sende ordre til lokal heis")
    }

    cleared := []localsingle.Order{{Floor: 1, Button: localsingle.BtnHallUp}}
    hom, cabCalls, _ = onClearedOrders(hom, cabCalls, "1", cleared)

    if hom[1][0].Status != Inactive {
        t.Fatal("ordren burde være Inactive etter clearing")
    }
    if hom[1][0].AssignedElevator != "" {
        t.Fatal("assigned elevator burde være tom")
    }
}

// ============================================================================
// Test: Peer dør -> assigned ordre frigjøres
// ============================================================================

func TestPeerDeath_ReleasesOrders(t *testing.T) {
    var hom HallOrderMatrix
    hom[0][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "2", Version: 5}
    hom[3][1] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "1", Version: 3}

    oldPeers := []Peer{{ID: "2", Status: Alive}}
    newPeers := []Peer{{ID: "2", Status: Dead}}

    hom, _, _ = onPeerEvent(hom, oldPeers, newPeers)

    if hom[0][0].Status != Pending || hom[0][0].AssignedElevator != "" {
        t.Fatal("peer 2 sin ordre burde frigjøres")
    }
    if hom[3][1].Status != Assigned || hom[3][1].AssignedElevator != "1" {
        t.Fatal("peer 1 sin ordre burde være uberørt")
    }
}

// ============================================================================
// Test: Cab-ordre synkronisering
// ============================================================================

func TestCabOrderSync(t *testing.T) {
    cabCalls := make(CabCallsMap)
    cabCalls["1"] = [N_FLOORS]bool{}
    var pending [N_FLOORS]bool

    cabCalls, pending, cmds := onCabButtonEvent(cabCalls, pending, "1",
        elevio.ButtonEvent{Floor: 3, Button: elevio.BT_Cab})

    if !cabCalls["1"][3] || !pending[3] {
        t.Fatal("cab call og pending burde være satt")
    }

    _, _, localOrders := drainCommands(cmds)
    if localOrders == 0 {
        t.Fatal("burde sende cab-ordre til lokal heis")
    }

    var hom HallOrderMatrix
    msg := NetMsg{
        SenderID:        "2",
        HallOrderMatrix: hom,
        CabCalls:        CabCallsMap{"1": {false, false, false, true}, "2": {}},
        SenderState:     localsingle.ElevatorState{Floor: 0},
    }

    _, _, pending, _ = onNetMsg(hom, cabCalls, "1", pending, nil, msg)

    if pending[3] {
        t.Fatal("pending burde fjernes etter bekreftelse")
    }
}

// ============================================================================
// Test: Tie-break — laveste ID vinner (numerisk med strconv.Atoi)
// ============================================================================

func TestTieBreak_LowestIDWins(t *testing.T) {
    var hom HallOrderMatrix
    hom[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "5", Version: 4}

    cabCalls := make(CabCallsMap)
    var pending [N_FLOORS]bool

    var remoteHom HallOrderMatrix
    remoteHom[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "2", Version: 4}

    msg := buildNetMsg("2", remoteHom, cabCalls, localsingle.ElevatorState{})
    hom, _, _, _ = onNetMsg(hom, cabCalls, "1", pending, nil, msg)

    if hom[1][0].AssignedElevator != "2" {
        t.Fatalf("2 burde vinne tie-break (lavere), fikk %s", hom[1][0].AssignedElevator)
    }
}

func TestTieBreak_LocalAlreadyLowest(t *testing.T) {
    var hom HallOrderMatrix
    hom[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "1", Version: 4}

    cabCalls := make(CabCallsMap)
    var pending [N_FLOORS]bool

    var remoteHom HallOrderMatrix
    remoteHom[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "3", Version: 4}

    msg := buildNetMsg("3", remoteHom, cabCalls, localsingle.ElevatorState{})
    hom, _, _, _ = onNetMsg(hom, cabCalls, "1", pending, nil, msg)

    if hom[1][0].AssignedElevator != "1" {
        t.Fatalf("1 burde beholde (allerede lavest), fikk %s", hom[1][0].AssignedElevator)
    }
}

func TestTieBreak_EmptyLocalAdoptsRemote(t *testing.T) {
    var hom HallOrderMatrix
    hom[1][0] = OrderMatrixEntry{Status: Confirmed, AssignedElevator: "", Version: 4}

    cabCalls := make(CabCallsMap)
    var pending [N_FLOORS]bool

    var remoteHom HallOrderMatrix
    remoteHom[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "3", Version: 4}

    msg := buildNetMsg("3", remoteHom, cabCalls, localsingle.ElevatorState{})
    hom, _, _, _ = onNetMsg(hom, cabCalls, "1", pending, nil, msg)

    if hom[1][0].AssignedElevator != "3" {
        t.Fatalf("tom local burde adoptere remote, fikk '%s'", hom[1][0].AssignedElevator)
    }
}

func TestTieBreak_EmptyRemoteDoesNotWin(t *testing.T) {
    var hom HallOrderMatrix
    hom[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "2", Version: 4}

    cabCalls := make(CabCallsMap)
    var pending [N_FLOORS]bool

    var remoteHom HallOrderMatrix
    remoteHom[1][0] = OrderMatrixEntry{Status: Confirmed, AssignedElevator: "", Version: 4}

    msg := buildNetMsg("3", remoteHom, cabCalls, localsingle.ElevatorState{})
    hom, _, _, _ = onNetMsg(hom, cabCalls, "1", pending, nil, msg)

    if hom[1][0].AssignedElevator != "2" {
        t.Fatalf("tom remote burde ikke vinne, fikk '%s'", hom[1][0].AssignedElevator)
    }
}

// ============================================================================
// Test: Flere ordrer samtidig konvergerer
// ============================================================================

func TestMultipleOrders_Converge(t *testing.T) {
    var hom HallOrderMatrix
    cabCalls := make(CabCallsMap)
    var pending [N_FLOORS]bool

    hom, _ = onHallButtonEvent(hom, elevio.ButtonEvent{Floor: 0, Button: elevio.BT_HallUp})
    hom, _ = onHallButtonEvent(hom, elevio.ButtonEvent{Floor: 2, Button: elevio.BT_HallDown})
    hom, _ = onHallButtonEvent(hom, elevio.ButtonEvent{Floor: 3, Button: elevio.BT_HallDown})

    var e2Hom HallOrderMatrix
    msg := buildNetMsg("1", hom, cabCalls, localsingle.ElevatorState{})
    e2Hom, _, _, _ = onNetMsg(e2Hom, cabCalls, "2", pending, nil, msg)

    msg2 := buildNetMsg("2", e2Hom, cabCalls, localsingle.ElevatorState{})
    hom, _, _, _ = onNetMsg(hom, cabCalls, "1", pending, nil, msg2)

    for f := range N_FLOORS {
        for b := range N_HALL {
            if hom[f][b] != e2Hom[f][b] {
                t.Fatalf("matrisene divergerer på [%d][%d]: e1=%+v, e2=%+v",
                    f, b, hom[f][b], e2Hom[f][b])
            }
        }
    }
}

// ============================================================================
// Test: Allerede Pending -> ingen versjonsbump
// ============================================================================

func TestRepeatedButtonPress_NoEffect(t *testing.T) {
    var hom HallOrderMatrix
    hom, _ = onHallButtonEvent(hom, elevio.ButtonEvent{Floor: 1, Button: elevio.BT_HallUp})
    v := hom[1][0].Version

    hom, _ = onHallButtonEvent(hom, elevio.ButtonEvent{Floor: 1, Button: elevio.BT_HallUp})

    if hom[1][0].Version != v {
        t.Fatal("versjon burde ikke øke ved gjentatt trykk")
    }
}

// ============================================================================
// Test: Tom cleared-liste gjør ingenting
// ============================================================================

func TestClearedOrders_EmptyList(t *testing.T) {
    var hom HallOrderMatrix
    hom[0][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "1", Version: 3}
    cabCalls := make(CabCallsMap)
    cabCalls["1"] = [N_FLOORS]bool{}

    hom, _, cmds := onClearedOrders(hom, cabCalls, "1", []localsingle.Order{})

    if hom[0][0].Status != Assigned {
        t.Fatal("ingenting burde endres")
    }
    if len(cmds) != 0 {
        t.Fatal("ingen kommandoer forventet")
    }
}

// ============================================================================
// Test: Heis i bevegelse claimer ikke
// ============================================================================

func TestMovingElevator_DoesNotClaim(t *testing.T) {
    var hom HallOrderMatrix
    hom[2][0] = OrderMatrixEntry{Status: Confirmed, Version: 4}
    cabCalls := make(CabCallsMap)

    movingState := makeLocalState(1, localsingle.DirUp, localsingle.BehaviourMoving)
    updatedHom, _, cmds := onNewLocalState(hom, nil, "1", cabCalls, movingState)

    if updatedHom[2][0].Status == Assigned {
        t.Fatal("heis i bevegelse burde ikke claime")
    }
    _, _, lo := drainCommands(cmds)
    if lo != 0 {
        t.Fatal("ingen lokale ordrer under bevegelse")
    }
}

// ============================================================================
// Test: Ignorer egen melding
// ============================================================================

func TestOnNetMsg_IgnoresOwnMessage(t *testing.T) {
    var hom HallOrderMatrix
    hom[0][0] = OrderMatrixEntry{Status: Pending, Version: 1}
    cabCalls := make(CabCallsMap)
    var pending [N_FLOORS]bool

    msg := buildNetMsg("1", HallOrderMatrix{}, cabCalls, localsingle.ElevatorState{})
    hom, _, _, cmds := onNetMsg(hom, cabCalls, "1", pending, nil, msg)

    if len(cmds) != 0 {
        t.Fatal("burde ikke produsere kommandoer for egen melding")
    }
    if hom[0][0].Status != Pending {
        t.Fatal("burde ikke endre matrise for egen melding")
    }
}

// ============================================================================
// Test: Heartbeat
// ============================================================================

func TestHeartbeat(t *testing.T) {
    cmds := onHeartbeatTick()
    if len(cmds) != 1 || cmds[0]._type != broadcastNetMessage {
        t.Fatal("heartbeat burde gi nøyaktig én broadcast")
    }
}

// ============================================================================
// Test: HRA-hjelpefunksjoner
// ============================================================================

func TestHRAHelpers(t *testing.T) {
    if directionToString(localsingle.DirUp) != "up" {
        t.Fatal("DirUp -> up")
    }
    if directionToString(localsingle.DirDown) != "down" {
        t.Fatal("DirDown -> down")
    }
    if directionToString(localsingle.DirStop) != "stop" {
        t.Fatal("DirStop -> stop")
    }
    if behaviourToString(localsingle.BehaviourIdle) != "idle" {
        t.Fatal("Idle -> idle")
    }
    if behaviourToString(localsingle.BehaviourMoving) != "moving" {
        t.Fatal("Moving -> moving")
    }
    if behaviourToString(localsingle.BehaviourDoorOpen) != "doorOpen" {
        t.Fatal("DoorOpen -> doorOpen")
    }

    cabCalls := make(CabCallsMap)
    cabCalls["1"] = [N_FLOORS]bool{true, false, false, true}
    state := localsingle.ElevatorState{Floor: 2, Direction: localsingle.DirUp, Behaviour: localsingle.BehaviourMoving}
    hra := localStateToHRA("1", state, cabCalls)

    if hra.Floor != 2 || hra.Direction != "up" || hra.Behavior != "moving" {
        t.Fatalf("feil HRA-state: %+v", hra)
    }
    if !hra.CabRequests[0] || hra.CabRequests[1] || !hra.CabRequests[3] {
        t.Fatalf("feil cab requests: %+v", hra.CabRequests)
    }
}

// ============================================================================
// Test: Full simulering — to heiser, knappetrykk, synk, claim, clear
// ============================================================================

func TestMockMain_TwoElevators(t *testing.T) {
    var e1Hom, e2Hom HallOrderMatrix
    e1Cab := make(CabCallsMap)
    e2Cab := make(CabCallsMap)
    e1Cab["1"] = [N_FLOORS]bool{}
    e1Cab["2"] = [N_FLOORS]bool{}
    e2Cab["1"] = [N_FLOORS]bool{}
    e2Cab["2"] = [N_FLOORS]bool{}
    var e1Pending, e2Pending [N_FLOORS]bool

    e1State := localsingle.ElevatorState{Floor: 0, Direction: localsingle.DirStop, Behaviour: localsingle.BehaviourIdle}
    e2State := localsingle.ElevatorState{Floor: 3, Direction: localsingle.DirStop, Behaviour: localsingle.BehaviourIdle}

    e1Peers := []Peer{{ID: "2", Status: Alive, State: e2State}}
    e2Peers := []Peer{{ID: "1", Status: Alive, State: e1State}}

    // Knappetrykk
    e1Hom, _ = onHallButtonEvent(e1Hom, elevio.ButtonEvent{Floor: 2, Button: elevio.BT_HallUp})
    t.Log("e1: hall up etasje 2 -> Pending")

    // e1 -> e2
    msg := buildNetMsg("1", e1Hom, e1Cab, e1State)
    e2Hom, e2Cab, e2Pending, _ = onNetMsg(e2Hom, e2Cab, "2", e2Pending, e2Peers, msg)

    if e2Hom[2][0].Status != Confirmed {
        t.Fatalf("e2 burde ha Confirmed, fikk %d", e2Hom[2][0].Status)
    }

    // e2 -> e1
    msg = buildNetMsg("2", e2Hom, e2Cab, e2State)
    e1Hom, e1Cab, e1Pending, _ = onNetMsg(e1Hom, e1Cab, "1", e1Pending, e1Peers, msg)

    if e1Hom[2][0].Status != Confirmed {
        t.Fatalf("e1 burde ha Confirmed, fikk %d", e1Hom[2][0].Status)
    }

    // Konvergens
    for f := range N_FLOORS {
        for b := range N_HALL {
            if e1Hom[f][b] != e2Hom[f][b] {
                t.Fatalf("divergens på [%d][%d]", f, b)
            }
        }
    }
    t.Log("✓ Konvergert etter bekreftelse")

    // e1 claimer
    order := OrderLocation{Floor: 2, Button: elevio.BT_HallUp, Entry: e1Hom[2][0]}
    e1Hom, _ = claimOrder(e1Hom, "1", order)

    if e1Hom[2][0].Status != Assigned || e1Hom[2][0].AssignedElevator != "1" {
        t.Fatal("e1 burde ha Assigned til seg selv")
    }

    // e2 mottar claim
    msg = buildNetMsg("1", e1Hom, e1Cab, e1State)
    e2Hom, e2Cab, e2Pending, _ = onNetMsg(e2Hom, e2Cab, "2", e2Pending, e2Peers, msg)

    if e2Hom[2][0].AssignedElevator != "1" {
        t.Fatalf("e2 burde se assignment til 1, fikk %s", e2Hom[2][0].AssignedElevator)
    }
    t.Log("✓ e2 ser assignment")

    // e1 rydder
    cleared := []localsingle.Order{{Floor: 2, Button: localsingle.BtnHallUp}}
    e1Hom, e1Cab, _ = onClearedOrders(e1Hom, e1Cab, "1", cleared)

    if e1Hom[2][0].Status != Inactive {
        t.Fatal("burde være Inactive etter clear")
    }

    // e2 mottar clear
    msg = buildNetMsg("1", e1Hom, e1Cab, e1State)
    e2Hom, e2Cab, e2Pending, _ = onNetMsg(e2Hom, e2Cab, "2", e2Pending, e2Peers, msg)

    if e2Hom[2][0].Status != Inactive {
        t.Fatalf("e2 burde se Inactive, fikk %d", e2Hom[2][0].Status)
    }

    // Endelig konvergens
    for f := range N_FLOORS {
        for b := range N_HALL {
            if e1Hom[f][b] != e2Hom[f][b] {
                t.Fatalf("endelig divergens på [%d][%d]", f, b)
            }
        }
    }
    t.Log("✓ Full livssyklus fullført")
}