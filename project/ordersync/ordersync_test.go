package ordersync	// AI-generert testkode

import (
    elevio "project/elevio"
    localsingle "project/localsingleelevator"
    "testing"
    "time"
)

// ============================================================================
// executeCommands tests
// ============================================================================

func TestExecuteCommands_SendOrderToLocal(t *testing.T) {
    localOrderChan := make(chan elevio.ButtonEvent, 10)
    tx := make(chan NetMsg, 10)
    lightCommandChan := make(chan elevio.DriverCommand, 10)

    evt := elevio.ButtonEvent{Floor: 2, Button: elevio.BT_HallUp}
    commands := []command{
        {_type: sendOrderToLocal, value: evt},
    }

    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    myID := ElevID("elev1")
    localState := LocalState{}

    executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

    select {
    case got := <-localOrderChan:
        if got.Floor != 2 || got.Button != elevio.BT_HallUp {
            t.Fatalf("expected floor=2 BT_HallUp, got floor=%d button=%d", got.Floor, got.Button)
        }
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected a ButtonEvent on localOrderChan")
    }

    assertEmpty(t, tx, lightCommandChan)
}

func TestExecuteCommands_BroadcastNetMessage(t *testing.T) {
    localOrderChan := make(chan elevio.ButtonEvent, 10)
    tx := make(chan NetMsg, 10)
    lightCommandChan := make(chan elevio.DriverCommand, 10)

    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[1][0] = OrderMatrixEntry{Status: Pending, Version: 3}

    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{true, false, false, true}

    myID := ElevID("elev1")
    localState := LocalState{
        Floor:     2,
        Direction: localsingle.DirUp,
        Behaviour: localsingle.BehaviourMoving,
    }

    commands := []command{
        {_type: broadcastNetMessage, value: nil},
    }

    executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

    select {
    case msg := <-tx:
        if msg.SenderID != myID {
            t.Fatalf("expected SenderID=%s, got %s", myID, msg.SenderID)
        }
        if msg.HallOrderMatrix[1][0].Status != Pending || msg.HallOrderMatrix[1][0].Version != 3 {
            t.Fatalf("HallOrderMatrix mismatch: %+v", msg.HallOrderMatrix[1][0])
        }
        if msg.SenderState.Floor != 2 || msg.SenderState.Direction != localsingle.DirUp {
            t.Fatalf("SenderState mismatch: %+v", msg.SenderState)
        }
        if msg.CabCalls["elev1"][0] != true || msg.CabCalls["elev1"][3] != true {
            t.Fatalf("CabCalls mismatch: %+v", msg.CabCalls)
        }
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected a NetMsg on tx")
    }

    assertEmptyLocal(t, localOrderChan)
    assertEmptyLight(t, lightCommandChan)
}

func TestExecuteCommands_BroadcastClonesMap(t *testing.T) {
    localOrderChan := make(chan elevio.ButtonEvent, 10)
    tx := make(chan NetMsg, 10)
    lightCommandChan := make(chan elevio.DriverCommand, 10)

    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{true, false, false, false}
    myID := ElevID("elev1")
    localState := LocalState{}

    commands := []command{
        {_type: broadcastNetMessage, value: nil},
    }

    executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

    msg := <-tx

    cabCalls["elev1"] = LocalCabCalls{false, false, false, false}

    if msg.CabCalls["elev1"][0] != true {
        t.Fatal("broadcast should clone cabCalls; mutation of original affected sent message")
    }
}

func TestExecuteCommands_SetButtonLamp(t *testing.T) {
    localOrderChan := make(chan elevio.ButtonEvent, 10)
    tx := make(chan NetMsg, 10)
    lightCommandChan := make(chan elevio.DriverCommand, 10)

    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    myID := ElevID("elev1")
    localState := LocalState{}

    args := buttonLampArgs{
        Floor:  3,
        Button: elevio.BT_HallDown,
        Value:  true,
    }

    commands := []command{
        {_type: setButtonLamp, value: args},
    }

    executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

    select {
    case cmd := <-lightCommandChan:
        if cmd.Type != elevio.CommandSetButtonLamp {
            t.Fatalf("expected CommandSetButtonLamp, got %d", cmd.Type)
        }
        if cmd.Floor != 3 {
            t.Fatalf("expected floor=3, got %d", cmd.Floor)
        }
        if cmd.Button != elevio.BT_HallDown {
            t.Fatalf("expected BT_HallDown, got %d", cmd.Button)
        }
        if cmd.Value != true {
            t.Fatalf("expected value=true, got %v", cmd.Value)
        }
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected a DriverCommand on lightCommandChan")
    }

    assertEmptyLocal(t, localOrderChan)
    assertEmptyTx(t, tx)
}

