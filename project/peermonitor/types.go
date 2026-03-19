package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)

const (
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
