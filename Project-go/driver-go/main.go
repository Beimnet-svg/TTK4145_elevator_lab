package main

import (
	"Driver-go/doors"
	"Driver-go/elevio"
	"fmt"
)

var d elevio.MotorDirection


var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors  = make(chan int)
var drv_obstr   = make(chan bool)
var drv_stop    = make(chan bool)    



// func init(){

//     d = elevio.MD_Up

//     for {
//         go elevio.PollFloorSensor(drv_floors)
//         floor = <- drv_floors
//         if floor != -1 {
//             d = elevio.MD_Stop
//         }
//         elevio.SetMotorDirection(d)
//     }
    
// }

func main(){

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
        case a := <- drv_buttons:
            fmt.Printf("%+v\n", a)
            elevio.SetButtonLamp(a.Button, a.Floor, true)
            
            
        case a := <- drv_floors:
            fmt.Printf("%+v\n", a)
            doors.OpenDoor(drv_floors)
            doors.CloseDoor(drv_floors,drv_obstr)

            if a == numFloors-1 {
                d = elevio.MD_Down
            } else if a == 0 {
                d = elevio.MD_Up
                
            }
            elevio.SetMotorDirection(d)
            
            
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                    //If between floors, dont open doors
                }
            }
        }
    }    
}
