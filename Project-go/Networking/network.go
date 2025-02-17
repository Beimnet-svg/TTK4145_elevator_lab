package networking

import (
	"Project-go/driver-go/elevio"
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"strconv"
)

func OrderRecieved(buffer []byte) elevio.Elevator {
	var e elevio.Elevator
	buf := bytes.NewBuffer(buffer)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&e)
	if err != nil {
		fmt.Println("Error decoding buffer:", err)
	}
	return e

}

func Reciever() {
	//check first 2 digits, when checking we now if it is alive or order
	conn, _ := net.ListenPacket("udp", ":20007")
	defer conn.Close()

	buffer := make([]byte, 1024)
	fmt.Println("Listening for UDP on port 20007........")
	for {
		n, addr, _ := conn.ReadFrom(buffer)

		if buffer[1] != 2 {
			//Send two first elements of buffer to master-slave-dist
		} else {
			e := OrderRecieved(buffer)
			//Based on 3rd element in buffer, decide what type of order it is(requests, currently served orders or all orders)
			switch e.Master {
			case 0:
				//New requests or serviced orders
				//Send e to order manager, use 3rd buffer element to decide if it is new requests(0) or serviced orders(1)
			case 1:
				//New set of orders or all orders
				//If 3rd element is 1 we save them, if it is 0 we set as new orders

			}
		}

		fmt.Printf("Received %s from %s\n", string(buffer[0:n]), addr)
	}

	//switch case for deciding if it is alive or order, if alive handle here, if order, call order

}

func Sender(e elevio.Elevator, alive bool) {
	//Call this when we want to send a message

	serverAddr := ":20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	if alive {
		message := strconv.Itoa(e.ElevatorID) + strconv.Itoa(e.Master)
		content := []byte(message)
		conn.Write(content)

	} else {
		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		err := enc.Encode(e)
		if err != nil {
			fmt.Println("Error encoding elevator:", err)
		}
		content := buffer.Bytes()
		conn.Write(content)
	}
	

}

func Main_Networking() {
	//main
	go Reciever()
}
