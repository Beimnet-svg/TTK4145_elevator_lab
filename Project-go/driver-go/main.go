package main

import (
	elevator_motion "Driver-go/Elevator_Motion"
	"Driver-go/elevio"
	"fmt"
)

var d elevio.MotorDirection
var floor int

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

func init() {
	d = elevio.MD_Up
	go elevio.PollFloorSensor(drv_floors)
	for {
		floor = <-drv_floors
		if floor != -1 {
			d = elevio.MD_Stop
			elevio.SetMotorDirection(d)
			break
		}
	}
}

func main() {

	numFloors := 4

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
			fmt.Printf("Button pressed: %+v\n", a)
			if a.Floor > floor {
				elevator_motion.MoveUp(floor, a.Floor, drv_floors)
			} else if a.Floor < floor {
				elevator_motion.MoveDown(floor, a.Floor, drv_floors)
			}
			floor = a.Floor
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
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
