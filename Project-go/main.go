package main

import (
	networking "Project-go/Networking"
	timer "Project-go/driver-go/Timer"
	"Project-go/driver-go/elevator_fsm"
	"fmt"
	"time"

	"Project-go/driver-go/elevio"
)

//Hva skal skje:
//-Hver gang allorders blir updated skal den polle msg arrived i fsm
//-En annen thread holder styr på I am alive
//Fortsett med å finne ut hvor sendUpdates skal være, hver gang master eller slave gjør en update.
var numFloors = 4

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)

var doorTimer = make(chan bool)
var orderArrived = make(chan [3][4][3]bool)
var heartbeat = make(chan networking.HeartbeatMessage)
var stateUpdateChan= make(chan elevio.Elevator)
// var ackChan= make(chan int)

func main() {

	
	myElevator:=elevator_fsm.GetElevator()
	elevio.Init("localhost:15657", numFloors)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.PollTimer(doorTimer)
	go networking.UnifiedReceiver(orderArrived, heartbeat, stateUpdateChan )

	//Heartbeat senders
	if myElevator.Master{
		go networking.HeartbeatSenderMaster(myElevator)
	}else{
		go networking.HeartbeatSenderSlave(myElevator)
	}
	

	//Networking go routine
	//Acceptence tests
	//1. test if door is closed before running

	go elevator_fsm.Main_FSM(drv_buttons, drv_floors, drv_obstr,
		drv_stop, doorTimer, orderArrived)


		heartbeatTimeout := 2000 * time.Millisecond // Time threshold for heartbeat loss
		heartbeatTimer := time.NewTimer(heartbeatTimeout) // Initial timer
	
		for {
			select {
			case hb := <-heartbeat:
				// Reset timer on heartbeat reception
				if !heartbeatTimer.Stop() {
					<-heartbeatTimer.C 
				}
				heartbeatTimer.Reset(heartbeatTimeout)
				if myElevator.Master {
					fmt.Println("Received heartbeat from:", hb.ElevID)
					// Update active elevators list, if necessary
				} else if hb.Master {
					fmt.Println("Received heartbeat from master")
				}
	
			case <-heartbeatTimer.C:
				// No heartbeat received within the timeout → assume failure
				myElevator.Master = true // Promote to master
				go networking.HeartbeatSenderMaster(myElevator) 
			}
		}

}
