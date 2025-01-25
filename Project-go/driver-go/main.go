package main

import (

	elevator_motion "Driver-go/Elevator_Motion"

	"Driver-go/doors"

	"Driver-go/elevio"
	"fmt"
)

var d elevio.MotorDirection

var currentFloor int


var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

var b = make([]elevio.ButtonEvent, 10)


func init_elevator() {
	go elevio.PollFloorSensor(drv_floors)
	d = elevio.MD_Up
	elevio.SetMotorDirection(d)


	currentFloor = <-drv_floors
	d = elevio.MD_Stop
	elevio.SetMotorDirection(d)

}

func main() {


	numFloors := 4

	init_elevator()

	elevio.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)


	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.AddToQueue(a.Button, a.Floor, b)
			elevio.LightButtons(b, numFloors)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			currentFloor = a
			elevio.RemoveFromQueue(currentFloor, d, b)
			d = elevator_motion.SetDirection(currentFloor, b[0].Floor, d)
			elevio.SetMotorDirection(d)

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