func TestExecuteCommands_MultipleCommands(t *testing.T) {
    localOrderChan := make(chan elevio.ButtonEvent, 10)
    tx := make(chan NetMsg, 10)
    lightCommandChan := make(chan elevio.DriverCommand, 10)

    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    myID := ElevID("elev1")
    localState := LocalState{Floor: 1}

    evt := elevio.ButtonEvent{Floor: 0, Button: elevio.BT_Cab}
    lampArgs := buttonLampArgs{Floor: 0, Button: elevio.BT_Cab, Value: true}

    commands := []command{
        {_type: sendOrderToLocal, value: evt},
        {_type: setButtonLamp, value: lampArgs},
        {_type: broadcastNetMessage, value: nil},
    }

    executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

    select {
    case got := <-localOrderChan:
        if got.Floor != 0 {
            t.Fatalf("expected floor=0, got %d", got.Floor)
        }
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected ButtonEvent on localOrderChan")
    }

    select {
    case cmd := <-lightCommandChan:
        if cmd.Floor != 0 || cmd.Button != elevio.BT_Cab || cmd.Value != true {
            t.Fatalf("unexpected light command: %+v", cmd)
        }
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected DriverCommand on lightCommandChan")
    }

    select {
    case msg := <-tx:
        if msg.SenderID != myID {
            t.Fatalf("expected SenderID=%s, got %s", myID, msg.SenderID)
        }
    case <-time.After(100 * time.Millisecond):
        t.Fatal("expected NetMsg on tx")
    }
}

