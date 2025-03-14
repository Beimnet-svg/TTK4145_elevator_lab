package elevator_fsm

import (
	config "Project-go/Config"
	requests "Project-go/driver-go/Requests"
	timer "Project-go/driver-go/Timer"
	"Project-go/driver-go/elevio"
	"fmt"
)

var (
	e = elevio.Elevator{
		CurrentFloor:     0,
		Direction:        elevio.MD_Up,
		Behaviour:        elevio.EB_Idle,
		Requests:         [config.NumberFloors][config.NumberBtn]int{},
		ActiveOrders:     [config.NumberFloors][config.NumberBtn]bool{},
		NumFloors:        config.NumberFloors,
		DoorOpenDuration: config.DoorOpenDuration,
		Master:           false,
		Obstruction:      false,
	}
	OrderCounter = 0
)

var AllActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool

func GetElevator() elevio.Elevator {
	return e
}

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

	default:

	}

}

func FSM_onMsgArrived(orders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {

	e.ActiveOrders = orders[e.ElevatorID]
	//Store all active orders in the network in case of master failure
	AllActiveOrders = orders

	switch e.Behaviour {
	case elevio.EB_Idle:
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)
		switch e.Behaviour {
		case elevio.EB_Moving:
			elevio.SetMotorDirection(e.Direction)

		case elevio.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer(e.DoorOpenDuration)

		case elevio.EB_Idle:
		}
	default:

	}
	elevio.LightButtons(e)
}

func init_elevator(drv_floors chan int) {
	// Take input argument from terminal as elevator ID

	for a := 0; a < e.NumFloors; a++ {
		for i := elevio.ButtonType(0); i < config.NumberBtn; i++ {
			elevio.SetButtonLamp(i, a, false)
		}
	}

	fmt.Println("After loop")

	e.Direction = elevio.MD_Up
	fmt.Println(e.Direction)

	elevio.SetMotorDirection(e.Direction)

	elevio.SetDoorOpenLamp(false)
	e.CurrentFloor = <-drv_floors

	e.Direction = elevio.MD_Stop
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(e.CurrentFloor)

}

func SetElevatorID(ID int) {
	e.ElevatorID = ID
}

func FSM_onButtonPress(b elevio.ButtonEvent) {
	if e.Behaviour == elevio.EB_DoorOpen && requests.ReqestShouldClearImmideatly(e, b.Floor, b.Button) {
		timer.StartTimer(e.DoorOpenDuration)
	} else {
		// Ordercounter is used to ensure that new requests can be distinguised from old ones
		// Instead of deleting requests, we only use the ones with counter values higher than the latest used in the order manager
		OrderCounter += 1
		e = elevio.AddToQueue(b.Button, b.Floor, e, OrderCounter)
		fmt.Println(e.Requests)
	}
}

func FSM_Obstruction(obstruction bool) {
	e.Obstruction = obstruction

	if e.Behaviour == elevio.EB_DoorOpen {
		if obstruction {
			timer.StartTimer(e.DoorOpenDuration)
			elevio.SetDoorOpenLamp(true)
		}
	}
}

func FSM_doorTimeOut() {

	if e.Obstruction && e.Behaviour == elevio.EB_DoorOpen {
		timer.StartTimer(e.DoorOpenDuration)
		return
	}

	switch e.Behaviour {
	case elevio.EB_DoorOpen:
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)

		switch e.Behaviour {
		case elevio.EB_DoorOpen:
			timer.StartTimer(e.DoorOpenDuration)

		case elevio.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)

		case elevio.EB_Idle:

			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)

		}

	}

}

func Main_FSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int,
	drv_obstr chan bool, drv_stop chan bool, doorTimer chan bool,
	msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {

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
				timer.StopTimer()
				FSM_doorTimeOut()
			}
		case a := <-msgArrived:

			FSM_onMsgArrived(a)
		case a := <-setMaster:
			fmt.Print("Master has been set to: ", a, "\n")
			e.Master = a
			
		}

	}

}
