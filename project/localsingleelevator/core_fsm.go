package localsingle

func (elevator *elevator) onInitBetweenFloors() command {
	elevator.state.Direction = DirDown
	elevator.state.Behaviour = BehaviourMoving
	return command{_type: setMotorDirection, value: DirDown}
}

func (elevator *elevator) onRequestButtonPress(buttonFloor int, buttonType ButtonType) []command {
	elevator.requests[buttonFloor][buttonType] = true
	var commands []command

	switch elevator.state.Behaviour {
	case BehaviourDoorOpen:
		if elevator.shouldClearImmediately(buttonFloor, buttonType) {
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
		switch pair.behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{_type: setDoorOpenLamp, value: true})
			commands = append(commands, command{_type: resetDoorTimer})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
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
		commands = append(commands, command{_type: sendLocalState})
		return commands
	}
	return commands
}
