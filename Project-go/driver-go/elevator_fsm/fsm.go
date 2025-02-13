package elevator_fsm

import (
	doors "Driver-go/Doors"
	elevator_motion "Driver-go/Elevator_Motion"
	requests "Driver-go/Requests"
	"Driver-go/elevio"
	"fmt"
)

var (
	//Define elevator type
	e = elevio.Elevator{
		CurrentFloor: 0,
		Direction:    elevio.MD_Stop,
		Behaviour:    elevio.EB_Idle,
		Requests:     [4][3]int{},
		NumFloors:    4,
	}
)

func FSM_onFloorArrival(floor int, drv_button chan elevio.ButtonEvent) {

	elevio.SetFloorIndicator(floor)
	e.CurrentFloor = floor

	switch e.Behaviour {
		case elevio.EB_Moving:
			if requests.RequestShouldStop(e) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e = requests.RequestClearAtCurrentFloor(e)
				elevio.LightButtons(e)
				e.Direction = elevio.MD_Stop
				e.Behaviour = elevio.EB_DoorOpen
				elevio.SetDoorOpenLamp(true)
				//Start timer
			}
			break
		default:
			break

	}	

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

	switch e.Behaviour {
		case elevio.EB_DoorOpen:
			if requests.ReqestShouldClearImmideatly(e, b.Floor, b.Button) {
				//Reset timer
			} else {
				e = elevio.AddToQueue(b.Button, b.Floor, e)
			}

		case elevio.EB_Moving:
			e = elevio.AddToQueue(b.Button, b.Floor, e)
			break
		case elevio.EB_Idle:
			e = elevio.AddToQueue(b.Button, b.Floor, e)
			e.Direction, e.Behaviour = requests.RequestChooseDir(e, b.Button)
			switch e.Behaviour {
				case elevio.EB_Moving:
					elevio.SetMotorDirection(e.Direction)
					break
				case elevio.EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					//Start timer
					e = requests.RequestClearAtCurrentFloor(e)
					break
				case elevio.EB_Idle:
					break
		
			}
	}

	elevio.LightButtons(e)
	
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
