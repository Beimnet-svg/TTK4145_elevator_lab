package main

import (
	networking "Project-go/Networking"
	timer "Project-go/driver-go/Timer"
	"Project-go/driver-go/elevator_fsm"

	"Project-go/driver-go/elevio"
)

//Hva skal skje:
//-Hver gang allorders blir updated skal den polle msg arrived i fsm
//-En annen thread holder styr på I am alive
var numFloors = 4

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

var doorTimer = make(chan bool)
var orderArrived = make(chan [3][4][3]bool)
var heartbeat = make(chan networking.HeartbeatMessage)

func main() {

	

	elevio.Init("localhost:15657", numFloors)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.PollTimer(doorTimer)
	go networking.UnifiedReceiver(orderArrived, heartbeat)
	go networking.HeartBeatSender(elevator_fsm.GetElevator())

	

	//Networking go routine
	//Acceptence tests
	//1. test if door is closed before running

	go elevator_fsm.Main_FSM(drv_buttons, drv_floors, drv_obstr,
		drv_stop, doorTimer, orderArrived)


	for {
		
			//Reset timer everytime you get a master alive; for slaves  or 
			//Update which elevators are alive, for master
			//Update master if master dies.
		
	}

}
