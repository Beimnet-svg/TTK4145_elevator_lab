package elevator_motion

import (
	"Driver-go/elevio"
	"fmt"
)

func SetDirection(currentFloor int, targetFloor int, d elevio.MotorDirection) elevio.MotorDirection {
	if currentFloor > targetFloor {
		d = elevio.MD_Down
		return d
	}
	if currentFloor < targetFloor {
		d = elevio.MD_Up
		return d
	} else {
		fmt.Printf("Stopped")
		d = elevio.MD_Stop
		return d
	}

}


