package requests

import "Driver-go/elevio"

func RequestShouldStop(e elevio.Elevator) bool {
	switch e.Direction {
	case elevio.MD_Down:
		if hasRequestAtFloor(e, elevio.BT_HallDown) || hasRequestAtFloor(e, elevio.BT_Cab) || !requestsBelow(e) {
			return true
		}
		return false

	case elevio.MD_Up:
		if hasRequestAtFloor(e, elevio.BT_HallUp) || hasRequestAtFloor(e, elevio.BT_Cab) || !requestsAbove(e) {
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
func hasRequestAtFloor(e elevio.Elevator, btnType elevio.ButtonType) bool {
	for _, req := range e.Requests {
		if req.Floor == e.CurrentFloor && req.Button == btnType {
			return true
		}
	}
	return false
}

// Placeholder functions: You should implement requestsAbove(e) and requestsBelow(e)
func requestsAbove(e elevio.Elevator) bool {
	// Iterate through `e.Requests` to check if there's any request above `e.CurrentFloor`
	for _, req := range e.Requests {
		if req.Floor > e.CurrentFloor {
			return true
		}
	}
	return false
}

func requestsBelow(e elevio.Elevator) bool {
	// Iterate through `e.Requests` to check if there's any request below `e.CurrentFloor`
	for _, req := range e.Requests {
		if req.Floor < e.CurrentFloor {
			return true
		}
	}
	return false
}

func RequestChooseDir(e elevio.Elevator, b elevio.ButtonType) (elevio.MotorDirection, elevio.ElevatorBehaviour) {
	switch e.Direction {
		case elevio.MD_Up:
			if requestsAbove(e) {
				return elevio.MD_Up, elevio.EB_Moving
			} else if hasRequestAtFloor(e,b) {
				return elevio.MD_Down, elevio.EB_DoorOpen
			} else if requestsBelow(e) {
				return elevio.MD_Down, elevio.EB_Moving
			}
			return elevio.MD_Stop, elevio.EB_Idle
		case elevio.MD_Down:
			if requestsBelow(e) {
				return elevio.MD_Down, elevio.EB_Moving
			} else if hasRequestAtFloor(e,b) {
				return elevio.MD_Up, elevio.EB_DoorOpen
			} else if requestsAbove(e) {
				return elevio.MD_Up, elevio.EB_Moving
			}
			return elevio.MD_Stop, elevio.EB_Idle

		case elevio.MD_Stop:
			if hasRequestAtFloor(e,b) {
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
	if (e.CurrentFloor == floor &&
	(e.Direction == elevio.MD_Up && b == elevio.BT_HallUp) ||
		(e.Direction == elevio.MD_Down && b == elevio.BT_HallDown) ||
		e.Direction == elevio.MD_Stop ||
		b == elevio.BT_Cab) {
		return true
		}
	return false
}



func RequestClearAtCurrentFloor(e elevio.Elevator) elevio.Elevator {
	for i, req := range e.Requests {
		if req.Floor == e.CurrentFloor && req.Button==elevio.BT_Cab {
			e.Requests = append(e.Requests[:i], e.Requests[i+1:]...)
		}
	}
	return e
	

}
