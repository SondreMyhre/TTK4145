package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)

const (
	PEER_TIMEOUT            = 5 * time.Second
	PEER_TICK_INTERVAL      = 50 * time.Millisecond
	HEART_BEAT_TICK_INTERVAL = 1 * time.Second
)

type ElevID = ordersync.ElevID
type PeerStatus = ordersync.PeerStatus
type PeerUpdate = ordersync.PeerUpdate
type PeerMsg = []PeerUpdate

const (
	StatusAlive = ordersync.StatusAlive
	StatusDead  = ordersync.StatusDead
)

type HeartBeat struct {
	SenderID ElevID
}

type Peer struct {
	ID         ElevID
	PeerStatus PeerStatus
	lastSeen   time.Time
}

type PeerConfig struct {
	Timeout         time.Duration
	TickInterval    time.Duration
	HeartBeatTicker time.Duration
}
