package elevator_fsm

import (
	doors "Driver-go/Doors"
	elevator_motion "Driver-go/Elevator_Motion"
	"Driver-go/elevio"
	"fmt"
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
	fmt.Printf("Her")
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
	fmt.Println("After loop")

	e.Direction = elevio.MD_Up
	fmt.Println(e.Direction)

	elevio.SetMotorDirection(e.Direction)

	elevio.SetDoorOpenLamp(false)
	e.CurrentFloor = <-drv_floors
	fmt.Println("After set dir")

	e.CurrentFloor = e.CurrentFloor
	e.Direction = elevio.MD_Stop
	elevio.SetMotorDirection(e.Direction)
	fmt.Println("End of init")

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

}

func Main_FSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool) {
	fmt.Println("here")
	init_elevator(drv_floors)

	for {
		select {
		case a := <-drv_buttons:
			FSM_onButtonPress(a)

		case a := <-drv_floors:

			FSM_onFloorArrival(a, drv_buttons)

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