func TestExecuteCommands_EmptyCommands(t *testing.T) {
    localOrderChan := make(chan elevio.ButtonEvent, 10)
    tx := make(chan NetMsg, 10)
    lightCommandChan := make(chan elevio.DriverCommand, 10)

    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    myID := ElevID("elev1")
    localState := LocalState{}

    executeCommands(nil, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)
    executeCommands([]command{}, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

    assertEmpty(t, tx, lightCommandChan)
    assertEmptyLocal(t, localOrderChan)
}

// ============================================================================
// onCabButtonEvent tests
// ============================================================================

func TestOnCabButtonEvent_AddsCabCallAndPending(t *testing.T) {
    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{}
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")
    evt := elevio.ButtonEvent{Floor: 3, Button: elevio.BT_Cab}

    updatedCabCalls, updatedPending, commands := onCabButtonEvent(cabCalls, pendingCabCalls, myID, evt)

    if !updatedCabCalls[myID][3] {
        t.Fatal("expected cab call at floor 3 to be true")
    }
    if !updatedPending[3] {
        t.Fatal("expected pending cab call at floor 3")
    }

    assertHasCommandType(t, commands, sendOrderToLocal)
    assertHasCommandType(t, commands, broadcastNetMessage)
}

func TestOnCabButtonEvent_SendsCorrectButtonEvent(t *testing.T) {
    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{}
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")
    evt := elevio.ButtonEvent{Floor: 1, Button: elevio.BT_Cab}

    _, _, commands := onCabButtonEvent(cabCalls, pendingCabCalls, myID, evt)

    for _, cmd := range commands {
        if cmd._type == sendOrderToLocal {
            got := cmd.value.(elevio.ButtonEvent)
            if got.Floor != 1 || got.Button != elevio.BT_Cab {
                t.Fatalf("expected floor=1 BT_Cab, got floor=%d button=%d", got.Floor, got.Button)
            }
            return
        }
    }
    t.Fatal("expected sendOrderToLocal command")
}

func TestOnCabButtonEvent_MultipleCabCalls(t *testing.T) {
    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{}
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    cabCalls, pendingCabCalls, _ = onCabButtonEvent(cabCalls, pendingCabCalls, myID, elevio.ButtonEvent{Floor: 0, Button: elevio.BT_Cab})
    cabCalls, pendingCabCalls, _ = onCabButtonEvent(cabCalls, pendingCabCalls, myID, elevio.ButtonEvent{Floor: 3, Button: elevio.BT_Cab})

    if !cabCalls[myID][0] || !cabCalls[myID][3] {
        t.Fatalf("expected cab calls at floors 0 and 3, got %+v", cabCalls[myID])
    }
    if !pendingCabCalls[0] || !pendingCabCalls[3] {
        t.Fatalf("expected pending at floors 0 and 3, got %+v", pendingCabCalls)
    }
}

// ============================================================================
// onHallButtonEvent tests
// ============================================================================

func TestOnHallButtonEvent_InactiveTosPending(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    evt := elevio.ButtonEvent{Floor: 2, Button: elevio.BT_HallUp}

    updatedMatrix, commands := onHallButtonEvent(hallOrderMatrix, evt)

    if updatedMatrix[2][0].Status != Pending {
        t.Fatalf("expected Pending, got %d", updatedMatrix[2][0].Status)
    }
    if updatedMatrix[2][0].Version != 1 {
        t.Fatalf("expected Version=1, got %d", updatedMatrix[2][0].Version)
    }
    if updatedMatrix[2][0].AssignedElevator != "" {
        t.Fatalf("expected empty AssignedElevator, got %s", updatedMatrix[2][0].AssignedElevator)
    }

    assertHasCommandType(t, commands, broadcastNetMessage)
}

func TestOnHallButtonEvent_HallDown(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    evt := elevio.ButtonEvent{Floor: 1, Button: elevio.BT_HallDown}

    updatedMatrix, _ := onHallButtonEvent(hallOrderMatrix, evt)

    if updatedMatrix[1][1].Status != Pending {
        t.Fatalf("expected Pending for HallDown at floor 1, got %d", updatedMatrix[1][1].Status)
    }
}

func TestOnHallButtonEvent_AlreadyPending_NoVersionBump(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[1][0] = OrderMatrixEntry{Status: Pending, Version: 5}
    evt := elevio.ButtonEvent{Floor: 1, Button: elevio.BT_HallUp}

    updatedMatrix, _ := onHallButtonEvent(hallOrderMatrix, evt)

    if updatedMatrix[1][0].Version != 5 {
        t.Fatalf("should not bump version for already-pending order, got %d", updatedMatrix[1][0].Version)
    }
    if updatedMatrix[1][0].Status != Pending {
        t.Fatalf("status should remain Pending, got %d", updatedMatrix[1][0].Status)
    }
}

func TestOnHallButtonEvent_AlreadyAssigned_NoChange(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[0][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 3}
    evt := elevio.ButtonEvent{Floor: 0, Button: elevio.BT_HallUp}

    updatedMatrix, _ := onHallButtonEvent(hallOrderMatrix, evt)

    if updatedMatrix[0][0].Status != Assigned {
        t.Fatalf("expected status to remain Assigned, got %d", updatedMatrix[0][0].Status)
    }
    if updatedMatrix[0][0].Version != 3 {
        t.Fatalf("expected version to remain 3, got %d", updatedMatrix[0][0].Version)
    }
}

func TestOnHallButtonEvent_AlreadyConfirmed_NoChange(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[2][1] = OrderMatrixEntry{Status: Confirmed, Version: 7}
    evt := elevio.ButtonEvent{Floor: 2, Button: elevio.BT_HallDown}

    updatedMatrix, _ := onHallButtonEvent(hallOrderMatrix, evt)

    if updatedMatrix[2][1].Status != Confirmed {
        t.Fatalf("expected status to remain Confirmed, got %d", updatedMatrix[2][1].Status)
    }
    if updatedMatrix[2][1].Version != 7 {
        t.Fatalf("expected version to remain 7, got %d", updatedMatrix[2][1].Version)
    }
}

// ============================================================================
// onNewLocalState tests
// ============================================================================

func TestOnNewLocalState_IdleWithNoPendingOrders_NoCommands(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    var peerList []Peer
    myID := ElevID("elev1")

    newState := localsingle.LocalSingleElevator{
        State: localsingle.ElevatorState{
            Floor:     1,
            Direction: localsingle.DirStop,
            Behaviour: localsingle.BehaviourIdle,
        },
    }

    _, localState, commands := onNewLocalState(hallOrderMatrix, peerList, myID, newState)

    if localState.Floor != 1 {
        t.Fatalf("expected floor=1, got %d", localState.Floor)
    }
    if localState.Behaviour != localsingle.BehaviourIdle {
        t.Fatalf("expected BehaviourIdle, got %d", localState.Behaviour)
    }
    if len(commands) != 0 {
        t.Fatalf("expected no commands, got %d", len(commands))
    }
}

func TestOnNewLocalState_UpdatesLocalStateFields(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    var peerList []Peer
    myID := ElevID("elev1")

    newState := localsingle.LocalSingleElevator{
        State: localsingle.ElevatorState{
            Floor:     3,
            Direction: localsingle.DirDown,
            Behaviour: localsingle.BehaviourMoving,
        },
    }

    _, localState, _ := onNewLocalState(hallOrderMatrix, peerList, myID, newState)

    if localState.Floor != 3 {
        t.Fatalf("expected floor=3, got %d", localState.Floor)
    }
    if localState.Direction != localsingle.DirDown {
        t.Fatalf("expected DirDown, got %d", localState.Direction)
    }
    if localState.Behaviour != localsingle.BehaviourMoving {
        t.Fatalf("expected BehaviourMoving, got %d", localState.Behaviour)
    }
}

// ============================================================================
// onClearedOrders tests
// ============================================================================

func TestOnClearedOrders_ClearsHallOrder(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[2][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 5}

    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{}
    myID := ElevID("elev1")

    clearedOrders := []localsingle.Order{
        {Floor: 2, Button: localsingle.BtnHallUp},
    }

    updatedMatrix, _, commands := onClearedOrders(hallOrderMatrix, cabCalls, myID, clearedOrders)

    if updatedMatrix[2][0].Status != Inactive {
        t.Fatalf("expected Inactive after clearing, got %d", updatedMatrix[2][0].Status)
    }
    if updatedMatrix[2][0].AssignedElevator != "" {
        t.Fatalf("expected empty AssignedElevator, got %s", updatedMatrix[2][0].AssignedElevator)
    }
    if updatedMatrix[2][0].Version != 6 {
        t.Fatalf("expected version=6, got %d", updatedMatrix[2][0].Version)
    }

    assertHasCommandType(t, commands, setButtonLamp)
    assertHasCommandType(t, commands, broadcastNetMessage)

    // Verify lamp is turned off
    for _, cmd := range commands {
        if cmd._type == setButtonLamp {
            args := cmd.value.(buttonLampArgs)
            if args.Floor == 2 && args.Button == elevio.BT_HallUp && args.Value != false {
                t.Fatal("expected lamp value=false for cleared hall order")
            }
        }
    }
}

func TestOnClearedOrders_ClearsHallDown(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[1][1] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 3}

    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{}
    myID := ElevID("elev1")

    clearedOrders := []localsingle.Order{
        {Floor: 1, Button: localsingle.BtnHallDown},
    }

    updatedMatrix, _, commands := onClearedOrders(hallOrderMatrix, cabCalls, myID, clearedOrders)

    if updatedMatrix[1][1].Status != Inactive {
        t.Fatalf("expected Inactive, got %d", updatedMatrix[1][1].Status)
    }
    if updatedMatrix[1][1].Version != 4 {
        t.Fatalf("expected version=4, got %d", updatedMatrix[1][1].Version)
    }

    assertHasCommandType(t, commands, setButtonLamp)
}

