## PeerMonitor

### Responsibility
- Tracks liveness of peers based on heartbeats.
- Detects timeouts and emits `DeadPeers` events.
- Optionally emits `PeerUp` when peers reappear.

PeerMonitor is purely about **presence**; it does not assign orders.

### Owns (mutable state)
- `Peer-> struct wih ID, Status, CabcallBackups and Lastseen `
- `PeerStates-> Peers, map`
- `Peerupdate -> Peers`
- `PeerConfig-> Timeout time.Duration, TickPeriod time.Duration`
- `PeerInputs-> Heartbeat(recieves Rx from TrUDP), Tick`
-`PeerOutpus(sends updated PeerUpdate to Ordersync)`

### Run() interface

#### Inputs (receive-only)
- `rx <-chan NetMsg`
- `tick <-chan time.Time` *(optional)*  
  If ticks are provided externally; otherwise PeerMonitor runs its own ticker. (Mest sannsynlig egen)

#### Outputs (send-only)
- `PeerUpdate(Peer list)`


### Functional core vs Imperative shell

#### Core (testable)
- `Update(lastSeen, heartbeat) -> newLastSeen`
- `DetectDead(lastSeen, now, timeout) -> deadIDs`

#### Shell
- select-loop consuming heartbeats
- periodic tick to run detection