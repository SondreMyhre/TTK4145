package localsingle

func (elevator *LocalSingleElevator) onInitBetweenFloors() Command {
	elevator.state.direction = DirDown
	elevator.state.behaviour = BehaviourMoving
	return Command{_type: setMotorDirection, value: DirDown}
}

func (elevator *LocalSingleElevator) onRequestButtonPress(buttonFloor int, buttonType ButtonType) []Command {
	elevator.requests[buttonFloor][buttonType] = true
	var commands []Command

	switch elevator.state.behaviour {
	case BehaviourDoorOpen:
		if elevator.shouldClearImmediately(buttonFloor, buttonType) {
			commands = append(commands, Command{_type: resetDoorTimer, value: nil})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, Command{_type: sendClearedOrders, value: cleared})
			}
		}
		commands = append(commands, elevator.generateLightCommands()...)
		return commands
	case BehaviourMoving:
		commands = append(commands, elevator.generateLightCommands()...)
		return commands
	case BehaviourIdle:
		pair := elevator.chooseDirection()
		elevator.state.direction = pair.direction
		elevator.state.behaviour = pair.behaviour
		switch pair.behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, Command{_type: setDoorOpenLamp, value: true})
			commands = append(commands, Command{_type: resetDoorTimer, value: true})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, Command{_type: sendClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			commands = append(commands, Command{_type: setMotorDirection, value: elevator.state.direction})

		}
		commands = append(commands, elevator.generateLightCommands()...)
		return commands
	}

	return nil
}

func (elevator *LocalSingleElevator) onFloorArrival(newFloor int) []Command {
	elevator.state.floor = newFloor
	var commands []Command
	commands = append(commands, Command{_type: setFloorIndicator, value: elevator.state.floor})

	if elevator.state.behaviour != BehaviourMoving {
		return commands
	}

	if elevator.shouldStop() {
		commands = append(commands, Command{_type: setMotorDirection, value: DirStop})
		commands = append(commands, Command{_type: setDoorOpenLamp, value: true})
		cleared := elevator.clearAtCurrentFloor()
		if len(cleared) > 0 {
			commands = append(commands, Command{_type: sendClearedOrders, value: cleared})
		}
		commands = append(commands, Command{_type: resetDoorTimer, value: nil})

		commands = append(commands, elevator.generateLightCommands()...)
		elevator.state.behaviour = BehaviourDoorOpen
	}

	return commands
}

func (elevator *LocalSingleElevator) onDoorTimeout() []Command {
	var commands []Command

	if elevator.state.behaviour != BehaviourDoorOpen {
		return nil
	}

	if elevator.obstructed {
		commands = append(commands, Command{_type: resetDoorTimer, value: nil})
		return commands
	}

	switch elevator.state.behaviour {
	case BehaviourDoorOpen:
		pair := elevator.chooseDirection()
		elevator.state.direction = pair.direction
		elevator.state.behaviour = pair.behaviour

		switch elevator.state.behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, Command{_type: resetDoorTimer, value: nil})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, Command{_type: sendClearedOrders, value: cleared})
			}
			commands = append(commands, elevator.generateLightCommands()...)
		case BehaviourMoving:
			commands = append(commands, Command{_type: setDoorOpenLamp, value: false})
			commands = append(commands, Command{_type: setMotorDirection, value: elevator.state.direction})
		case BehaviourIdle:
			commands = append(commands, Command{_type: setDoorOpenLamp, value: false})
			commands = append(commands, Command{_type: setMotorDirection, value: elevator.state.direction})
		}

	}
	return commands
}

func (elevator *LocalSingleElevator) onObstruction(obstructed bool) []Command {
	elevator.obstructed = obstructed
	if elevator.state.behaviour == BehaviourDoorOpen {
		var commands []Command
		commands = append(commands, Command{_type: resetDoorTimer, value: nil})
		return commands
	}
	return nil
}
