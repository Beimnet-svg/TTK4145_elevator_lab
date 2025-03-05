package networking

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	ordermanager "Project-go/OrderManager"
	"Project-go/driver-go/elevator_fsm"
	"Project-go/driver-go/elevio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"time"
)

type OrderMessageSlave struct {
	ElevID int
	Master bool
	e      elevio.Elevator
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

func decodeMessage(buffer []byte) (*OrderMessage, error) {
	buf := bytes.NewBuffer(buffer)
	dec := gob.NewDecoder(buf)
	var message OrderMessage
	err := dec.Decode(&message)
	return &message, err
}

func Sender(msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		localElev := elevator_fsm.GetElevator()

		if masterslavedist.Disconnected {
			ordermanager.UpdateOrders(*localElev, msgArrived)
			continue
		}

		if localElev.Master {
			orders := ordermanager.AllActiveOrders
			SenderMaster(*localElev, orders)
		} else {
			SenderSlave(*localElev)
		}
		if localElev.Master {
			ordermanager.UpdateOrders(*localElev, msgArrived)
		}
	}
}

func Receiver(msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	// Listen for incoming UDP packets on port 20007
	conn, err := net.ListenPacket("udp", ":20007")
	if err != nil {
		log.Fatal("Error listening on port 20007:", err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		// Wait for a message to be received
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Fatal("Error reading from connection:", err)
		}

		// Decode the received message
		msg, err := decodeMessage(buffer[:n])
		if err != nil {
			log.Fatal("Error decoding message:", err)
		}
		fmt.Println("msg in Reciever: \n", msg)
		localElev := elevator_fsm.GetElevator()

		// Process the received message
		if msg.Slave != nil {
			ordermanager.UpdateOrders(msg.Slave.e, msgArrived)
			masterslavedist.AliveRecieved(msg.Slave.ElevID, msg.Slave.Master, localElev)
		} else if msg.Master != nil {
			masterslavedist.AliveRecieved(msg.Master.ElevID, msg.Master.Master, localElev)
			msgArrived <- msg.Master.Orders
		}
	}
}
func SenderSlave(e elevio.Elevator) {
	//Call this when we want to send a message

	// Create an instance of the struct
	message := OrderMessage{
		Slave: &OrderMessageSlave{
			ElevID: e.ElevatorID,
			Master: false,
			e:      e,
		},
	}

	// Call this when we want to send a message
	serverAddr := ":20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(message)
	if err != nil {
		fmt.Println("Error encoding orders:", err)
	}
	content := buffer.Bytes()
	conn.Write(content)

}

func SenderMaster(e elevio.Elevator, orders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	//Call this when we want to send a message

	// Create an instance of the struct
	message := OrderMessage{
		Master: &OrderMessageMaster{
			ElevID: e.ElevatorID,
			Master: true,
			Orders: orders,
		},
	}

	serverAddr := ":20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	//Master sending out orders to all elevators, including which elev should take it
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(message)
	if err != nil {
		fmt.Println("Error encoding orders:", err)
	}
	content := buffer.Bytes()
	conn.Write(content)
	//Master sending out orders to all elevators, including which elev should take it

}
