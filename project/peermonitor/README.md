## PeerMonitor - Failure Detection via Heartbeats

### Overall Responsibility

Detects when peer elevators become unavailable and notifies OrderSync so it can reassign their orders.

**Single concern**: Monitors LIVENESS only - whether a peer is alive or dead. Does NOT manage orders or state.

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

**Main algorithm**:
```
For each peer in known list:
  if now - lastHeartbeatTime > PEER_TIMEOUT:
    if status was Alive:
      status := Dead
      emit PeerUpdate{peer, Dead}
  else:
    if status was Dead:
      status := Alive
      emit PeerUpdate{peer, Alive}
```

**Tick-based detection**:
- Timer ticks periodically (every PEER_TICK_INTERVAL)
- Check all peers for timeout
- Emit events when status changes
- OrderSync hears the events and can reassign orders

---

### State Management - Local Peer List

**Owns (mutable state)**:
- `peerList []Peer` - Last known state of each peer
  - ID, Status, Last seen heartbeat time
  - State snapshot (floor, direction, behavior)

**Updates from**:
- Heartbeat messages (contains peer state snapshots)
- Timeout detection (sets status to dead)
- Peer rediscovery (sets status back to alive)

**Does NOT own**:
- Global order matrix (that's OrderSync's job)
- Hall request state (OrderSync)
- Elevator positions (only for heartbeat metadata)

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

**Timeline example** (4 elevators, peer 2 fails):
```
T=0s:   All elevators alive
T=5s:   Elevator 2 dies (network partition)
        Peer 2 stops sending heartbeats
        
T=7s:   (Peer 2 timeout = 6s) PeerMonitor 1 detects timeout
        Emits: PeerUpdate{Elev2, StatusDead}
        OrderSync 1 receives → reassigns orders from Elev2
        
T=8s:   PeerMonitor 3 detects timeout
T=9s:   PeerMonitor 4 detects timeout
        All three emit same event
        All three reassign Elev2's orders independently
        All compute same reassignments (deterministic)

T=20s:  Elevator 2 recovers from partition
        Starts sending heartbeats again
        
T=21s:  PeerMonitor 1 detects heartbeat resumed
        Emits: PeerUpdate{Elev2, StatusAlive}
        OrderSync refreshes Elev2's view
```

---

### Design Principles

**1. Simplicity**: Does ONE thing - detect timeouts
- No state management complexity
- No ordering/causality issues
- Easy to reason about failures

**2. Failure mode**: Missing heartbeat = dead
- Fault-tolerant: any message loss triggers timeout
- Eventually recovers: when heartbeats resume
- Safe: false positives (declaring alive -> dead) are handled
  - Temporary re-assignment not a problem
  - Eventually correct when heartbeat returns

**3. Symmetric**: All peers run identical timeout logic
- No master peer
- Each peer independently detects failures
- Each computes remediation independently (via OrderSync)

**4. Integration point**: Changes only peerList and peerEventChan
- Doesn't touch order state
- Doesn't touch elevator control
- OrderSync decides what to do with the failure notification

---

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

---

### Completeness Checklist

✓ Tracks all known peer elevators
✓ Detects when peer stops sending heartbeats  
✓ Announces peer death to OrderSync
✓ Detects when peer returns (heartbeat resumes)
✓ Announces peer recovery to OrderSync
✓ Handles edge case: never seen peer before
✓ Handles edge case: peer state corrupted

---

### Testing Considerations

To test PeerMonitor:
1. Feed heartbeats on peerMonitorRx
2. Simulate timeout by not feeding heartbeat
3. Check peerEventChan for death announcement
4. Resume heartbeat
5. Check peerEventChan for alive announcement

Simple, deterministic behavior with predictable state transitions.
