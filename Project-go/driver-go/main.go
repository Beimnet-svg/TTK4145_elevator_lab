package main

import (
	timer "Driver-go/Timer"
	"Driver-go/elevator_fsm"

	"Driver-go/elevio"
)

var numFloors = 4

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

var doorTimer = make(chan bool)

func main() {

	elevio.Init("localhost:15657", numFloors)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.PollTimer(doorTimer)

	//Networking go routine
	//Acceptence tests
	//1. test if door is closed before running

	go elevator_fsm.Main_FSM(drv_buttons, drv_floors, drv_obstr, drv_stop, doorTimer)

	for {
		//Implement ex 4 in here
	}

}
