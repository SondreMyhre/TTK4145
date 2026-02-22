## Node/main (Wiring + Supervision)

### Responsibility
- Creates all channels and connects modules (the "wiring diagram").
- Routes raw button events into Cab vs Hall streams.
- Starts all module goroutines via `Run(...)`.
- Hosts supervision policy (restart on panic/exit). (watchdog with process pair)
- Owns process-level lifecycle (context cancellation, shutdown).

### Owns (mutable state)
- Channel instances and their buffering.
- Supervisor state (restart counters, backoff timers). (Maybe not?)
- Optional: node-wide configuration.


### Suggested wiring (high-level)
- ElevIO -> OrderSync
- ElevIO -> LocalSingleElevator (Floor/Obstruction)
- LocalSingleElevator -> ElevIO (driverCommandChan)
- LocalSingleElevator -> OrderSync (localStateChan, clearedOrdersChan)
- OrderSync <-> TransportUDP (tx, rx)	// Kan vurdere egen heartbeatChan
- OrderSync -> LocalSingleElevator (localOrderChan)
- PeerMonitor -> OrderSync (deadPeersChan)

### Supervision policy
- A **supervisor** restarts a child goroutine if it panics or returns unexpectedly.
- A **watchdog** detects lack-of-progress/timeouts and triggers recovery (optional).
- Prefer:
  - transport/elevio errors become events
  - supervisor only handles hard crashes

### Template for modules AI-KODE så ta med en klype salt

\module
  api.go
  types.go
  core.go (kun splitt i flere om gir mening for eksempel core_fsm.go og core_requests.go )
  shell.go
  (eventuelt tester)

api.go:

	package module

	import (
		"context"
		"time"
	)

	type Config struct {
		TickPeriod time.Duration // 0 => no internal ticker
	}

	// Inputs: kun receive-only
	type Inputs struct {
		// Stil C: flere små kanaler
		// Navnene bør være domene-ord (ikke generiske)
		A <-chan MsgA
		B <-chan MsgB
		// Optional external tick channel, if you don't want internal ticker
		Tick <-chan time.Time
	}

	// Outputs: kun send-only
	type Outputs struct {
		X chan<- OutX
		Y chan<- OutY
	}

	// Run = shell entrypoint (one goroutine owns all state in this module)
	func Run(ctx context.Context, cfg Config, in Inputs, out Outputs) {
		run(ctx, cfg, in, out)
	}


types.go:

	package module

	// --- Inbound messages (events) ---

	type MsgA struct {
		// small, intention-revealing payload
		ID int
	}

	type MsgB struct {
		Flag bool
	}

	// --- Outbound messages ---

	type OutX struct {
		Value int
	}

	type OutY struct {
		Text string
	}

	// --- Internal state: ONLY owned/mutated by this module's goroutine ---

	type state struct {
		counter int
		flag    bool
	}

	func initState() state {
		return state{}
	}

core.go:

	package module

	type cmdKind int

	const (
		cmdSendX cmdKind = iota
		cmdSendY
	)

	type cmd struct {
		kind cmdKind
		x    OutX
		y    OutY
	}

	func onA(s *state, m MsgA) []cmd {
		s.counter += m.ID
		return []cmd{{kind: cmdSendX, x: OutX{Value: s.counter}}}
	}

	func onB(s *state, m MsgB) []cmd {
		s.flag = m.Flag
		if s.flag {
			return []cmd{{kind: cmdSendY, y: OutY{Text: "flag enabled"}}}
		}
		return nil
	}

	func onTick(s *state) []cmd {
		// optional periodic logic
		return nil
	}


shell.go:

	package module

	import (
		"context"
		"time"
	)

	func run(ctx context.Context, cfg Config, in Inputs, out Outputs) {
		s := initState()

		var tickerC <-chan time.Time
		var ticker *time.Ticker
		if cfg.TickPeriod > 0 {
			ticker = time.NewTicker(cfg.TickPeriod)
			tickerC = ticker.C
			defer ticker.Stop()
		} else if in.Tick != nil {
			tickerC = in.Tick
		}

		apply := func(cmds []cmd) {
			for _, c := range cmds {
				switch c.kind {
				case cmdSendX:
					out.X <- c.x
				case cmdSendY:
					out.Y <- c.y
				}
			}
		}

		for {
			select {
			case <-ctx.Done():
				return

			case m := <-in.A:
				apply(onA(&s, m))

			case m := <-in.B:
				apply(onB(&s, m))

			case <-tickerC:
				apply(onTick(&s))
			}
		}
	}
