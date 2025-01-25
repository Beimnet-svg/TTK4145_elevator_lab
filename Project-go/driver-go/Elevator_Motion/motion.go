package elevator_motion

import (
	"Driver-go/elevio"
)

func SetDirection(currentFloor int, targetFloor int, d elevio.MotorDirection) elevio.MotorDirection {
	if currentFloor >= targetFloor {
		d = elevio.MD_Down
		return d
	}
	if currentFloor <= targetFloor {
		d = elevio.MD_Up
		return d
	} else {
		d = elevio.MD_Stop
		return d
	}

}
