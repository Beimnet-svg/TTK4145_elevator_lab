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

func RequestChooseDir() {

}

func ReqestShouldClearImmideatly() {

}

func RequestClearAtCurrentFloor() {

}
