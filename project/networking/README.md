## Networking - UDP Broadcast Transport

### Overall Responsibility

Provides the **physical network communication layer** for the distributed elevator system.

**Single concern**: Send and receive UDP broadcast messages.

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

### Configuration

From `config/config.go`:
- `BROADCAST_ADDRESS`: IP address for broadcast (e.g., 255.255.255.255 for local subnet)
- `BROADCAST_PORT`: UDP port for Elevator network (e.g., 15123)