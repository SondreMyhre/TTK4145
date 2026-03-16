package elevatorcontroller

func onInitBetweenFloors(elevator elevator) (elevator, []command) {
	var commands []command
	elevator.state.Direction = DirDown
	elevator.state.Behaviour = BehaviourMoving
	commands = append(commands, command{kind: setMotorDirection, value: DirDown})
	return elevator, commands
}

func onNewRequestMatrix(elevator elevator, newRequests RequestMatrix) (elevator, []command) {
	var commands []command

	elevator.requests = newRequests

	switch elevator.state.Behaviour {
	case BehaviourDoorOpen:
		if elevator.shouldStop() {
			commands = append(commands, command{kind: resetDoorTimer})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{kind: sendClearedOrders, value: cleared})
			}
		}
		return elevator, commands
	case BehaviourMoving:
		return elevator, commands
	case BehaviourIdle:
		pair := elevator.chooseDirection()
		elevator.state.Direction = pair.direction
		elevator.state.Behaviour = pair.behaviour
		commands = append(commands, command{kind: sendLocalState, value: elevator.state})
		switch pair.behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{kind: setDoorOpenLamp, value: true})
			commands = append(commands, command{kind: resetDoorTimer})
			clearedOrders := elevator.clearAtCurrentFloor()
			if len(clearedOrders) > 0 {
				commands = append(commands, command{kind: sendClearedOrders, value: clearedOrders})
			}
		case BehaviourMoving:
			commands = append(commands, command{kind: setMotorDirection, value: elevator.state.Direction})

		}
		return elevator, commands
	}

	return elevator, commands
}

func onFloorArrival(elevator elevator, newFloor int) (elevator, []command) {
	elevator.state.Floor = newFloor

	if elevator.state.MotorStuck {
		elevator.state.MotorStuck = false
	}

	var commands []command
	commands = append(commands, command{kind: setFloorIndicator, value: elevator.state.Floor})

	if elevator.state.Behaviour != BehaviourMoving {
		return elevator, commands
	}

	if elevator.shouldStop() {
		commands = append(commands, command{kind: setMotorDirection, value: DirStop})
		commands = append(commands, command{kind: setDoorOpenLamp, value: true})
		cleared := elevator.clearAtCurrentFloor()
		if len(cleared) > 0 {
			commands = append(commands, command{kind: sendClearedOrders, value: cleared})
		}
		commands = append(commands, command{kind: resetDoorTimer})

		elevator.state.Behaviour = BehaviourDoorOpen
		commands = append(commands, command{kind: sendLocalState, value: elevator.state})
	}

	return elevator, commands
}

func onDoorTimeout(elevator elevator) (elevator, []command) {
	var commands []command

	if elevator.state.Behaviour != BehaviourDoorOpen {
		return elevator, commands
	}

	if elevator.state.Obstructed {
		commands = append(commands, command{kind: resetDoorTimer})
		return elevator, commands
	}

	switch elevator.state.Behaviour {
	case BehaviourDoorOpen:
		pair := elevator.chooseDirection()
		elevator.state.Direction = pair.direction
		elevator.state.Behaviour = pair.behaviour
		commands = append(commands, command{kind: sendLocalState, value: elevator.state})

		switch elevator.state.Behaviour {
		case BehaviourDoorOpen:
			commands = append(commands, command{kind: resetDoorTimer})
			cleared := elevator.clearAtCurrentFloor()
			if len(cleared) > 0 {
				commands = append(commands, command{kind: sendClearedOrders, value: cleared})
			}
		case BehaviourMoving:
			commands = append(commands, command{kind: setDoorOpenLamp, value: false})
			commands = append(commands, command{kind: setMotorDirection, value: elevator.state.Direction})
		case BehaviourIdle:
			commands = append(commands, command{kind: setDoorOpenLamp, value: false})
			commands = append(commands, command{kind: setMotorDirection, value: elevator.state.Direction})
		}

	}
	return elevator, commands
}

func onObstruction(elevator elevator, obstructed bool) (elevator, []command) {
	var commands []command
	elevator.state.Obstructed = obstructed
	if elevator.state.Behaviour == BehaviourDoorOpen {
		commands = append(commands, command{kind: resetDoorTimer})
		commands = append(commands, command{kind: sendLocalState, value: elevator.state})
		return elevator, commands
	}
	return elevator, commands
}

func onMotorTimeout(elevator elevator) (elevator, []command) {
	var commands []command

	if elevator.state.Behaviour != BehaviourMoving {
		return elevator, commands
	}

	elevator.state.MotorStuck = true

	commands = append(commands, command{kind: sendLocalState, value: elevator.state})
	return elevator, commands
}
