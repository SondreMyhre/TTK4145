package localsingle

func (elevator *elevator) onInitBetweenFloors() []command {
	var commands []command
	elevator.state.Direction = DirDown
	elevator.state.Behaviour = BehaviourMoving
	commands = append(commands, command{_type: sendLocalState, value: elevator.state})
	commands = append(commands, command{_type: setMotorDirection, value: DirDown})
	return commands
}

func (elevator *elevator) onNewRequestMatrix(newRequests RequestMatrix) []command {
	var commands []command

	elevator.requests = newRequests

	switch elevator.state.Behaviour {
	case BehaviourDoorOpen:
		if elevator.shouldStop() {
			commands = append(commands, command{_type: resetDoorTimer})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
			}
		}
		return commands
	case BehaviourMoving:
		return commands
	case BehaviourIdle:
		pair := elevator.chooseDirection()
		elevator.state.Direction = pair.direction
		elevator.state.Behaviour = pair.behaviour
		commands = append(commands, command{_type: sendLocalState, value: elevator.state})
		switch pair.behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{_type: setDoorOpenLamp, value: true})
			commands = append(commands, command{_type: resetDoorTimer})
			clearedOrders := elevator.clearAtCurrentFloor()
			if len(clearedOrders) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: clearedOrders})
			}
		case BehaviourMoving:
			commands = append(commands, command{_type: setMotorDirection, value: elevator.state.Direction})

		}
		return commands
	}

	return commands
}

func (elevator *elevator) onFloorArrival(newFloor int) []command {
	elevator.state.Floor = newFloor

	if elevator.state.MotorStuck {
		elevator.state.MotorStuck = false
	}

	var commands []command
	commands = append(commands, command{_type: setFloorIndicator, value: elevator.state.Floor})

	if elevator.state.Behaviour != BehaviourMoving {
		return commands
	}

	if elevator.shouldStop() {
		commands = append(commands, command{_type: setMotorDirection, value: DirStop})
		commands = append(commands, command{_type: setDoorOpenLamp, value: true})
		cleared := elevator.clearAtCurrentFloor()
		if len(cleared) > 0 {
			commands = append(commands, command{_type: sendClearedOrders, value: cleared})
		}
		commands = append(commands, command{_type: resetDoorTimer})

		elevator.state.Behaviour = BehaviourDoorOpen
		commands = append(commands, command{_type: sendLocalState, value: elevator.state})
	}

	return commands
}

func (elevator *elevator) onDoorTimeout() []command {
	var commands []command

	if elevator.state.Behaviour != BehaviourDoorOpen {
		return commands
	}

	if elevator.state.Obstructed {
		commands = append(commands, command{_type: resetDoorTimer})
		return commands
	}

	switch elevator.state.Behaviour {
	case BehaviourDoorOpen:
		pair := elevator.chooseDirection()
		elevator.state.Direction = pair.direction
		elevator.state.Behaviour = pair.behaviour
		commands = append(commands, command{_type: sendLocalState, value: elevator.state})

		switch elevator.state.Behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{_type: resetDoorTimer})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			commands = append(commands, command{_type: setDoorOpenLamp, value: false})
			commands = append(commands, command{_type: setMotorDirection, value: elevator.state.Direction})
		case BehaviourIdle:
			commands = append(commands, command{_type: setDoorOpenLamp, value: false})
			commands = append(commands, command{_type: setMotorDirection, value: elevator.state.Direction})
		}

	}
	return commands
}

func (elevator *elevator) onObstruction(obstructed bool) []command {
	var commands []command
	elevator.state.Obstructed = obstructed
	if elevator.state.Behaviour == BehaviourDoorOpen {
		commands = append(commands, command{_type: resetDoorTimer})
		commands = append(commands, command{_type: sendLocalState, value: elevator.state})
		return commands
	}
	return commands
}

func (elevator *elevator) onMotorTimeout() []command {
	var commands []command
	
	if elevator.state.Behaviour != BehaviourMoving {
		return commands
	}

	elevator.state.MotorStuck = true

    commands = append(commands, command{_type: sendLocalState, value: elevator.state})
    return commands
}