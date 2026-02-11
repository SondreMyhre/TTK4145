# ordersync/README.md

## OrderSync

### Responsibility
- Distributed synchronization and assignment of **hall orders** across all elevators.
- Maintains global hall-order state and who is assigned to what.
- Receives:
  - local hall button presses
  - local elevator state updates
  - cleared orders
  - network messages
  - dead peer notifications
- Produces:
  - assignments to the local elevator
  - outbound network messages (heartbeat/claim/assign/done)

OrderSync does **not** control motors/lamps directly.

---

### Owns (mutable state)
- Hall order state:
  - `hallOrderMatrix [N_FLOORS][N_HALL_BTNS]HallOrderState` *(or equivalent)*
  - `assignedTo map[HallOrder]ElevID`
- World view:
  - `lastKnownStates map[ElevID]ElevatorState`
- Protocol/bookkeeping:
  - `seq counters`, dedup cache *(recommended)*

---

### Run() interface

#### Inputs (receive-only)
- `HallButton <-chan HallOrder`
- `LocalState <-chan ElevatorState`
- `Cleared <-chan ClearedOrders`
- `NetRx <-chan NetMsg`
- `DeadPeers <-chan []ElevID`
- `Tick <-chan time.Time` *(optional, or internal ticker)*

#### Outputs (send-only)
- `AssignToLocal chan<- HallAssignment`
- `NetTx chan<- NetMsg`

Optional:
- `Err chan<- error` (if you want OrderSync to surface issues upward)

---

### Functional core vs Imperative shell

#### Core (testable)
- `Step(state, event) -> (newState, effects)`
- No IO, no timers; deterministic transitions.

Effects are value objects such as:
- `SendNet(NetMsg)`
- `AssignLocal(HallAssignment)`
- `RecomputeAssignments`
- `ReleaseDeadPeerOrders(peerID)`
- `EmitDone(order)`
- `BuildHeartbeat`

#### Shell
- select-loop consuming all inputs
- periodic tick to generate heartbeats / recompute
- `apply(effects)` by sending on channels

---

### Suggested files







DETTE ER SKREVET AV AI, KUN BRUK TIL EKSEMPEL TIL FCIS

Functional core example:  
    
    func Step(s State, e Event) (State, []Effect)

Imperative shell example:

    func Run(
    ctx context.Context,
    in <-chan Event,
    out chan<- Msg,
    driver chan<- Cmd,
    // evt. andre avhengigheter
) {
    s := InitState()

    // timers eies her
    var doorTimerC <-chan time.Time

    apply := func(effects []Effect) {
        for _, ef := range effects {
            switch ef.Type {
            case EffectSendMsg:
                out <- ef.Msg
            case EffectDriverCmd:
                driver <- ef.Cmd
            case EffectStartTimer:
                // start/reset timer, sett doorTimerC = timer.C
            }
        }
    }

    for {
        select {
        case <-ctx.Done():
            return

        case ev := <-in:
            s2, effects := Step(s, ev)
            s = s2
            apply(effects)

        case <-doorTimerC:
            s2, effects := Step(s, Event{Type: DoorTimeout})
            s = s2
            apply(effects)
        }
    }
}


DENNE ER GANSKE RYDDIG Å SE PÅ:

https://github.com/angrycompany16/TTK4145-sanntidslab/blob/main/elevalgo/elevator_process.go