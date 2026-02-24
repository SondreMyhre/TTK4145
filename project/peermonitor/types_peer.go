package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)


// types

type ElevID = ordersync.ElevID
type NetMsg = ordersync.NetMsg
type Peer = ordersync.Peer
type PeerStatus = ordersync.PeerStatus

const (
	StatusAlive = ordersync.StatusAlive
	StatusDead  = ordersync.StatusDead
)


type PeerUpdate struct {
	Peers []Peer // Current list of all peers with their status
}

type PeerConfig struct {
	Timeout time.Duration // How long before a peer is declared dead
	TickInterval   time.Duration // internall ticker
}


