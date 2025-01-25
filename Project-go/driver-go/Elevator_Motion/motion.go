package elevator_motion

import (
	"Driver-go/elevio"
)

// MoveUp moves the elevator upwards until it reaches the specified floor
func MoveUp(currentFloor, targetFloor int, drv_floors chan int) {
	if currentFloor >= targetFloor {
		return
	}
	elevio.SetMotorDirection(elevio.MD_Up)
	for {
		floor := <-drv_floors
		if floor == targetFloor {
			elevio.SetMotorDirection(elevio.MD_Stop)
			break
		}
	}
}

// MoveDown moves the elevator downwards until it reaches the specified floor
func MoveDown(currentFloor, targetFloor int, drv_floors chan int) {
	if currentFloor <= targetFloor {
		return
	}
	elevio.SetMotorDirection(elevio.MD_Down)
	for {
		floor := <-drv_floors
		if floor == targetFloor {
			elevio.SetMotorDirection(elevio.MD_Stop)
			break
		}
	}
}

func setDirection(currentFloor, targetFloor int, drv_floors chan int, d int)  {
	if currentFloor >= targetFloor {
		
		return
	}
	if currentFloor <= targetFloor {
		
		return
	}
	
