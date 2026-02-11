## Node (Wiring + Supervision)

### Responsibility
- Creates all channels and connects modules (the "wiring diagram").
- Routes raw button events into Cab vs Hall streams.
- Starts all module goroutines via `Run(...)`.
- Hosts supervision policy (restart on panic/exit).
- Owns process-level lifecycle (context cancellation, shutdown).

---

### Owns (mutable state)
- Channel instances and their buffering.
- Supervisor state (restart counters, backoff timers).
- Optional: node-wide configuration.

---

### Suggested wiring (high-level)
- ElevIO -> ButtonRouter -> (CabButton -> LocalSingleElevator) + (HallButton -> OrderSync)
- ElevIO -> LocalSingleElevator (Floor/Obstruction)
- LocalSingleElevator -> ElevIO (DriverCmd)
- LocalSingleElevator -> OrderSync (StateOut, Cleared)
- OrderSync <-> TransportUDP (NetTx, NetRx)
- OrderSync -> LocalSingleElevator (AssignedHall)
- PeerMonitor -> OrderSync (DeadPeers)
- OrderSync -> (Heartbeat NetMsg) -> TransportUDP -> PeerMonitor (if using heartbeats)

---

### Supervision policy
- A **supervisor** restarts a child goroutine if it panics or returns unexpectedly.
- A **watchdog** detects lack-of-progress/timeouts and triggers recovery (optional).
- Prefer:
  - transport/elevio errors become events
  - supervisor only handles hard crashes

---

### Suggested files
