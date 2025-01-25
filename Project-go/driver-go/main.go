package main

import (
	elevator_motion "Driver-go/Elevator_Motion"

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

var b []elevio.ButtonEvent

func init_elevator() {

	for a := 0; a < numFloors; a++ {
		for i := elevio.ButtonType(0); i < 3; i++ {
			elevio.SetButtonLamp(i, a, false)
		}
	}

	d = elevio.MD_Up
	elevio.SetMotorDirection(d)

	elevio.SetDoorOpenLamp(false)
	currentFloor = <-drv_floors
	d = elevio.MD_Stop
	elevio.SetMotorDirection(d)

}

func main() {

	elevio.Init("localhost:15657", numFloors)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	init_elevator()
	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)

			b = elevio.AddToQueue(a.Button, a.Floor, b)
			elevio.LightButtons(b, numFloors)
			d = elevator_motion.SetDirection(currentFloor, b[0].Floor, d)
			elevio.SetMotorDirection(d)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			currentFloor = a
			d = elevator_motion.SetDirection(currentFloor, b[0].Floor, d)
			elevio.SetMotorDirection(d)

			if d == elevio.MD_Stop {
				b = doors.OpenDoor(b[0].Floor, currentFloor, b, numFloors, drv_buttons)
				b = elevio.RemoveFromQueue(currentFloor, d, b)
				elevio.LightButtons(b, numFloors)
				if len(b) != 0 {
					d = elevator_motion.SetDirection(currentFloor, b[0].Floor, d)
					elevio.SetMotorDirection(d)
				}

			}

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
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
