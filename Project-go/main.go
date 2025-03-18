package main

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	networking "Project-go/Networking"
	ordermanager "Project-go/OrderManager"
	doortimer "Project-go/SingleElev/Doortimer"
	elevfsm "Project-go/SingleElev/ElevFsm"
	elevio "Project-go/SingleElev/Elevio"
	"os"
	"strconv"
)

var (
	drvButtons = make(chan elevio.ButtonEvent)
	drvFloors  = make(chan int)
	drvObstr   = make(chan bool)
	drvStop    = make(chan bool)
	doorTimer  = make(chan bool)

	activeOrdersArrived = make(chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool)

	setMaster          = make(chan bool)
	elevDied           = make(chan int)
	elevInactive       = make(chan bool)
	resetInactiveTimer = make(chan int)
	setDisconnected    = make(chan bool)
)

func main() {

	elevio.Init("localhost:15657", config.NumberFloors)

	ID := os.Args[1]
	ID_e, _ := strconv.Atoi(ID)
	config.SetElevID(ID_e)
	elevfsm.SetElevID()

	go elevfsm.InitElevator(drvFloors)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
	go doortimer.PollDoorTimer(doorTimer)

	go networking.Receiver(activeOrdersArrived, setMaster)
	go networking.Sender(activeOrdersArrived, setDisconnected)

	go elevfsm.CheckInactiveElev(resetInactiveTimer)
	go masterslavedist.WatchdogTimer(setMaster, elevDied, elevInactive)
	go masterslavedist.ResetInactiveTimer(resetInactiveTimer, elevInactive)
	go masterslavedist.CheckMasterTimerTimeout()
	go masterslavedist.SetDisconnected(setDisconnected)

	go ordermanager.ApplyBackupOrders(setMaster, activeOrdersArrived)
	go ordermanager.ResetOrderCounter(elevDied)

	go networking.Print() // To be deleted

	go elevfsm.MainFsm(drvButtons, drvFloors, drvObstr,
		drvStop, doorTimer, activeOrdersArrived, setMaster, elevInactive, resetInactiveTimer)

	myelevator := elevfsm.GetElevator()
	go masterslavedist.InitializeMasterSlaveDist(myelevator, activeOrdersArrived, setMaster)

	for {

	}

}
