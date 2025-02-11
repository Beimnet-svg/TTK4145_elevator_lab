package elevator_fsm

import (
	doors "Driver-go/Doors"
	elevator_motion "Driver-go/Elevator_Motion"
	"Driver-go/elevio"
	"fmt"
	"time"
)

var (
	//Define elevator type
	e = elevio.Elevator{
		CurrentFloor: 0,
		Direction:    elevio.MD_Stop,
		Behaviour:    elevio.EB_Idle,
		Requests:     []elevio.ButtonEvent{},
		NumFloors:    4,
	}
)

func FSM_onFloorArrival(floor int, drv_button chan elevio.ButtonEvent) {

	elevio.SetFloorIndicator(floor)
	e.CurrentFloor = floor
	e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
	elevio.SetMotorDirection(e.Direction)

	fmt.Printf("%+v\n", floor)
	e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
	elevio.SetMotorDirection(e.Direction)

	if e.Direction == elevio.MD_Stop {
		e.Requests = doors.OpenDoor(e.CurrentFloor, e.Requests, e.NumFloors, drv_button)
		e.Requests = elevio.RemoveFromQueue(e.CurrentFloor, e.Direction, e.Requests)
		elevio.LightButtons(e.Requests, e.NumFloors)
		if len(e.Requests) != 0 {
			e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
			elevio.SetMotorDirection(e.Direction)
		}

	}

	// elevio.SetFloorIndicator(floor)
	// e.CurrentFloor = floor
	// fmt.Printf("Her")
	// e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
	// elevio.SetMotorDirection(e.Direction)
	// e.Behaviour=elevio.EB_Moving

	// switch e.Behaviour {
	// case elevio.EB_Moving:
	// 	if(requests.RequestShouldStop(*e)){
	// 		elevio.SetMotorDirection(elevio.MD_Stop)
	// 		elevio.SetDoorOpenLamp(true)
	// 		e.Requests=doors.OpenDoor(e.CurrentFloor, e.Requests, 4, DRV_buttons)
	// 		e.Behaviour = elevio.EB_DoorOpen
	// 		e.Direction = elevio.MD_Stop
	// 		e.Requests = elevio.RemoveFromQueue(e.CurrentFloor, e.Direction, e.Requests)
	// 		}
	// 		break
	// 	default:
	// 		break

	// }

}
func init_elevator(drv_floors chan int) {

	for a := 0; a < e.NumFloors; a++ {
		for i := elevio.ButtonType(0); i < 3; i++ {
			elevio.SetButtonLamp(i, a, false)
		}
	}

	e.Direction = elevio.MD_Up

	elevio.SetMotorDirection(e.Direction)

	elevio.SetDoorOpenLamp(false)
	e.CurrentFloor = <-drv_floors

	e.CurrentFloor = e.CurrentFloor
	e.Direction = elevio.MD_Stop
	elevio.SetMotorDirection(e.Direction)

}

func FSM_onButtonPress(b elevio.ButtonEvent) {

	fmt.Printf("%+v\n", b)

	e.Requests = elevio.AddToQueue(b.Button, b.Floor, e.Requests)
	elevio.LightButtons(e.Requests, e.NumFloors)
	e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
	elevio.SetMotorDirection(e.Direction)
	e.Behaviour = elevio.EB_Moving
}

func FSM_doorTimeOut() {
	fmt.Println("\n\nFSM_doorTimeOut()")
	elevio.SetDoorOpenLamp(false) // Close the door light

	if e.Behaviour == elevio.EB_DoorOpen {
		time.Sleep(3 * time.Second)
		if len(e.Requests) == 0 {
			// No requests, set to idle
			e.Behaviour = elevio.EB_Idle
			e.Direction = elevio.MD_Stop
		} else {
			// Choose new direction and behavior
			e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)

			if e.Direction == elevio.MD_Stop {
				e.Behaviour = elevio.EB_Idle
			} else {
				e.Behaviour = elevio.EB_Moving
				elevio.SetMotorDirection(e.Direction) // Start moving
			}
		}
	}

	fmt.Println("\nNew state:")
	fmt.Printf("Floor: %d, Direction: %d, Behaviour: %d\n", e.CurrentFloor, e.Direction, e.Behaviour)
}

func Main_FSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool) {
	init_elevator(drv_floors)

	for {
		select {
		case a := <-drv_buttons:
			FSM_onButtonPress(a)

		case a := <-drv_floors:

			FSM_onFloorArrival(a, drv_buttons)
			FSM_doorTimeOut()

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a && doors.IsDoorOpen {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(e.Direction)
			}

		case a := <-drv_stop:
			fmt.Println("Stopped has been pressed", a)

		}
	}

}
