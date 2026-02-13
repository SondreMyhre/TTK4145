## TransportUDP

### Responsibility
- Owns UDP socket IO.
- Encodes/decodes `NetMsg`.
- Receives network messages and emits them on `Rx`.
- Accepts outbound messages on `Tx` and sends them over UDP.

TransportUDP contains **no domain logic** (no assignments, no scheduling).

### Owns (mutable state)
- UDP socket/conn
- peer/broadcast configuration
- IO buffers

### Run() interface

#### Inputs (receive-only)
- `Tx <-chan NetMsg`  
  Outgoing network messages from OrderSync (maybe not PeerMonitor?).
<!-- - `Peers <-chan []PeerAddr` *(optional)* This will most likely come trengs internally in TransportUDP and not via channel 
  If peers are updated dynamically. -->

#### Outputs (send-only)
- `Rx chan<- NetMsg`  
  Incoming messages to OrderSync and PeerMonitor.
<!-- - `Err chan<- error` *(recommended)*  
  Network/codec errors as events. -->

### Functional core vs Imperative shell

#### Core (small, testable)
- `Encode(NetMsg) []byte`
- `Decode([]byte) (NetMsg, error)`

#### Shell
- socket recv loop -> decode -> `Rx`
- socket send loop <- `Tx` -> encode -> send