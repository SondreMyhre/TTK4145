package ordersync

import (
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
	"strconv"
)

func onCabButtonEvent(cabCalls CabCallsMap, pendingCabCalls [N_FLOORS]bool, myID ElevID, buttonEvent elevio.ButtonEvent) (CabCallsMap, [N_FLOORS]bool, []command) {
	pendingCabCalls[buttonEvent.Floor] = true

	localCabCalls := cabCalls[myID]
	localCabCalls[buttonEvent.Floor] = true
	cabCalls[myID] = localCabCalls

	return cabCalls, pendingCabCalls, []command{{_type: sendOrderToLocal, value: buttonEvent}, {_type: broadcastNetMessage}}
}

func onHallButtonEvent(hallOrderMatrix HallOrderMatrix, buttonEvent elevio.ButtonEvent) (HallOrderMatrix, []command) {
	floor := buttonEvent.Floor
	btn := buttonEvent.Button

	entry := hallOrderMatrix[floor][btn]

	if entry.Status == Inactive {
		entry.Status = Pending
		entry.AssignedElevator = ""
		entry.Version++
		hallOrderMatrix[floor][btn] = entry
	}

	return hallOrderMatrix, []command{{_type: broadcastNetMessage}}
}

func onNewLocalState(hallOrderMatrix HallOrderMatrix, peerList []Peer, myID ElevID, cabCalls CabCallsMap, newLocalState localsingle.ElevatorState) (HallOrderMatrix, localsingle.ElevatorState, []command) {

	var commands []command
	var localState localsingle.ElevatorState
	broadcastNeeded := false
	localState.Floor = newLocalState.Floor
	localState.Direction = newLocalState.Direction
	localState.Behaviour = newLocalState.Behaviour
	localState.Obstructed = newLocalState.Obstructed

	if localState.Obstructed {
		hallOrderMatrix = releaseOrdersForPeer(hallOrderMatrix, myID)
		broadcastNeeded = true
	} else {
		orders := detertmineMyOrders(hallOrderMatrix, myID, localState, cabCalls, peerList)
		for _, order := range orders {
			var claimCommands []command
			hallOrderMatrix, claimCommands = claimOrder(hallOrderMatrix, myID, order)
			commands = append(commands, claimCommands...)
			broadcastNeeded = true
		}
	}

	if broadcastNeeded {
		commands = append(commands, command{_type: broadcastNetMessage})
	}

	return hallOrderMatrix, localState, commands
}

func onClearedOrders(hallOrderMatrix HallOrderMatrix, cabCalls CabCallsMap, myID ElevID, clearedOrders []localsingle.Order) (HallOrderMatrix, CabCallsMap, []command) {
	var commands []command

	if len(clearedOrders) == 0 {
		return hallOrderMatrix, cabCalls, commands
	}

	for _, order := range clearedOrders {
		floor := order.Floor
		btn := localsingle.ButtonTypeToElevio(order.Button)

		if btn == elevio.BT_HallUp || btn == elevio.BT_HallDown {
			entry := &hallOrderMatrix[floor][btn]
			if entry.Status != Inactive {
				entry.Status = Inactive
				entry.AssignedElevator = ""
				entry.Version++

				commands = append(commands, command{
					_type: setButtonLamp,
					value: buttonLampArgs{
						Floor:  floor,
						Button: btn,
						Value:  false,
					},
				})
			}
		} else if btn == elevio.BT_Cab {
			localCabCalls := cabCalls[myID]
			localCabCalls[floor] = false
			cabCalls[myID] = localCabCalls
			commands = append(commands, command{
				_type: setButtonLamp,
				value: buttonLampArgs{
					Floor:  floor,
					Button: btn,
					Value:  false,
				},
			})
		}
	}
	commands = append(commands, command{_type: broadcastNetMessage})

	return hallOrderMatrix, cabCalls, commands
}

func claimOrder(hallOrderMatrix HallOrderMatrix, myID ElevID, order OrderLocation) (HallOrderMatrix, []command) {
	var commands []command

	entry := &hallOrderMatrix[order.Floor][order.Button]
	entry.Status = Assigned
	entry.AssignedElevator = myID
	entry.Version++

	commands = []command{
		{_type: sendOrderToLocal, value: elevio.ButtonEvent{Floor: order.Floor, Button: order.Button}},
		{_type: setButtonLamp, value: buttonLampArgs{Floor: order.Floor, Button: order.Button, Value: true}},
	}

	return hallOrderMatrix, commands
}

