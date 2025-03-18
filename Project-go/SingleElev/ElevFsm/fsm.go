package elevfsm

import (
	config "Project-go/Config"
	doortimer "Project-go/SingleElev/Doortimer"
	elevio "Project-go/SingleElev/Elevio"
	requests "Project-go/SingleElev/Requests"
	"fmt"
	"time"
)

var (
	e = elevio.Elevator{
		CurrentFloor:     0,
		Direction:        elevio.MD_Stop,
		Behaviour:        elevio.EB_Moving,
		Requests:         [config.NumberFloors][config.NumberBtn]int{},
		ActiveOrders:     [config.NumberFloors][config.NumberBtn]bool{},
		NumFloors:        config.NumberFloors,
		DoorOpenDuration: config.DoorOpenDuration,
		Master:           false,
		Obstruction:      false,
		Inactive:         true,
	}

	OrderCounter    = 0
	AllActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool
)

func SetElevID() {
	e.ElevatorID = config.ElevID
}

func GetElevator() elevio.Elevator {
	return e
}

func GetAllActiveOrders() [config.NumberElev][config.NumberFloors][config.NumberBtn]bool {
	return AllActiveOrders
}

func CheckInactiveElev(resetInactiveTimer chan int) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			if e.Behaviour == elevio.EB_Idle {
				resetInactiveTimer <- 1
			}
		}
	}
}

func InitElevator(drv_floors chan int) {

	for a := 0; a < e.NumFloors; a++ {
		for i := elevio.ButtonType(0); i < config.NumberBtn; i++ {
			elevio.SetButtonLamp(i, a, false)
		}
	}

	elevio.SetMotorDirection(elevio.MD_Up)

	elevio.SetDoorOpenLamp(false)
	e.CurrentFloor = <-drv_floors

	e.Direction = elevio.MD_Stop
	e.Behaviour = elevio.EB_Idle
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(e.CurrentFloor)

}

func fsmOnFloorArrival(floor int, resetInactiveTimer chan int) {

	elevio.SetFloorIndicator(floor)
	e.CurrentFloor = floor

	switch e.Behaviour {
	case elevio.EB_Moving:
		if requests.RequestShouldStop(e) {
			resetInactiveTimer <- 1
			elevio.SetMotorDirection(elevio.MD_Stop)
			e.Behaviour = elevio.EB_DoorOpen
			elevio.SetDoorOpenLamp(true)
			doortimer.StartDoorTimer(e.DoorOpenDuration)
		}

	default:

	}

}

func fsmOnMsgArrived(orders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, resetInactiveTimer chan int) {

	//Store all active orders in the network in case of master failure
	AllActiveOrders = orders
	e.ActiveOrders = orders[e.ElevatorID]

	switch e.Behaviour {
	case elevio.EB_Idle:
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)
		switch e.Behaviour {
		case elevio.EB_Moving:
			resetInactiveTimer <- 1
			elevio.SetMotorDirection(e.Direction)

		case elevio.EB_DoorOpen:
			resetInactiveTimer <- 1
			elevio.SetDoorOpenLamp(true)
			doortimer.StartDoorTimer(e.DoorOpenDuration)

		case elevio.EB_Idle:
		}
	default:

	}
	elevio.LightButtons(AllActiveOrders, e.ElevatorID)
}

func fsmOnButtonPress(b elevio.ButtonEvent) {
	if e.Behaviour == elevio.EB_DoorOpen && requests.ReqestShouldClearImmideatly(e, b.Floor, b.Button) {
		doortimer.StartDoorTimer(e.DoorOpenDuration)

	} else {
		OrderCounter += 1
		e = elevio.AddToQueue(b.Button, b.Floor, e, OrderCounter)
		fmt.Println(e.Requests)
	}
}

func fsmOnObstruction(obstruction bool) {
	e.Obstruction = obstruction

	if e.Behaviour == elevio.EB_DoorOpen {
		if obstruction {
			doortimer.StartDoorTimer(e.DoorOpenDuration)
			elevio.SetDoorOpenLamp(true)
		}
	}
}

func fsmOnDoorTimeOut(resetInactiveTimer chan int) {

	if e.Obstruction && e.Behaviour == elevio.EB_DoorOpen {
		doortimer.StartDoorTimer(e.DoorOpenDuration)
		return
	}

	switch e.Behaviour {
	case elevio.EB_DoorOpen:
		e.Direction, e.Behaviour = requests.RequestChooseDir(e)

		switch e.Behaviour {
		case elevio.EB_DoorOpen:
			resetInactiveTimer <- 1
			doortimer.StartDoorTimer(e.DoorOpenDuration)

		case elevio.EB_Moving:
			resetInactiveTimer <- 1
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)

		case elevio.EB_Idle:

			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Direction)

		}

	}

}

func MainFsm(drvButtons chan elevio.ButtonEvent, drvFloors chan int,
	drvObstr chan bool, drvStop chan bool, doorTimer chan bool,
	activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool, elevInactive chan bool, resetInactiveTimer chan int) {

	for {
		select {
		case a := <-drvButtons:
			fsmOnButtonPress(a)

		case a := <-drvFloors:

			fsmOnFloorArrival(a, resetInactiveTimer)

		case a := <-drvObstr:
			fsmOnObstruction(a)

		case a := <-drvStop:
			fmt.Println("Stopped has been pressed", a)

		case a := <-doorTimer:

			if a {
				doortimer.StopDoorTimer()
				fsmOnDoorTimeOut(resetInactiveTimer)
			}
		case a := <-elevInactive:
			e.Inactive = a
		case a := <-activeOrdersArrived:

			fsmOnMsgArrived(a, resetInactiveTimer)
		case a := <-setMaster:
			e.Master = a
			//Store all orders a disconnected master had as new requests that will be processed by the new master
			if !a {
				e.Requests = [config.NumberFloors][config.NumberBtn]int{}
				for floor := 0; floor < e.NumFloors; floor++ {
					for button := 0; button < config.NumberBtn; button++ {
						if e.ActiveOrders[floor][button] {
							e.Requests[floor][button] = OrderCounter
						}
					}
				}
				fmt.Println("Requests", e.Requests)
			}
		}
	}
}
