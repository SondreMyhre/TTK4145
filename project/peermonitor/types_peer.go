package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)

// types

type ElevID = ordersync.ElevID
type NetMsg = ordersync.NetMsg
type PeerStatus = ordersync.PeerStatus
type PeerUpdate = ordersync.PeerUpdate
type PeerMsg = ordersync.PeerMsg

const (
	StatusAlive = ordersync.StatusAlive
	StatusDead  = ordersync.StatusDead
)

type Peer struct {
	//Peermonitors peer struct differs from ordersyncs
	ID         ElevID
	PeerStatus PeerStatus
	lastSeen   time.Time
}

type PeerConfig struct {
	Timeout      time.Duration // How long before a peer is declared dead
	TickInterval time.Duration // internall ticker
}
