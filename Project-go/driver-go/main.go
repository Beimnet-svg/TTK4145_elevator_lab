package main

import (
	"Driver-go/elevator_fsm"

	doors "Driver-go/Doors"

	"Driver-go/elevio"
	"fmt"
)

var d elevio.MotorDirection
var numFloors = 4
var currentFloor int

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)
var e = elevio.Elevator{
	CurrentFloor: 0,
	Direction:    elevio.MD_Stop,
	Behaviour:    elevio.EB_Idle,
	Requests:     []elevio.ButtonEvent{},
	NumFloors:    4,
}


var b []elevio.ButtonEvent

func init_elevator(e elevio.Elevator) {

	for a := 0; a < numFloors; a++ {
		for i := elevio.ButtonType(0); i < 3; i++ {
			elevio.SetButtonLamp(i, a, false)
		}
	}

	e.Direction = elevio.MD_Up
	elevio.SetMotorDirection(e.Direction)

	elevio.SetDoorOpenLamp(false)
	e.CurrentFloor = <-drv_floors
	e.CurrentFloor=e.CurrentFloor
	e.Direction = elevio.MD_Stop
	elevio.SetMotorDirection(e.Direction)

}

func main() {

	elevio.Init("localhost:15657", numFloors)
	

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	init_elevator(e)
	for {
		select {
		case a := <-drv_buttons:
		elevator_fsm.FSM_onButtonPress(&e,a )
			

		case a := <-drv_floors:
			
			elevator_fsm.FSM_onFloorArrival(a, &e)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a && doors.IsDoorOpen {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(e.Direction)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
