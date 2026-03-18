package config

import(
	"time"
)

const (
    N_FLOORS                = 4
    BROADCAST_PORT          = 50000
    POLL_RATE               = 20 * time.Millisecond
    PEER_TIMEOUT            = 5 * time.Second
    PEER_TICK_INTERVAL      = 50 * time.Millisecond
    HEARTBEAT_TICK_INTERVAL = 1 * time.Millisecond
	MOTORTIMEOUT            = 3500 * time.Millisecond
	SERVERADDR              = "localhost:15657"
)