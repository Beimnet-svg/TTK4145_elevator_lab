package elevator_fsm

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	requests "Project-go/driver-go/Requests"
	timer "Project-go/driver-go/Timer"
	"Project-go/driver-go/elevio"
	"fmt"
	"os"
	"strconv"
)

var (
	//Define elevator type
	e = elevio.Elevator{
		CurrentFloor:     0,
		Direction:        elevio.MD_Up,
		Behaviour:        elevio.EB_Idle,
		Requests:         [config.NumberFloors][config.NumberBtn]int{},
		ActiveOrders:     [config.NumberFloors][config.NumberBtn]bool{},
		NumFloors:        config.NumberFloors,
		DoorOpenDuration: config.DoorOpenDuration,
		Master:           false,
		Obstruction: 	  false,
	}
	OrderCounter = 0
)

var allActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool

func GetElevator() *elevio.Elevator {
	return &e
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
	allActiveOrders = orders

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
	ID := os.Args[1]

	fmt.Print("ID: ", ID)

	e.ElevatorID, _ = strconv.Atoi(ID)

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
			fmt.Print("Door is stuck in Eb_dooropen\n")

			timer.StartTimer(e.DoorOpenDuration)

		case elevio.EB_Moving:
			fmt.Print("Door is stuck in Eb_moving1\n")

			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)
			fmt.Print("Door is stuck in Eb_moving  2\n")

		case elevio.EB_Idle:
			fmt.Print("Door is stuck in Eb_Idle\n")

			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)

		}

	default:
		fmt.Print("Door isnt opening for orders\n")

	}

}

func Main_FSM(drv_buttons chan elevio.ButtonEvent, drv_floors chan int,
	drv_obstr chan bool, drv_stop chan bool, doorTimer chan bool,
	msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {

	fmt.Println("here")
	init_elevator(drv_floors)
	masterslavedist.InitializeMasterSlaveDist(&e)

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
		case a := <-msgArrived:
			//Add ignore messages on same IP
			fmt.Println("At message arrived: \n", a)
			FSM_onMsgArrived(a)
		case a := <-setMaster:
			e.Master = a

		}

	}

}
