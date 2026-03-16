package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)

const (
	peerTimeout           = 5 * time.Second
	peerTickInterval      = 50 * time.Millisecond
	heartBeatTickInterval = 1 * time.Second
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
