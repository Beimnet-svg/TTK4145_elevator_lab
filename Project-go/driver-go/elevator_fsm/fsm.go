package elevator_fsm

import (
	doors "Driver-go/Doors"
	elevator_motion "Driver-go/Elevator_Motion"
	"Driver-go/elevio"
	"fmt"
)

var (
	 //Define elevator type
	DRV_buttons = make(chan elevio.ButtonEvent)
	
)



func FSM_onFloorArrival(floor int, e *elevio.Elevator) {

	elevio.SetFloorIndicator(floor)
	e.CurrentFloor = floor
	fmt.Printf("Her")
	e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
	elevio.SetMotorDirection(e.Direction)

	fmt.Printf("%+v\n", floor)
	e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
	elevio.SetMotorDirection(e.Direction)

	if e.Direction == elevio.MD_Stop {
		e.Requests = doors.OpenDoor( e.CurrentFloor, e.Requests, e.NumFloors, DRV_buttons)
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

func FSM_onButtonPress(e *elevio.Elevator, b elevio.ButtonEvent) {
	
	fmt.Printf("%+v\n", b)

	e.Requests = elevio.AddToQueue(b.Button, b.Floor, e.Requests)
	elevio.LightButtons(e.Requests, e.NumFloors)
	e.Direction = elevator_motion.SetDirection(e.CurrentFloor, e.Requests[0].Floor, e.Direction)
	elevio.SetMotorDirection(e.Direction)
	e.Behaviour = elevio.EB_Moving
}