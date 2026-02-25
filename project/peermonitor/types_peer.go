package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)

// types

type ElevID = ordersync.ElevID
type PeerStatus = ordersync.PeerStatus
type PeerUpdate = ordersync.PeerUpdate
type PeerMsg = ordersync.PeerMsg

const (
	StatusAlive = ordersync.StatusAlive
	StatusDead  = ordersync.StatusDead
)

type HeartBeat struct {
	SenderID ElevID
}

type Peer struct {
	//Peermonitors peer struct differs from ordersyncs
	ID         ElevID
	PeerStatus PeerStatus
	lastSeen   time.Time
}

type PeerConfig struct {
	timeout         time.Duration // How long before a peer is declared dead
	tickInterval    time.Duration // internall ticker
	heartBeatTicker time.Duration
}
