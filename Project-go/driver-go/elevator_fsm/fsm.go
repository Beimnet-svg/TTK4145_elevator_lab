package elevator_fsm

import (
	networking "Project-go/Networking"
	requests "Project-go/driver-go/Requests"
	timer "Project-go/driver-go/Timer"
	"Project-go/driver-go/elevio"
	"fmt"
)

var (
	//Define elevator type
	e = elevio.Elevator{
		CurrentFloor:     0,
		Direction:        elevio.MD_Up,
		Behaviour:        elevio.EB_Idle,
		Requests:         [4][3]int{},
		ActiveOrders:     [4][3]bool{},
		NumFloors:        4,
		DoorOpenDuration: 3,
		ElevatorID:       0,
		Master:           true,
	}
	OrderCounter = 0
)

var allActiveOrders [4][3][3]bool

func FSM_onFloorArrival(floor int, drv_button chan elevio.ButtonEvent) {

	elevio.SetFloorIndicator(floor)
	e.CurrentFloor = floor

	switch e.Behaviour {
	case elevio.EB_Moving:
		if requests.RequestShouldStop(e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			e.Behaviour = elevio.EB_DoorOpen
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer(e.DoorOpenDuration)
		}
		break
	default:
		break

	}

}

func FSM_onMsgArrived(orders [4][3][3]bool) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			e.ActiveOrders[i][j] = orders[i][j][e.ElevatorID]
		}
	}
	allActiveOrders = orders
	switch e.Behaviour {
	case elevio.EB_Idle:
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)
		switch e.Behaviour {
		case elevio.EB_Moving:
			elevio.SetMotorDirection(e.Direction)
			break
		case elevio.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer(e.DoorOpenDuration)
			break
		case elevio.EB_Idle:
			break
		}
	default:
		break
	}
	elevio.LightButtons(e)
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
	if e.Behaviour == elevio.EB_DoorOpen && requests.ReqestShouldClearImmideatly(e, b.Floor, b.Button) {
		timer.StartTimer(e.DoorOpenDuration)
	} else {
		OrderCounter += 1
		e = elevio.AddToQueue(b.Button, b.Floor, e, OrderCounter)
	}
}

func FSM_doorTimeOut() {

	switch e.Behaviour {
	case elevio.EB_DoorOpen:
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)

		switch e.Behaviour {
		case elevio.EB_DoorOpen:
			fmt.Print("Door is stuck in Eb_dooropen\n")

			timer.StartTimer(e.DoorOpenDuration)
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

func Main_FSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int,
	drv_obstr chan bool, drv_stop chan bool, doorTimer chan bool,
	msgArrived chan [4][3][3]bool) {

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
			if a && e.Behaviour == elevio.EB_DoorOpen {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(e.Direction)
			}

		case a := <-drv_stop:
			fmt.Println("Stopped has been pressed", a)

		case a := <-doorTimer:

			if a {
				fmt.Print("After if")
				timer.StopTimer()
				FSM_doorTimeOut()
			}
		case a := <-msgArrived:
			//Add ignore messages on same IP
			fmt.Println("At message arrived: \n", a)
			FSM_onMsgArrived(a)
		default:
			switch e.Master {
			case false:

				networking.SenderSlave(e)

			case true:

				networking.SenderMaster(e, allActiveOrders)
			}

		}

	}

}
