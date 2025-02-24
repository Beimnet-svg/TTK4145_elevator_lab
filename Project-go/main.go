package main

import (
	networking "Project-go/Networking"
	timer "Project-go/driver-go/Timer"
	"Project-go/driver-go/elevator_fsm"

	"Project-go/driver-go/elevio"
)

var numFloors = 4

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

var doorTimer = make(chan bool)
var msgArrived = make(chan [4][3][3]bool)

func main() {

	elevio.Init("localhost:15657", numFloors)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.PollTimer(doorTimer)
	go networking.Receiver(msgArrived)

	//Networking go routine
	//Acceptence tests
	//1. test if door is closed before running

	go elevator_fsm.Main_FSM(drv_buttons, drv_floors, drv_obstr,
		drv_stop, doorTimer, msgArrived)

	for {
		//Send alive message (Elevator nr + Slave or not)
	}

}
