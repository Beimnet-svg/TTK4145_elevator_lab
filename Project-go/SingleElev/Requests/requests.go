package requests

import (
	config "Project-go/Config"
	elevio "Project-go/SingleElev/Elevio"
)

func RequestShouldStop(e elevio.Elevator) bool {
	switch e.Direction {
	case elevio.MD_Down:
		if e.ActiveOrders[e.CurrentFloor][elevio.BT_HallDown] || !requestsBelow(e) || e.ActiveOrders[e.CurrentFloor][elevio.BT_Cab] {
			return true
		}
		return false

	case elevio.MD_Up:
		if e.ActiveOrders[e.CurrentFloor][elevio.BT_HallUp] || e.ActiveOrders[e.CurrentFloor][elevio.BT_Cab] || !requestsAbove(e) {
			return true
		}
		return false

	case elevio.MD_Stop:
	default:
		return true
	}

	return false
}

func hasRequestAtFloor(e elevio.Elevator) bool {
	for btn := 0; btn < config.NumberBtn; btn++ {
		if e.ActiveOrders[e.CurrentFloor][btn] {
			return true
		}

	}
	return false
}

func requestsAbove(e elevio.Elevator) bool {
	for a := e.CurrentFloor + 1; a < config.NumberFloors; a++ {
		for i := elevio.ButtonType(0); i < config.NumberBtn; i++ {
			if e.ActiveOrders[a][i] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevio.Elevator) bool {
	for a := 0; a < e.CurrentFloor; a++ {
		for i := elevio.ButtonType(0); i < config.NumberBtn; i++ {
			if e.ActiveOrders[a][i] {
				return true
			}
		}
	}
	return false
}

func RequestChooseDir(e elevio.Elevator) (elevio.MotorDirection, elevio.ElevatorBehaviour) {
	switch e.Direction {
	case elevio.MD_Up:
		if requestsAbove(e) {
			return elevio.MD_Up, elevio.EB_Moving
		} else if hasRequestAtFloor(e) {
			return elevio.MD_Down, elevio.EB_DoorOpen
		} else if requestsBelow(e) {
			return elevio.MD_Down, elevio.EB_Moving
		}
		return elevio.MD_Stop, elevio.EB_Idle
	case elevio.MD_Down:
		if requestsBelow(e) {
			return elevio.MD_Down, elevio.EB_Moving
		} else if hasRequestAtFloor(e) {
			return elevio.MD_Up, elevio.EB_DoorOpen
		} else if requestsAbove(e) {
			return elevio.MD_Up, elevio.EB_Moving
		}
		return elevio.MD_Stop, elevio.EB_Idle

	case elevio.MD_Stop:
		if hasRequestAtFloor(e) {
			return elevio.MD_Stop, elevio.EB_DoorOpen
		} else if requestsAbove(e) {
			return elevio.MD_Up, elevio.EB_Moving
		} else if requestsBelow(e) {
			return elevio.MD_Down, elevio.EB_Moving
		}
		return elevio.MD_Stop, elevio.EB_Idle
	}
	//Should never reach this point
	return elevio.MD_Stop, elevio.EB_Idle

}

func ReqestShouldClearImmideatly(e elevio.Elevator, floor int, b elevio.ButtonType) bool {
	if e.CurrentFloor == floor &&
		((e.Direction == elevio.MD_Up && b == elevio.BT_HallUp) ||
			(e.Direction == elevio.MD_Down && b == elevio.BT_HallDown) ||
			(e.Direction == elevio.MD_Stop) ||
			(b == elevio.BT_Cab)) {
		return true
	}
	return false
}

// Removing orders from allActiveOrders in the ordermanager based on the elevator's current floor and direction. Used in the ordermanager
func RequestClearAtCurrentFloor(e elevio.Elevator, allActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {
	if e.Behaviour == elevio.EB_DoorOpen {
		allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_Cab] = false
		switch e.Direction {
		case elevio.MD_Up:
			if !requestsAbove(e) && !allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] = false
			} else {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] = false
			}
		case elevio.MD_Down:
			if !requestsBelow(e) && !allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] = false
			} else {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] = false
			}
		case elevio.MD_Stop:
			if requestsBelow(e) && allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] = false
			} else if requestsAbove(e) && allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] = false
			} else if allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] = false

			} else if allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] {
				allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] = false
			}
		default:
			allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallUp] = false
			allActiveOrders[e.ElevatorID][e.CurrentFloor][elevio.BT_HallDown] = false
		}
	}

	return allActiveOrders

}