func onNetMsg(hallOrderMatrix HallOrderMatrix, cabCalls CabCallsMap, myID ElevID, pendingCabCalls [N_FLOORS]bool, peerList []Peer, msg NetMsg) (HallOrderMatrix, CabCallsMap, [N_FLOORS]bool, []command) {
	var commands []command
	broadcastNeeded := false
	senderID := msg.SenderID

	if senderID == myID {
		return hallOrderMatrix, cabCalls, pendingCabCalls, commands
	}

	for i := range peerList {
		if peerList[i].ID == senderID {
			peerList[i].State = msg.SenderState
			break
		}
	}

	if msg.SenderState.Obstructed {
		hallOrderMatrix = releaseOrdersForPeer(hallOrderMatrix, senderID)
		broadcastNeeded = true
	}

	for floor := range N_FLOORS {
		for btn := range N_HALL {
			remote := msg.HallOrderMatrix[floor][btn]
			local := &hallOrderMatrix[floor][btn]

			if remote.Version > local.Version {
				*local = remote

				switch local.Status {
				case Inactive:
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
							Value:  false,
						},
					})
				case Pending:
					local.Status = Confirmed
					local.Version++
					broadcastNeeded = true
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
							Value:  true,
						},
					})
				}

			} else if remote.Version == local.Version {
				switch remote.Status {
				case Pending:
					if local.Status == Pending {
						local.Status = Confirmed
						local.Version++
						broadcastNeeded = true
						commands = append(commands, command{
							_type: setButtonLamp,
							value: buttonLampArgs{
								Floor:  floor,
								Button: elevio.ButtonType(btn),
								Value:  true,
							},
						})
					}

				case Confirmed, Assigned:
					if remote.Status > local.Status {
						*local = remote
					}
					if local.AssignedElevator != remote.AssignedElevator && local.AssignedElevator != "" {
						remoteInt, _ := strconv.Atoi(string(remote.AssignedElevator))
						localInt, _ := strconv.Atoi(string(local.AssignedElevator))
						if remoteInt < localInt && remote.AssignedElevator != "" {
							*local = remote
						}
					}
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
							Value:  true,
						},
					})
				}

			}
		}
	}

	if msg.CabCalls != nil {
		cabCalls[senderID] = msg.CabCalls[senderID]
		for i := range pendingCabCalls {
			if pendingCabCalls[i] && msg.CabCalls[myID][i] {
				pendingCabCalls[i] = false
				commands = append(commands, command{
					_type: setButtonLamp,
					value: buttonLampArgs{
						Floor:  i,
						Button: elevio.BT_Cab,
						Value:  true,
					},
				})
			}
		}
	}

	if broadcastNeeded {
		commands = append(commands, command{_type: broadcastNetMessage})
	}

	return hallOrderMatrix, cabCalls, pendingCabCalls, commands
}

func onHeartbeatTick() []command {
	return []command{{_type: broadcastNetMessage}}
}

func onPeerEvent(hallOrderMatrix HallOrderMatrix, oldPeerList []Peer, newPeerList []Peer) (HallOrderMatrix, []Peer, []command) {
	var commands []command

	for _, newPeer := range newPeerList {
		oldStatus := findPeerStatus(oldPeerList, newPeer.ID)

		if oldStatus == Alive && newPeer.Status == Dead {
			hallOrderMatrix = releaseOrdersForPeer(hallOrderMatrix, newPeer.ID)
		}

	}
	commands = append(commands, command{_type: broadcastNetMessage})
	return hallOrderMatrix, newPeerList, commands
}

func findPeerStatus(peerList []Peer, id ElevID) PeerStatus {
	for _, peer := range peerList {
		if peer.ID == id {
			return peer.Status
		}
	}
	return PeerStatus(-1)
}

func releaseOrdersForPeer(hallOrderMatrix HallOrderMatrix, deadID ElevID) HallOrderMatrix {
	for floor := range N_FLOORS {
		for btn := range N_HALL {
			entry := &hallOrderMatrix[floor][btn]
			if entry.AssignedElevator == deadID && entry.Status == Assigned {
				entry.Status = Pending
				entry.AssignedElevator = ""
				entry.Version++
			}
		}
	}
	return hallOrderMatrix
}
