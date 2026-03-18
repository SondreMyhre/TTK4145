package config

import(
	"time"
)

const (
    N_FLOORS                = 4
    BROADCAST_PORT          = 50000
	BROADCAST_ADDRESS       = "10.100.23.255"
    POLL_RATE               = 20 * time.Millisecond
	MOTORTIMEOUT            = 3500 * time.Millisecond
    PEER_TIMEOUT            = 5 * time.Second
    PEER_TICK_INTERVAL      = 50 * time.Millisecond
    HEARTBEAT_TICK_INTERVAL = 50 * time.Millisecond
	NETMSG_TICK_INTERVAL    = 50 * time.Millisecond
)