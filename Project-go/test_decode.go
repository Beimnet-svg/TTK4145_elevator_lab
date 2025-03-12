package main

import (
	config "Project-go/Config"
	"Project-go/driver-go/elevio"
	"bytes"
	"encoding/gob"
	"fmt"
)

var E = elevio.Elevator{
	CurrentFloor:     0,
	Direction:        elevio.MD_Up,
	Behaviour:        elevio.EB_Idle,
	Requests:         [config.NumberFloors][config.NumberBtn]int{},
	ActiveOrders:     [config.NumberFloors][config.NumberBtn]bool{},
	NumFloors:        config.NumberFloors,
	DoorOpenDuration: config.DoorOpenDuration,
	Master:           false,
	Obstruction:      false,
}

type OrderMessageSlave struct {
	ElevID int
	Master bool
	E      elevio.Elevator
}

type OrderMessageMaster struct {
	ElevID int
	Master bool
	Orders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool
}

type OrderMessage struct {
	Slave  *OrderMessageSlave
	Master *OrderMessageMaster
}

func init() {
	gob.Register(OrderMessageSlave{})
	gob.Register(OrderMessageMaster{})
	gob.Register(OrderMessage{})
}

// Make struct into byte slice
func decodeMessage(buffer []byte) (*OrderMessage, error) {
	buf := bytes.NewBuffer(buffer)
	dec := gob.NewDecoder(buf)
	var message OrderMessage
	err := dec.Decode(&message)
	return &message, err
}

func encodeMessage() []byte{
	E.Requests = [config.NumberFloors][config.NumberBtn]int{
		{1, 0, 0},
		{0, 2, 0},
		{0, 4, 0},
		{5, 0, 0},
	}
	E.ElevatorID = 1
	message := OrderMessage{
		Slave: &OrderMessageSlave{
			ElevID: E.ElevatorID,
			Master: false,
			E:      E,
		},
	}
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(message)
	if err != nil {
		fmt.Println("Error encoding orders:", err)
	}
	fmt.Println(message.Slave.E.Requests)
	content := buffer.Bytes()
	
	return content
}

func main(){

	messageBytes := encodeMessage()
	
	msg, _ := decodeMessage(messageBytes)
	fmt.Println(msg.Slave.E.Requests)
	fmt.Println(msg.Slave.E.ElevatorID)
}