func TestOnClearedOrders_ClearsCabOrder(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix

    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{true, false, true, false}
    myID := ElevID("elev1")

    clearedOrders := []localsingle.Order{
        {Floor: 0, Button: localsingle.BtnCab},
    }

    _, updatedCabCalls, commands := onClearedOrders(hallOrderMatrix, cabCalls, myID, clearedOrders)

    if updatedCabCalls[myID][0] != false {
        t.Fatal("expected cab call at floor 0 to be cleared")
    }
    if updatedCabCalls[myID][2] != true {
        t.Fatal("expected cab call at floor 2 to remain")
    }

    assertHasCommandType(t, commands, setButtonLamp)
    assertHasCommandType(t, commands, broadcastNetMessage)
}

func TestOnClearedOrders_ClearsMultipleOrders(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[0][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 2}
    hallOrderMatrix[3][1] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 4}

    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{false, true, false, false}
    myID := ElevID("elev1")

    clearedOrders := []localsingle.Order{
        {Floor: 0, Button: localsingle.BtnHallUp},
        {Floor: 3, Button: localsingle.BtnHallDown},
        {Floor: 1, Button: localsingle.BtnCab},
    }

    updatedMatrix, updatedCabCalls, commands := onClearedOrders(hallOrderMatrix, cabCalls, myID, clearedOrders)

    if updatedMatrix[0][0].Status != Inactive {
        t.Fatalf("expected floor 0 HallUp Inactive, got %d", updatedMatrix[0][0].Status)
    }
    if updatedMatrix[3][1].Status != Inactive {
        t.Fatalf("expected floor 3 HallDown Inactive, got %d", updatedMatrix[3][1].Status)
    }
    if updatedCabCalls[myID][1] != false {
        t.Fatal("expected cab call at floor 1 to be cleared")
    }

    // Should have 3 setButtonLamp + 1 broadcastNetMessage
    lampCount := 0
    for _, cmd := range commands {
        if cmd._type == setButtonLamp {
            lampCount++
        }
    }
    if lampCount != 3 {
        t.Fatalf("expected 3 setButtonLamp commands, got %d", lampCount)
    }
    assertHasCommandType(t, commands, broadcastNetMessage)
}

