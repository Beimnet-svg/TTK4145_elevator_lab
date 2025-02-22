package elevator_fsm

import (
	requests "Driver-go/Requests"
	timer "Driver-go/Timer"
	"Driver-go/elevio"
	"fmt"
)

var (
	//Define elevator type
	e = elevio.Elevator{
		CurrentFloor:     0,
		Direction:        elevio.MD_Up,
		Behaviour:        elevio.EB_Idle,
		Requests:         [4][3]int{},
		NumFloors:        4,
		DoorOpenDuration: 3,
		Obstruction:      false,
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
			e.Behaviour = elevio.EB_DoorOpen
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer(e.DoorOpenDuration)
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

	e.Direction = elevio.MD_Stop
	elevio.SetMotorDirection(elevio.MD_Stop)
	fmt.Println("End of init")
	elevio.SetFloorIndicator(e.CurrentFloor)

}

func FSM_onButtonPress(b elevio.ButtonEvent) {

	switch e.Behaviour {
	case elevio.EB_DoorOpen:
		
		if requests.ReqestShouldClearImmideatly(e, b.Floor, b.Button) {
			timer.StartTimer(e.DoorOpenDuration)

		} else {
			e = elevio.AddToQueue(b.Button, b.Floor, e)
		}

	case elevio.EB_Moving:
		e = elevio.AddToQueue(b.Button, b.Floor, e)
		break
	case elevio.EB_Idle:
		e = elevio.AddToQueue(b.Button, b.Floor, e)
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)
		switch e.Behaviour {
		case elevio.EB_Moving:
			elevio.SetMotorDirection(e.Direction)
			break
		case elevio.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer(e.DoorOpenDuration)
			e = requests.RequestClearAtCurrentFloor(e)
			break
		case elevio.EB_Idle:
			break

		}
	}

	elevio.LightButtons(e)

}

func FSM_doorTimeOut() {

	if e.Obstruction && e.Behaviour == elevio.EB_DoorOpen {
		fmt.Println("Door timeout triggered but obstruction is active; keeping door open.")
		timer.StartTimer(e.DoorOpenDuration)
		return
	}

	switch e.Behaviour {
	case elevio.EB_DoorOpen:
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)

		switch e.Behaviour {
		case elevio.EB_DoorOpen:

			timer.StartTimer(e.DoorOpenDuration)
			e = requests.RequestClearAtCurrentFloor(e)
			elevio.LightButtons(e)
			break

		case elevio.EB_Moving:
			fmt.Print("Door is stuck in Eb_moving1\n")

			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)
			fmt.Print("Door is stuck in Eb_moving  2\n")

			break
		case elevio.EB_Idle:
			fmt.Print("Door is stuck in Eb_Idle\n")

			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)
			break
		}
		break
	default:
		fmt.Print("Door isnt opening for orders\n")

	}

}
func FSM_Obstruction(obstruction bool) {
	e.Obstruction = obstruction

	// If the door is open and an obstruction occurs, restart the timer.
	if e.Behaviour == elevio.EB_DoorOpen {
		if obstruction {
			fmt.Println("Obstruction detected while door is open. Keeping door open.")
			timer.StartTimer(e.DoorOpenDuration)
			elevio.SetDoorOpenLamp(true)
		} else {
			fmt.Println("Obstruction cleared while door is open.")
		}
	}
	// When the door is closed, we don't need to do anything.
}

func Main_FSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr chan bool, drv_stop chan bool, doorTimer chan bool) {
	fmt.Println("here")
	init_elevator(drv_floors)

	for {
		select {
		case a := <-drv_buttons:
			FSM_onButtonPress(a)

		case a := <-drv_floors:

			FSM_onFloorArrival(a, drv_buttons)

		case a := <-drv_obstr:
			FSM_Obstruction(a)

		case a := <-drv_stop:
			fmt.Println("Stopped has been pressed", a)

		case a := <-doorTimer:

			if a {
				fmt.Print("After if")
				timer.StopTimer()
				FSM_doorTimeOut()
			}

		}

	}

}
