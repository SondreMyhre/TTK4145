package localsingle

func (elevator *elevator) onInitBetweenFloors() command {
	elevator.State.Direction = DirDown
	elevator.State.Behaviour = BehaviourMoving
	return command{_type: setMotorDirection, value: DirDown}
}

func (elevator *elevator) onRequestButtonPress(buttonFloor int, buttonType ButtonType) []command {
	elevator.requests[buttonFloor][buttonType] = true
	var commands []command

	switch elevator.State.Behaviour {
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
		elevator.State.Direction = pair.direction
		elevator.State.Behaviour = pair.behaviour
		switch pair.behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{_type: setDoorOpenLamp, value: true})
			commands = append(commands, command{_type: resetDoorTimer})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			commands = append(commands, command{_type: setMotorDirection, value: elevator.State.Direction})

		}
		return commands
	}

	return commands
}

func (elevator *elevator) onFloorArrival(newFloor int) []command {
	elevator.State.Floor = newFloor
	var commands []command
	commands = append(commands, command{_type: setFloorIndicator, value: elevator.State.Floor})

	if elevator.State.Behaviour != BehaviourMoving {
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

		elevator.State.Behaviour = BehaviourDoorOpen
	}

	return commands
}

func (elevator *elevator) onDoorTimeout() []command {
	var commands []command

	if elevator.State.Behaviour != BehaviourDoorOpen {
		return commands
	}

	if elevator.obstructed {
		commands = append(commands, command{_type: resetDoorTimer})
		return commands
	}

	switch elevator.State.Behaviour {
	case BehaviourDoorOpen:
		pair := elevator.chooseDirection()
		elevator.State.Direction = pair.direction
		elevator.State.Behaviour = pair.behaviour

		switch elevator.State.Behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{_type: resetDoorTimer})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			commands = append(commands, command{_type: setDoorOpenLamp, value: false})
			commands = append(commands, command{_type: setMotorDirection, value: elevator.State.Direction})
		case BehaviourIdle:
			commands = append(commands, command{_type: setDoorOpenLamp, value: false})
			commands = append(commands, command{_type: setMotorDirection, value: elevator.State.Direction})
		}

	}
	return commands
}

func (elevator *elevator) onObstruction(obstructed bool) []command {
	var commands []command
	elevator.obstructed = obstructed
	if elevator.State.Behaviour == BehaviourDoorOpen {
		commands = append(commands, command{_type: resetDoorTimer})
		return commands
	}
	return commands
}