func TestOnClearedOrders_AlreadyInactive_NoLampCommand(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    // floor 2 HallUp is already Inactive (default)

    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{}
    myID := ElevID("elev1")

    clearedOrders := []localsingle.Order{
        {Floor: 2, Button: localsingle.BtnHallUp},
    }

    _, _, commands := onClearedOrders(hallOrderMatrix, cabCalls, myID, clearedOrders)

    // Should only have broadcastNetMessage, no setButtonLamp for already-inactive
    for _, cmd := range commands {
        if cmd._type == setButtonLamp {
            args := cmd.value.(buttonLampArgs)
            if args.Floor == 2 && args.Button == elevio.BT_HallUp {
                t.Fatal("should not emit setButtonLamp for already-Inactive order")
            }
        }
    }
}

// ============================================================================
// onNetMsg tests
// ============================================================================

func TestOnNetMsg_IgnoresOwnMessage(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    msg := NetMsg{
        SenderID: myID, // same as self
    }

    _, _, _, commands := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if len(commands) != 0 {
        t.Fatalf("expected no commands for own message, got %d", len(commands))
    }
}

func TestOnNetMsg_HigherVersion_AdoptsPending_BecomesConfirmed(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    var remoteMatrix HallOrderMatrix
    remoteMatrix[2][0] = OrderMatrixEntry{Status: Pending, Version: 1}

    msg := NetMsg{
        SenderID:        "elev2",
        HallOrderMatrix: remoteMatrix,
    }

    updatedMatrix, _, _, commands := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    // Should adopt and then confirm (Pending -> Confirmed, version++)
    if updatedMatrix[2][0].Status != Confirmed {
        t.Fatalf("expected Confirmed, got %d", updatedMatrix[2][0].Status)
    }
    if updatedMatrix[2][0].Version != 2 {
        t.Fatalf("expected Version=2 after confirming, got %d", updatedMatrix[2][0].Version)
    }

    assertHasCommandType(t, commands, broadcastNetMessage)
    assertHasCommandType(t, commands, setButtonLamp)
}

func TestOnNetMsg_HigherVersion_AdoptsInactive_TurnsOffLamp(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 3}

    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    var remoteMatrix HallOrderMatrix
    remoteMatrix[1][0] = OrderMatrixEntry{Status: Inactive, Version: 5}

    msg := NetMsg{
        SenderID:        "elev2",
        HallOrderMatrix: remoteMatrix,
    }

    updatedMatrix, _, _, commands := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if updatedMatrix[1][0].Status != Inactive {
        t.Fatalf("expected Inactive, got %d", updatedMatrix[1][0].Status)
    }
    if updatedMatrix[1][0].Version != 5 {
        t.Fatalf("expected Version=5, got %d", updatedMatrix[1][0].Version)
    }

    // Verify lamp off
    for _, cmd := range commands {
        if cmd._type == setButtonLamp {
            args := cmd.value.(buttonLampArgs)
            if args.Floor == 1 && args.Button == elevio.BT_HallUp {
                if args.Value != false {
                    t.Fatal("expected lamp off for Inactive order")
                }
                return
            }
        }
    }
    t.Fatal("expected setButtonLamp command to turn off lamp")
}

func TestOnNetMsg_SameVersion_RemotePending_ConfirmsLocally(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[0][1] = OrderMatrixEntry{Status: Pending, Version: 3}

    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    var remoteMatrix HallOrderMatrix
    remoteMatrix[0][1] = OrderMatrixEntry{Status: Pending, Version: 3}

    msg := NetMsg{
        SenderID:        "elev2",
        HallOrderMatrix: remoteMatrix,
    }

    updatedMatrix, _, _, commands := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if updatedMatrix[0][1].Status != Confirmed {
        t.Fatalf("expected Confirmed, got %d", updatedMatrix[0][1].Status)
    }
    if updatedMatrix[0][1].Version != 4 {
        t.Fatalf("expected Version=4, got %d", updatedMatrix[0][1].Version)
    }

    assertHasCommandType(t, commands, broadcastNetMessage)
    assertHasCommandType(t, commands, setButtonLamp)
}

