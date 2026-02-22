package localsingle

func (elevator *LocalSingleElevator) onInitBetweenFloors() command {
	elevator.State.Direction = DirDown
	elevator.State.Behaviour = BehaviourMoving
	return command{_type: setMotorDirection, value: DirDown}
}

func (elevator *LocalSingleElevator) onRequestButtonPress(buttonFloor int, buttonType ButtonType) []command {
	elevator.requests[buttonFloor][buttonType] = true
	var commands []command

	switch elevator.State.Behaviour {
	case BehaviourDoorOpen:
		if elevator.shouldClearImmediately(buttonFloor, buttonType) {
			commands = append(commands, command{_type: resetDoorTimer, value: nil})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
			}
		}
		commands = append(commands, elevator.generateLightCommands()...)
		return commands
	case BehaviourMoving:
		commands = append(commands, elevator.generateLightCommands()...)
		return commands
	case BehaviourIdle:
		pair := elevator.chooseDirection()
		elevator.State.Direction = pair.direction
		elevator.State.Behaviour = pair.behaviour
		switch pair.behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{_type: setDoorOpenLamp, value: true})
			commands = append(commands, command{_type: resetDoorTimer, value: true})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			commands = append(commands, command{_type: setMotorDirection, value: elevator.State.Direction})

		}
		commands = append(commands, elevator.generateLightCommands()...)
		return commands
	}

	return nil
}

func (elevator *LocalSingleElevator) onFloorArrival(newFloor int) []command {
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
		commands = append(commands, command{_type: resetDoorTimer, value: nil})

		commands = append(commands, elevator.generateLightCommands()...)
		elevator.State.Behaviour = BehaviourDoorOpen
	}

	return commands
}

func (elevator *LocalSingleElevator) onDoorTimeout() []command {
	var commands []command

	if elevator.State.Behaviour != BehaviourDoorOpen {
		return nil
	}

	if elevator.obstructed {
		commands = append(commands, command{_type: resetDoorTimer, value: nil})
		return commands
	}

	switch elevator.State.Behaviour {
	case BehaviourDoorOpen:
		pair := elevator.chooseDirection()
		elevator.State.Direction = pair.direction
		elevator.State.Behaviour = pair.behaviour

		switch elevator.State.Behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{_type: resetDoorTimer, value: nil})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{_type: sendClearedOrders, value: cleared})
			}
			commands = append(commands, elevator.generateLightCommands()...)
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

func (elevator *LocalSingleElevator) onObstruction(obstructed bool) []command {
	elevator.obstructed = obstructed
	if elevator.State.Behaviour == BehaviourDoorOpen {
		var commands []command
		commands = append(commands, command{_type: resetDoorTimer, value: nil})
		return commands
	}
	return nil
}
