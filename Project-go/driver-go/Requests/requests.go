package requests

import (
	"Project-go/driver-go/elevio"
)

func RequestShouldStop(e elevio.Elevator) bool {
	switch e.Direction {
	case elevio.MD_Down:
		if e.Requests[e.CurrentFloor][elevio.BT_HallDown] == 1 || !requestsBelow(e) || e.Requests[e.CurrentFloor][elevio.BT_Cab] == 1 {
			return true
		}
		return false

	case elevio.MD_Up:
		if e.Requests[e.CurrentFloor][elevio.BT_HallUp] == 1 || e.Requests[e.CurrentFloor][elevio.BT_Cab] == 1 || !requestsAbove(e) {
			return true
		}
		return false

	case elevio.MD_Stop:
	default:
		return true
	}

	return false
}

// Helper function to check if there's a request at the current floor
func hasRequestAtFloor(e elevio.Elevator) bool {
	for btn := 0; btn < 3; btn++ {
		if e.Requests[e.CurrentFloor][btn] != 0 {
			return true
		}

	}
	return false
}

func requestsAbove(e elevio.Elevator) bool {
	// Iterate from floor `e.CurrentFloor` to the top floor to check if there's any request above `e.CurrentFloor`
	for a := e.CurrentFloor + 1; a < e.NumFloors; a++ {
		for i := elevio.ButtonType(0); i < 3; i++ {
			if e.Requests[a][i] != 0 {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevio.Elevator) bool {
	// Iterate from floor 0 to `e.CurrentFloor` to check if there's any request below `e.CurrentFloor`
	for a := 0; a < e.CurrentFloor; a++ {
		for i := elevio.ButtonType(0); i < 3; i++ {
			if e.Requests[a][i] != 0 {
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

func RequestClearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {

	e.Requests[e.CurrentFloor][elevio.BT_Cab] = 0
	switch e.Direction {
	case elevio.MD_Up:
		if !requestsAbove(e) && e.Requests[e.CurrentFloor][elevio.BT_HallUp] == 0 {
			e.Requests[e.CurrentFloor][elevio.BT_HallDown] = 0
		} else {
			e.Requests[e.CurrentFloor][elevio.BT_HallUp] = 0
		}
	case elevio.MD_Down:
		if !requestsBelow(e) && e.Requests[e.CurrentFloor][elevio.BT_HallDown] == 0 {
			e.Requests[e.CurrentFloor][elevio.BT_HallUp] = 0
		} else {
			e.Requests[e.CurrentFloor][elevio.BT_HallDown] = 0
		}
	case elevio.MD_Stop:
		e.Requests[e.CurrentFloor][elevio.BT_HallUp] = 0
		e.Requests[e.CurrentFloor][elevio.BT_HallDown] = 0
		break
	}
	return e

}
