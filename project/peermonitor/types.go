package peermonitor

import (
	ordersync "project/ordersync"
	"time"
	config "project/config"
)

const (
	PEER_TIMEOUT             = config.PEER_TIMEOUT
	PEER_TICK_INTERVAL       = config.PEER_TICK_INTERVAL
	HEARTBEAT_TICK_INTERVAL  = config.HEARTBEAT_TICK_INTERVAL

	StatusAlive = ordersync.StatusAlive
	StatusDead  = ordersync.StatusDead
)

type ElevID = ordersync.ElevID
type PeerStatus = ordersync.PeerStatus
type PeerUpdate = ordersync.PeerUpdate
type PeerMsg = []PeerUpdate


type HeartBeat struct {
	SenderID ElevID
}

type Peer struct {
	ID         ElevID
	PeerStatus PeerStatus
	lastSeen   time.Time
}
