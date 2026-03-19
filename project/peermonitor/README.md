## PeerMonitor - Failure Detection via Heartbeats

### Overall Responsibility

Detects when peer elevators become unavailable and notifies OrderSync so it can reassign their orders.

**Mechanism**: Heartbeat with timeout detection
- Each elevator sends heartbeat periodically
- PeerMonitor watches for these heartbeats
- If heartbeat stops arriving, peer is declared dead
- When heartbeats resume, peer is declared alive again

---

### How It Works

**State maintained per peer**:
```go
type Peer struct {
  ID         ElevID                            // Unique peer identifier
  PeerStatus PeerStatus                        // StatusAlive or StatusDead
  state      elevatorcontroller.ElevatorState // Last seen state
}
```

**Tick-based detection**:
- Timer ticks periodically (every PEER_TICK_INTERVAL)
- Check all peers for timeout
- Emit events when status changes
- OrderSync hears the events and can reassign orders

---

### Information Flow - Adding Self and Detecting Others

**Heartbeat outbound** (broadcast to all peers):
```
ElevatorController sends on localStateChan
    ↓
OrderSync receives elevator state
    ↓
Assigner sends assignment decision
    ↓
OrderSync constructs NetMsg (order update)
    ↓
PeerMonitor sends heartbeat via peerMonitorTx
    ↓ peerMonitorTx
Network broadcasts
    ↓
(other peers receive but this one's heartbeat goes to all)
```

**Heartbeat inbound** (received from peer):
```
Network receives UDP heartbeat from peer
    ↓ peerMonitorRx
PeerMonitor receives heartbeat
    ↓
Update lastSeen timestamp for that peer
    ↓
Check if any peers have timed out
    ↓
If peer just died (was alive, now timeout):
    emit PeerUpdate{peer, StatusDead}
    ↓ peerEventChan
OrderSync receives
    ↓
OrderSync reassigns orders from dead peer
```

### Inputs (receive-only channels)

- `peerMonitorRx <-chan HeartBeat` (from Network)
  - Heartbeat messages from peer elevators
  - Contains: sender ID, sender state snapshot
  - Updates: lastSeenTime for that peer

### Outputs (send-only channels)

- `peerEventChan chan<- []PeerUpdate` (to OrderSync)
  - Notifications when peers die or resurrect
  - Example: {ElevID("elevator_2"), StatusDead}
  - OrderSync uses this to trigger reassignment

- `peerMonitorTx chan<- HeartBeat` (to Network)
  - Broadcast heartbeat to all peers
  - Tells others "I'm alive"
  - Contains: my ID, my current state

---

### Configuration

From `config/config.go`:
- `PEER_TIMEOUT`: How long to wait before declaring peer dead (e.g., 6 seconds)
- `PEER_TICK_INTERVAL`: How often to check for timeouts (e.g., 100ms)
- `HEARTBEAT_TICK_INTERVAL`: How often to send heartbeat (e.g., 100ms)

