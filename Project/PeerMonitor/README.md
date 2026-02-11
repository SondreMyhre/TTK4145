# peermonitor/README.md

## PeerMonitor

### Responsibility
- Tracks liveness of peers based on heartbeats.
- Detects timeouts and emits `DeadPeers` events.
- Optionally emits `PeerUp` when peers reappear.

PeerMonitor is purely about **presence**; it does not assign orders.

---

### Owns (mutable state)
- `lastSeen map[ElevID]time.Time`
- `timeout time.Duration`

---

### Run() interface

#### Inputs (receive-only)
- `Heartbeat <-chan HeartbeatMsg`
- `Tick <-chan time.Time` *(optional)*  
  If ticks are provided externally; otherwise PeerMonitor runs its own ticker.

#### Outputs (send-only)
- `Dead chan<- []ElevID`
- `PeerUp chan<- ElevID` *(optional)*
- `PeerList chan<- []ElevID` *(optional)*

---

### Functional core vs Imperative shell

#### Core (testable)
- `Update(lastSeen, heartbeat) -> newLastSeen`
- `DetectDead(lastSeen, now, timeout) -> deadIDs`

#### Shell
- select-loop consuming heartbeats
- periodic tick to run detection

---

### Suggested files