func TestOnNetMsg_SameVersion_RemoteConfirmed_AdoptsConfirmed(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[3][0] = OrderMatrixEntry{Status: Pending, Version: 2}

    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    var remoteMatrix HallOrderMatrix
    remoteMatrix[3][0] = OrderMatrixEntry{Status: Confirmed, Version: 2}

    msg := NetMsg{
        SenderID:        "elev2",
        HallOrderMatrix: remoteMatrix,
    }

    updatedMatrix, _, _, _ := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if updatedMatrix[3][0].Status != Confirmed {
        t.Fatalf("expected Confirmed, got %d", updatedMatrix[3][0].Status)
    }
}

func TestOnNetMsg_SameVersion_RemoteAssigned_AdoptsAssigned(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[1][1] = OrderMatrixEntry{Status: Confirmed, Version: 4}

    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    var remoteMatrix HallOrderMatrix
    remoteMatrix[1][1] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 4}

    msg := NetMsg{
        SenderID:        "elev2",
        HallOrderMatrix: remoteMatrix,
    }

    updatedMatrix, _, _, _ := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if updatedMatrix[1][1].Status != Assigned {
        t.Fatalf("expected Assigned, got %d", updatedMatrix[1][1].Status)
    }
    if updatedMatrix[1][1].AssignedElevator != "elev2" {
        t.Fatalf("expected AssignedElevator=elev2, got %s", updatedMatrix[1][1].AssignedElevator)
    }
}

func TestOnNetMsg_LowerVersion_NoChange(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[2][0] = OrderMatrixEntry{Status: Confirmed, Version: 10}

    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    var remoteMatrix HallOrderMatrix
    remoteMatrix[2][0] = OrderMatrixEntry{Status: Pending, Version: 5}

    msg := NetMsg{
        SenderID:        "elev2",
        HallOrderMatrix: remoteMatrix,
    }

    updatedMatrix, _, _, _ := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if updatedMatrix[2][0].Status != Confirmed {
        t.Fatalf("expected status unchanged (Confirmed), got %d", updatedMatrix[2][0].Status)
    }
    if updatedMatrix[2][0].Version != 10 {
        t.Fatalf("expected version unchanged (10), got %d", updatedMatrix[2][0].Version)
    }
}

func TestOnNetMsg_UpdatesCabCallsForSender(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{}
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    remoteCabCalls := make(CabCallsMap)
    remoteCabCalls["elev2"] = LocalCabCalls{true, false, true, false}
    remoteCabCalls["elev1"] = LocalCabCalls{}

    msg := NetMsg{
        SenderID: "elev2",
        CabCalls: remoteCabCalls,
    }

    _, updatedCabCalls, _, _ := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if !updatedCabCalls["elev2"][0] || !updatedCabCalls["elev2"][2] {
        t.Fatalf("expected elev2 cab calls at floors 0,2, got %+v", updatedCabCalls["elev2"])
    }
}

func TestOnNetMsg_PendingCabCallConfirmed_ClearsPending(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    cabCalls["elev1"] = LocalCabCalls{true, false, false, false}
    pendingCabCalls := LocalCabCalls{true, false, false, false}
    myID := ElevID("elev1")

    remoteCabCalls := make(CabCallsMap)
    remoteCabCalls["elev2"] = LocalCabCalls{}
    remoteCabCalls["elev1"] = LocalCabCalls{true, false, false, false} // remote confirms our cab call

    msg := NetMsg{
        SenderID: "elev2",
        CabCalls: remoteCabCalls,
    }

    _, _, updatedPending, commands := onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)

    if updatedPending[0] != false {
        t.Fatal("expected pending cab call at floor 0 to be cleared after confirmation")
    }

    assertHasCommandType(t, commands, setButtonLamp)
}

func TestOnNetMsg_NilCabCalls_NoPanic(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    cabCalls := make(CabCallsMap)
    var pendingCabCalls LocalCabCalls
    myID := ElevID("elev1")

    msg := NetMsg{
        SenderID: "elev2",
        CabCalls: nil,
    }

    // Should not panic
    _, _, _, _ = onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, msg)
}

// ============================================================================
// onHeartbeatTick tests
// ============================================================================

