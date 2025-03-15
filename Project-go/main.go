package main

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	networking "Project-go/Networking"
	ordermanager "Project-go/OrderManager"
	timer "Project-go/driver-go/Timer"
	"Project-go/driver-go/elevator_fsm"
	"os"
	"strconv"

	"Project-go/driver-go/elevio"
)

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

var doorTimer = make(chan bool)
var activeOrdersArrived = make(chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool)
var setMaster = make(chan bool)
var elevDied = make(chan int)

var elevInactive = make(chan bool)
var resetInactiveTimer = make(chan int)

func main() {

	elevio.Init("localhost:15657", config.NumberFloors)
	ID := os.Args[1]

	ID_e, _ := strconv.Atoi(ID)

	elevator_fsm.SetElevatorID(ID_e)

	go elevator_fsm.Init_elevator(drv_floors)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.PollTimer(doorTimer)
	go elevator_fsm.CheckInactiveElev(resetInactiveTimer)

	go networking.Receiver(activeOrdersArrived, setMaster)
	go networking.Sender(activeOrdersArrived)

	go masterslavedist.WatchdogTimer(setMaster, elevDied, elevInactive)
	go masterslavedist.ResetInactiveTimer(resetInactiveTimer, elevInactive)
	go masterslavedist.CheckMasterTimerTimeout()
	go ordermanager.ApplyBackupOrders(setMaster, activeOrdersArrived)
	go ordermanager.ResetOrderCounter(elevDied)

	go networking.Print()

	//Networking go routine
	//Acceptence tests
	//1. test if door is closed before running

	go elevator_fsm.Main_FSM(drv_buttons, drv_floors, drv_obstr,
		drv_stop, doorTimer, activeOrdersArrived, setMaster, elevInactive, resetInactiveTimer)

	myelevator := elevator_fsm.GetElevator()
	go masterslavedist.InitializeMasterSlaveDist(myelevator, activeOrdersArrived, setMaster)

	for {

	}

}
