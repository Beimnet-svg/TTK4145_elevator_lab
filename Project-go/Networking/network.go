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
	E      elevio.Elevator
}

type OrderMessageMaster struct {
	ElevID       int
	Master       bool
	Orders       [config.NumberElev][config.NumberFloors][config.NumberBtn]bool
	OrderCounter [config.NumberElev]int
	Inactive     bool
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

func Sender(activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setDisconnected chan bool) {
	ticker := time.NewTicker(config.SendDelay * time.Millisecond)
	for range ticker.C {
		localElev := elevator_fsm.GetElevator()

		if localElev.Master {
			orders := ordermanager.GetAllActiveOrder()
			SenderMaster(localElev, orders, setDisconnected)
			ordermanager.UpdateOrders(localElev, activeOrdersArrived)

		} else {
			SenderSlave(localElev, setDisconnected)
		}

	}
}

func Print() {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {

		localElev := elevator_fsm.GetElevator()
		fmt.Print("Active elevators:", masterslavedist.ActiveElev, "\n")
		fmt.Print("Master:", localElev.Master, "\n")
		fmt.Print("MasterID: ", masterslavedist.MasterID, "\n")
		fmt.Print("Disconnected: ", masterslavedist.Disconnected, "\n")
		fmt.Print(("Ordercounter: "), ordermanager.GetOrderCounter(), "\n")
		fmt.Print("All active orders: ", ordermanager.GetAllActiveOrder(), "\n")
	}
}

func Receiver(activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {
	// Listen for incoming UDP packets on port 20007
	localAdress, _ := net.ResolveUDPAddr("udp", ":20007")
	conn, err := net.ListenUDP("udp", localAdress)
	if err != nil {
		log.Fatal("Error listening on port 20007:", err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	// Flush any pending messages in the buffer
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	for {
		_, _, err := conn.ReadFrom(buffer)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				// No more messages to read, exit the flush loop.
				break
			}
			log.Println("Error while flushing UDP buffer:", err)
		}
	}
	// Remove the read deadline to resume normal operation
	conn.SetReadDeadline(time.Time{})

	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Println("Error reading from connection:", err)
			continue // Log the error and keep listening
		}

		msg, err := decodeMessage(buffer[:n])
		if err != nil {
			log.Println("Error decoding message:", err)
			continue // Skip this malformed message
		}

		localElev := elevator_fsm.GetElevator()

		//If we got msg from same elevator id as we have locally, skip it
		if msg.Slave != nil && msg.Slave.ElevID != localElev.ElevatorID && localElev.Master {
			//fmt.Println("Recieved request from", msg.Slave.ElevID, "with request", msg.Slave.E.Requests)
			ordermanager.UpdateOrders(msg.Slave.E, activeOrdersArrived)
			masterslavedist.AliveRecievedFromSlave(msg.Slave.ElevID, msg.Slave.E, setMaster)
		} else if msg.Slave != nil && msg.Slave.ElevID != localElev.ElevatorID {
			masterslavedist.AliveRecievedFromSlave(msg.Slave.ElevID, msg.Slave.E, setMaster)
		} else if msg.Master != nil && msg.Master.ElevID != localElev.ElevatorID {
			masterslavedist.AliveRecievedFromMaster(msg.Master.ElevID, msg.Master.Inactive, localElev, setMaster)
			if masterslavedist.MasterID == msg.Master.ElevID || masterslavedist.MasterID == -1 {
				ordermanager.UpdateOrderCounter(msg.Master.OrderCounter)
				activeOrdersArrived <- msg.Master.Orders
			}
		}
	}

}
func SenderSlave(E elevio.Elevator, setDisconnected chan bool) {

	message := OrderMessage{
		Slave: &OrderMessageSlave{
			ElevID: E.ElevatorID,
			Master: false,
			E:      E,
		},
	}

	broadcastAddr := "255.255.255.255"
	destinationAddr, err := net.ResolveUDPAddr("udp", broadcastAddr+":20007")

	var conn *net.UDPConn
	conn, err = net.DialUDP("udp", nil, destinationAddr)
	if err != nil {
		setDisconnected <- true
		setDisconnected <- true
		fmt.Println("Error dialing UDP:", err)
		return
	}

	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(message)
	if err != nil {
		fmt.Println("Error encoding orders:", err)
	}
	content := buffer.Bytes()
	conn.Write(content)

}

func SenderMaster(E elevio.Elevator, orders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setDisconnected chan bool) {

	message := OrderMessage{
		Master: &OrderMessageMaster{
			ElevID:       E.ElevatorID,
			Master:       true,
			Orders:       orders,
			OrderCounter: ordermanager.GetOrderCounter(),
			Inactive:     E.Inactive,
		},
	}

	broadcastAddr := "255.255.255.255"
	destinationAddr, err := net.ResolveUDPAddr("udp", broadcastAddr+":20007")

	var conn *net.UDPConn
	conn, err = net.DialUDP("udp", nil, destinationAddr)
	if err != nil {
		setDisconnected <- true
		setDisconnected <- true
		fmt.Println("Error dialing UDP:", err)
		return
	}

	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(message)
	if err != nil {
		fmt.Println("Error encoding orders:", err)
	}
	content := buffer.Bytes()
	conn.Write(content)

}