func TestOnHeartbeatTick_ReturnsBroadcast(t *testing.T) {
    commands := onHeartbeatTick()

    if len(commands) != 1 {
        t.Fatalf("expected 1 command, got %d", len(commands))
    }
    if commands[0]._type != broadcastNetMessage {
        t.Fatalf("expected broadcastNetMessage, got %d", commands[0]._type)
    }
}

// ============================================================================
// onPeerEvent tests
// ============================================================================

func TestOnPeerEvent_PeerDies_ReleasesAssignedOrders(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 5}
    hallOrderMatrix[3][1] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 3}
    hallOrderMatrix[0][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 2}

    oldPeerList := []Peer{
        {ID: "elev2", Status: Alive},
    }
    newPeerList := []Peer{
        {ID: "elev2", Status: Dead},
    }

    updatedMatrix, updatedPeers, commands := onPeerEvent(hallOrderMatrix, oldPeerList, newPeerList)

    // elev2's orders should be released to Pending
    if updatedMatrix[1][0].Status != Pending {
        t.Fatalf("expected floor 1 HallUp released to Pending, got %d", updatedMatrix[1][0].Status)
    }
    if updatedMatrix[1][0].AssignedElevator != "" {
        t.Fatalf("expected empty AssignedElevator after release, got %s", updatedMatrix[1][0].AssignedElevator)
    }
    if updatedMatrix[1][0].Version != 6 {
        t.Fatalf("expected Version=6 after release, got %d", updatedMatrix[1][0].Version)
    }

    if updatedMatrix[3][1].Status != Pending {
        t.Fatalf("expected floor 3 HallDown released to Pending, got %d", updatedMatrix[3][1].Status)
    }

    // elev1's order should be untouched
    if updatedMatrix[0][0].Status != Assigned {
        t.Fatalf("expected elev1's order to remain Assigned, got %d", updatedMatrix[0][0].Status)
    }
    if updatedMatrix[0][0].AssignedElevator != "elev1" {
        t.Fatalf("expected elev1's AssignedElevator unchanged, got %s", updatedMatrix[0][0].AssignedElevator)
    }

    if len(updatedPeers) != 1 || updatedPeers[0].Status != Dead {
        t.Fatalf("expected updated peer list with dead peer, got %+v", updatedPeers)
    }

    assertHasCommandType(t, commands, broadcastNetMessage)
}

func TestOnPeerEvent_PeerAlreadyDead_NoRelease(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[2][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 5}

    oldPeerList := []Peer{
        {ID: "elev2", Status: Dead},
    }
    newPeerList := []Peer{
        {ID: "elev2", Status: Dead},
    }

    updatedMatrix, _, _ := onPeerEvent(hallOrderMatrix, oldPeerList, newPeerList)

    // Should NOT release because peer was already Dead before
    if updatedMatrix[2][0].Status != Assigned {
        t.Fatalf("expected order to remain Assigned (peer was already dead), got %d", updatedMatrix[2][0].Status)
    }
    if updatedMatrix[2][0].Version != 5 {
        t.Fatalf("expected version unchanged, got %d", updatedMatrix[2][0].Version)
    }
}

func TestOnPeerEvent_PeerComesAlive_NoRelease(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[1][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 3}

    oldPeerList := []Peer{
        {ID: "elev2", Status: Dead},
    }
    newPeerList := []Peer{
        {ID: "elev2", Status: Alive},
    }

    updatedMatrix, _, _ := onPeerEvent(hallOrderMatrix, oldPeerList, newPeerList)

    if updatedMatrix[1][0].Status != Assigned {
        t.Fatalf("expected order to remain Assigned when peer comes alive, got %d", updatedMatrix[1][0].Status)
    }
}

func TestOnPeerEvent_NewPeerAppears_NoRelease(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[0][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 2}

    oldPeerList := []Peer{}
    newPeerList := []Peer{
        {ID: "elev3", Status: Alive},
    }

    updatedMatrix, _, _ := onPeerEvent(hallOrderMatrix, oldPeerList, newPeerList)

    if updatedMatrix[0][0].Status != Assigned {
        t.Fatalf("expected existing orders unchanged, got %d", updatedMatrix[0][0].Status)
    }
}

// ============================================================================
// claimOrder tests
// ============================================================================

