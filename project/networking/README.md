## Networking - UDP Broadcast Transport

### Overall Responsibility

Provides the **physical network communication layer** for the distributed elevator system.

**Single concern**: Send and receive UDP broadcast messages. That's it.

**Critical principle**: Networking contains ZERO domain logic.
- No assignments
- No scheduling  
- No order acknowledgment
- No reliability protocols
- Just: UDP in ↔ channel out, channel in ↔ UDP out

This clean separation makes it easy to test, replace, or debug networking without touching domain logic.

---

### How It Works

Two main flows:

**Outbound** (OrderSync & PeerMonitor → Broadcast):
```go
select {
case msg := <-tx:
  encoded := Encode(msg)
  sendUDP(encoded, BROADCAST_ADDRESS:BROADCAST_PORT)
}
```

**Inbound** (Receive → OrderSync & PeerMonitor):
```go
for {
  data := recvUDP()
  msg := Decode(data)
  
  // Route to both OrderSync and PeerMonitor
  orderSyncRx <- msg
  peerMonitorRx <- msg
}
```

**Simple and symmetric**: All peers receive the same broadcast. No unicast, no directed messages.

---

### Network Topology

```
┌─────────────────────────────┐
│   UDP Broadcast Address     │
│   (e.g., 255.255.255.255)  │
└──────────────┬──────────────┘
               │
    ┌──────────┼──────────────┬──────────────┐
    │          │              │              │
    ↓          ↓              ↓              ↓
  Elev 1     Elev 2        Elev 3        Elev 4
(Port X)    (Port X)      (Port X)      (Port X)
```

All elevators listen on the SAME broadcast address and port simultaneously. Messages sent by one are received by all (including the sender).

---

### Testing Transport

Networking is well-isolated for testing:

**Functional core (pure, testable)**:
```go
func Encode(msg NetMsg) []byte
  Input:  Domain message (orders, state, heartbeat)
  Output: JSON bytes for network transmission
  Logic:  Message structure → JSON serialization

func Decode([]byte) (NetMsg, error)
  Input:  Received JSON bytes
  Output: Parsed domain message
  Logic:  JSON deserialization → Message structure
```

**Imperative shell (with side effects)**:
```go
func Run(tx, peerTx chan, rx, peerRx chan) {
  select {
    case msg := <-tx:
      encoded := Encode(msg)
      socket.Write(encoded)        // side effect: network
    
    case msg := <-peerTx:
      encoded := Encode(msg)
      socket.Write(encoded)        // side effect: network
    
    // Receiving happens in separate goroutine
  }
}

func receiveLoop(rx, peerRx chan) {
  for {
    data := socket.Read()           // side effect: blocking I/O
    msg := Decode(data)
    rx <- msg                       // send to OrderSync
    peerRx <- msg                   // send to PeerMonitor (if needed)
  }
}
```

**Testing strategy**:
1. Test Encode/Decode independently with unit tests
   - No network needed
   - Just JSON in/out
   - 100% deterministic

2. Test Run() with mock channels
   - Send message on tx
   - Intercepted encoded form would go to network
   - For testing: mock socket reader/writer
   - Verify right format sent

3. Integration test with real network
   - Two processes on localhost
   - Send message from one's tx
   - Receive on other's rx
   - Verify round-trip

---

### Information Content

**What messages contain**:

Sent by **OrderSync** (via orderSyncTx):
- Hall order matrix updates
- Cab call state
- This elevator's current state snapshot
- What needs to change from last broadcast

Sent by **PeerMonitor** (via peerMonitorTx):
- Heartbeat (simple "I'm alive" with embedded state)

Received by **OrderSync** (via orderSyncRx):
- Same information from all other elevators
- Used to build the global worldview

Received by **PeerMonitor** (via peerMonitorRx):
- Heartbeats from other elevators
- Used to detect timeouts

**Broadcast is symmetric**: Every message sent is received by everyone (including sender). This is actually desirable because:
- Simplifies logic: same code path for local and peer state
- More reliable: loopback confirms own messages  
- Debugging: can see what you sent

---

### Network Reliability

**What we DON'T have**:
- TCP (which guarantees order and delivery)
- Acknowledgment/feedback
- Retransmission protocols
- Guaranteed delivery

**What we DO have**:
- Broadcast to multiple recipients simultaneously
- Low latency (no handshakes)
- Best-effort delivery

**So what happens if a message is lost?**

The system is **eventually consistent**:
```
NetMsg 1 (order at floor 2) - LOST
NetMsg 2 (order at floor 3) - RECEIVED
NetMsg 3 (order at floor 2 cleared) - RECEIVED
NetMsg 4 (order at floor 2) - RECEIVED (retransmit)

Net result: Order is known, despite loss in the middle
```

Why this works:
- OrderSync broadcasts the ENTIRE state periodically
- Lost messages are retransmitted
- Old information is harmless (triggers reassignment, not worse)
- System converges to consistency

**Message loss strategy**:
- OrderSync sends state updates frequently (not one-shot)
- If one message lost, next message has updated info
- PeerMonitor sends heartbeats frequently
- If one heartbeat lost, next one allows recovery
- Periodic broadcasts mean no message is critical

---

### Configuration

From `config/config.go`:
- `BROADCAST_ADDRESS`: IP address for broadcast (e.g., 255.255.255.255 for local subnet)
- `BROADCAST_PORT`: UDP port for Elevator network (e.g., 15123)
- `NETMSG_TICK_INTERVAL`: How often OrderSync broadcasts (e.g., 100ms)
- `HEARTBEAT_TICK_INTERVAL`: How often PeerMonitor sends heartbeats (e.g., 50ms)

---

### Inputs (receive-only channels)

- `tx <-chan NetMsg` (from OrderSync)
  - Order state updates to broadcast
  - One message per significant change

- `peerTx <-chan HeartBeat` (from PeerMonitor)
  - Heartbeat messages to broadcast
  - Nearly periodic (every HEARTBEAT_TICK_INTERVAL)

### Outputs (send-only channels)

- `rx chan<- NetMsg` (to OrderSync)
  - Incoming order state messages from peer elevators
  - Can include own messages (loopback)

- `peerRx chan<- HeartBeat` (to PeerMonitor)
  - Incoming heartbeat messages from all elevators
  - Can include own heartbeat (loopback)

---

### Design Decisions

**1. Broadcast only**: No directed unicast
- Every message to everyone
- Simpler protocol
- Works with network topology changes
- Minimal infrastructure requirements

**2. Unstructured**: No message IDs or sequencing
- Stateless in transport
- Receiver decides if message is "new" or "old"
- Domain logic handles idempotency

**3. Separation of concerns**: Two types of messages
- OrderSync messages: order state synchronization
- PeerMonitor messages: endpoint liveness detection
- Could be same channel, but cleaner as separate

**4. Loopback**: Receiver sees own sent messages
- Not necessary but not harmful
- Simplifies protocol (no special case for self)
- Can actually be useful for debugging

---

### Common Pitfalls to Avoid

❌ **DON'T** add ordering/sequencing to Networking
  - That's domain logic
  - OrderSync will decide if message is old/new
  
❌ **DON'T** add acknowledgment protocols
  - Networking just broadcasts
  - If reliability needed, domain layer handles it (via periodic resend)
  
❌ **DON'T** cache messages in Networking
  - Just route in/out
  - Let domains store what they need
  
❌ **DON'T** interpret message content
  - Just encode/decode
  - Don't validate semantics
  
✓ **DO** keep Networking completely domain-agnostic
✓ **DO** make Encode/Decode pure functions  
✓ **DO** centralize network config in main.go
✓ **DO** log network errors for debugging
