## PeerMonitor

### Responsibility
- Tracks liveness of peers based on heartbeats.
- Detects timeouts and emits `DeadPeers` events.
- Optionally emits `PeerUp` when peers reappear.

PeerMonitor is purely about **presence**; it does not assign orders.

### Owns (mutable state)
- `lastSeen map[ElevID]time.Time`
- `timeout time.Duration`
- `peers []Peer` liste med informasjon over peers??

### Run() interface

#### Inputs (receive-only)
- `rx <-chan NetMsg`
- `tick <-chan time.Time` *(optional)*  
  If ticks are provided externally; otherwise PeerMonitor runs its own ticker. (Mest sannsynlig egen)

#### Outputs (send-only)
- `deadPeers chan<- []Peer`
- `peerUp chan<- Peer` *(optional)*
- `peerList chan<- []Peer` *(optional)*


### Functional core vs Imperative shell

#### Core (testable)
- `Update(lastSeen, heartbeat) -> newLastSeen`
- `DetectDead(lastSeen, now, timeout) -> deadIDs`

#### Shell
- select-loop consuming heartbeats
- periodic tick to run detection