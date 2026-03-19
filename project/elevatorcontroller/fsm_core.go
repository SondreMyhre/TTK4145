package elevatorcontroller

func onInitBetweenFloors(elevator elevator) (elevator, []effect) {
	var effects []effect
	elevator.state.Direction = DirDown
	elevator.state.Behaviour = BehaviourMoving
	effects = append(effects, effect{kind: setMotorDirection, value: DirDown})
	return elevator, effects
}

func onNewRequestMatrix(elevator elevator, newRequests RequestMatrix) (elevator, []effect) {
	var effects []effect

	elevator.requests = newRequests

	switch elevator.state.Behaviour {
	case BehaviourDoorOpen:
		if shouldStop(elevator) {
			var cleared []Order
			effects = append(effects, effect{kind: resetDoorTimer})
			elevator, cleared = clearAtCurrentFloor(elevator)
			if len(cleared) > 0 {
				effects = append(effects, effect{kind: publishClearedOrders, value: cleared})
			}
		}
		return elevator, effects
	case BehaviourMoving:
		return elevator, effects
	case BehaviourIdle:
		pair := chooseDirection(elevator)
		elevator.state.Direction = pair.direction
		elevator.state.Behaviour = pair.behaviour
		effects = append(effects, effect{kind: publishLocalState, value: elevator.state})
		switch pair.behaviour {
		case BehaviourDoorOpen:
			var cleared []Order
			effects = append(effects, effect{kind: setDoorOpenLamp, value: true})
			effects = append(effects, effect{kind: resetDoorTimer})
			elevator, cleared = clearAtCurrentFloor(elevator)
			if len(cleared) > 0 {
				effects = append(effects, effect{kind: publishClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			effects = append(effects, effect{kind: setDoorOpenLamp, value: false})
			effects = append(effects, effect{kind: setMotorDirection, value: elevator.state.Direction})
		}
		return elevator, effects
	}

	return elevator, effects
}

func onFloorArrival(elevator elevator, newFloor int) (elevator, []effect) {
	elevator.state.Floor = newFloor

	if elevator.state.MotorStuck {
		elevator.state.MotorStuck = false
	}

	var effects []effect
	effects = append(effects, effect{kind: setFloorIndicator, value: elevator.state.Floor})
	effects = append(effects, effect{kind: publishLocalState, value: elevator.state})

	if elevator.state.Behaviour != BehaviourMoving {
		return elevator, effects
	}

	if shouldStop(elevator) {
		var cleared []Order
		effects = append(effects, effect{kind: setMotorDirection, value: DirStop})
		effects = append(effects, effect{kind: setDoorOpenLamp, value: true})
		elevator, cleared = clearAtCurrentFloor(elevator)
		if len(cleared) > 0 {
			effects = append(effects, effect{kind: publishClearedOrders, value: cleared})
			effects = append(effects, effect{kind: resetDoorTimer})
			elevator.state.Behaviour = BehaviourDoorOpen
		}
		
	}
	
	return elevator, effects
}

func onDoorTimeout(elevator elevator) (elevator, []effect) {
	var effects []effect

	if elevator.state.Behaviour != BehaviourDoorOpen {
		return elevator, effects
	}

	if elevator.state.Obstructed {
		effects = append(effects, effect{kind: resetDoorTimer})
		return elevator, effects
	}

	switch elevator.state.Behaviour {
	case BehaviourDoorOpen:
		pair := chooseDirection(elevator)
		elevator.state.Direction = pair.direction
		elevator.state.Behaviour = pair.behaviour
		effects = append(effects, effect{kind: publishLocalState, value: elevator.state})

		switch elevator.state.Behaviour {
		case BehaviourDoorOpen:
			var cleared []Order
			effects = append(effects, effect{kind: resetDoorTimer})
			elevator, cleared = clearAtCurrentFloor(elevator)
			if len(cleared) > 0 {
				effects = append(effects, effect{kind: publishClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			effects = append(effects, effect{kind: setDoorOpenLamp, value: false})
			effects = append(effects, effect{kind: setMotorDirection, value: elevator.state.Direction})
		case BehaviourIdle:
			effects = append(effects, effect{kind: setDoorOpenLamp, value: false})
			effects = append(effects, effect{kind: setMotorDirection, value: elevator.state.Direction})
		}

	}
	return elevator, effects
}

func onObstruction(elevator elevator, obstructed bool) (elevator, []effect) {
	var effects []effect
	elevator.state.Obstructed = obstructed
	if elevator.state.Behaviour == BehaviourDoorOpen {
		effects = append(effects, effect{kind: resetDoorTimer})
		effects = append(effects, effect{kind: publishLocalState, value: elevator.state})
		return elevator, effects
	}
	return elevator, effects
}

func onMotorTimeout(elevator elevator) (elevator, []effect) {
	var effects []effect

	if elevator.state.Behaviour != BehaviourMoving {
		return elevator, effects
	}

	elevator.state.MotorStuck = true

	effects = append(effects, effect{kind: publishLocalState, value: elevator.state})
	return elevator, effects
}
