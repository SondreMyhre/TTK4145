## OrderSync

### Responsibility
- Distributed synchronization and assignment of **hall orders** across all elevators.
- Maintains global hall-order state and who is assigned to what, amd control lights
- Receives via channels from ElevIO, PeerMonitor, TransportUDP and LocalSingle:
  - local hall button presses (from ElevIO)
  - local elevator state updates (from LocalSingle)
  - cleared orders (from LocalSingle)
  - network messages (from TransportUDP)
  - dead peer notifications (from PeerMonitor)
- Produces sends via channels to TransportUDP and LocalSingle:
  - assignments to the local elevator
  - outbound network messages

### Owns (mutable state)
- `hallOrderMatrix [N_FLOORS][N_HALL]OrderMatrixEntry` 
- `localState LocalState`
- `cabCalls CabCallsMap`
- `peerList []Peer`
- `pendingCabCalls [N_FLOOR]bool`

#### Inputs (receive-only)
- `buttonChan <-chan elevio.ButtonEvent`
- `localStateChan <-chan localsingle.ElevatorState`
- `clearedOrdersChan <-chan []localsingle.Order`
- `rx <-chan NetMsg`
- `peerEventChan <-chan []Peer`

#### Outputs (send-only)
- `localOrderChan chan<- elevio.ButtonEvent`
- `tx chan<- NetMsg`
- `lightCommandChan <-chan DriverCommand`