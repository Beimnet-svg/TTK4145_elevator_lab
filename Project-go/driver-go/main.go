package main

import (
	"Driver-go/elevator_fsm"

	"Driver-go/elevio"
)

var d elevio.MotorDirection
var numFloors = 4
var currentFloor int

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

var b []elevio.ButtonEvent

func main() {

	elevio.Init("localhost:15657", numFloors)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	//Networking go routine

	go elevator_fsm.Main_FSM(drv_buttons, drv_floors, drv_obstr, drv_stop)

	for {
		//Implement ex 4 in here
	}

}