func TestClaimOrder_SetsAssignedAndBroadcasts(t *testing.T) {
    var hallOrderMatrix HallOrderMatrix
    hallOrderMatrix[2][0] = OrderMatrixEntry{Status: Confirmed, Version: 4}

    myID := ElevID("elev1")
    order := OrderLocation{
        Floor:  2,
        Button: elevio.BT_HallUp,
        Entry:  hallOrderMatrix[2][0],
    }

    updatedMatrix, commands := claimOrder(hallOrderMatrix, myID, order)

    if updatedMatrix[2][0].Status != Assigned {
        t.Fatalf("expected Assigned, got %d", updatedMatrix[2][0].Status)
    }
    if updatedMatrix[2][0].AssignedElevator != myID {
        t.Fatalf("expected AssignedElevator=%s, got %s", myID, updatedMatrix[2][0].AssignedElevator)
    }
    if updatedMatrix[2][0].Version != 5 {
        t.Fatalf("expected Version=5, got %d", updatedMatrix[2][0].Version)
    }

    assertHasCommandType(t, commands, sendOrderToLocal)
    assertHasCommandType(t, commands, setButtonLamp)
    assertHasCommandType(t, commands, broadcastNetMessage)

    // Verify the sendOrderToLocal has correct event
    for _, cmd := range commands {
        if cmd._type == sendOrderToLocal {
            evt := cmd.value.(elevio.ButtonEvent)
            if evt.Floor != 2 || evt.Button != elevio.BT_HallUp {
                t.Fatalf("expected floor=2 BT_HallUp, got floor=%d button=%d", evt.Floor, evt.Button)
            }
        }
    }
}

// ============================================================================
// releaseOrdersForPeer tests
// ============================================================================

func TestReleaseOrdersForPeer_OnlyReleasesTargetPeer(t *testing.T) {
    var h HallOrderMatrix
    h[0][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 3}
    h[1][1] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev1", Version: 5}
    h[2][0] = OrderMatrixEntry{Status: Pending, Version: 2}
    h[3][0] = OrderMatrixEntry{Status: Assigned, AssignedElevator: "elev2", Version: 7}

    result := releaseOrdersForPeer(h, "elev2")

    if result[0][0].Status != Pending || result[0][0].AssignedElevator != "" || result[0][0].Version != 4 {
        t.Fatalf("expected floor 0 released: %+v", result[0][0])
    }
    if result[1][1].Status != Assigned || result[1][1].AssignedElevator != "elev1" {
        t.Fatalf("expected elev1's order unchanged: %+v", result[1][1])
    }
    if result[2][0].Status != Pending || result[2][0].Version != 2 {
        t.Fatalf("expected already-Pending order unchanged: %+v", result[2][0])
    }
    if result[3][0].Status != Pending || result[3][0].Version != 8 {
        t.Fatalf("expected floor 3 released: %+v", result[3][0])
    }
}

// ============================================================================
// findPeerStatus tests
// ============================================================================

func TestFindPeerStatus_Found(t *testing.T) {
    peerList := []Peer{
        {ID: "elev1", Status: Alive},
        {ID: "elev2", Status: Dead},
    }

    if findPeerStatus(peerList, "elev1") != Alive {
        t.Fatal("expected Alive for elev1")
    }
    if findPeerStatus(peerList, "elev2") != Dead {
        t.Fatal("expected Dead for elev2")
    }
}

func TestFindPeerStatus_NotFound(t *testing.T) {
    peerList := []Peer{
        {ID: "elev1", Status: Alive},
    }

    status := findPeerStatus(peerList, "elev99")
    if status != PeerStatus(-1) {
        t.Fatalf("expected -1 for unknown peer, got %d", status)
    }
}

// ============================================================================
// Helpers
// ============================================================================

func assertEmpty(t *testing.T, tx chan NetMsg, light chan elevio.DriverCommand) {
    t.Helper()
    assertEmptyTx(t, tx)
    assertEmptyLight(t, light)
}

func assertEmptyLocal(t *testing.T, ch chan elevio.ButtonEvent) {
    t.Helper()
    select {
    case v := <-ch:
        t.Fatalf("unexpected value on localOrderChan: %+v", v)
    default:
    }
}

func assertEmptyTx(t *testing.T, ch chan NetMsg) {
    t.Helper()
    select {
    case v := <-ch:
        t.Fatalf("unexpected value on tx: %+v", v)
    default:
    }
}

func assertEmptyLight(t *testing.T, ch chan elevio.DriverCommand) {
    t.Helper()
    select {
    case v := <-ch:
        t.Fatalf("unexpected value on lightCommandChan: %+v", v)
    default:
    }
}

func assertHasCommandType(t *testing.T, commands []command, expected commandType) {
    t.Helper()
    for _, cmd := range commands {
        if cmd._type == expected {
            return
        }
    }
    t.Fatalf("expected command type %d in commands, but not found", expected)
}