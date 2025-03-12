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

func Sender(msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		localElev := elevator_fsm.GetElevator()

		// If the elevator is percieved disconnected, we don't send any messages other than locally
		if masterslavedist.Disconnected {
			ordermanager.UpdateOrders(localElev, msgArrived)
			continue
		}

		if localElev.Master {
			orders := ordermanager.GetAllActiveOrder()
			SenderMaster(localElev, orders)
			ordermanager.UpdateOrders(localElev, msgArrived)

		} else {
			SenderSlave(localElev)
		}

	}
}

func Receiver(msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {
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

		// Ignore messages from localhost (see next section for improvements)
		// localAddr := conn.LocalAddr().(*net.UDPAddr)
		// if localAddr.IP.Equal(addrSender.(*net.UDPAddr).IP) {
		// 	continue
		// }

		msg, err := decodeMessage(buffer[:n])
		if err != nil {
			log.Println("Error decoding message:", err)
			continue // Skip this malformed message
		}

		localElev := elevator_fsm.GetElevator()

		//If we got msg from same elevator id as we have locally, skip it
		if msg.Slave != nil && msg.Slave.ElevID != localElev.ElevatorID {
			ordermanager.UpdateOrders(msg.Slave.E, msgArrived)
			masterslavedist.AliveRecieved(msg.Slave.ElevID, msg.Slave.Master, localElev, setMaster)
		} else if msg.Master != nil && msg.Master.ElevID != localElev.ElevatorID {
			masterslavedist.AliveRecieved(msg.Master.ElevID, msg.Master.Master, localElev, setMaster)
			msgArrived <- msg.Master.Orders
		}
	}

}
func SenderSlave(E elevio.Elevator) {

	message := OrderMessage{
		Slave: &OrderMessageSlave{
			ElevID: E.ElevatorID,
			Master: false,
			E:      E,
		},
	}
	fmt.Println(message.Slave.E.Requests)

	broadcastAddr := "255.255.255.255"
	destinationAddr, _ := net.ResolveUDPAddr("udp", broadcastAddr+":20007")
	conn, err := net.DialUDP("udp", nil, destinationAddr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
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

func SenderMaster(E elevio.Elevator, orders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {

	message := OrderMessage{
		Master: &OrderMessageMaster{
			ElevID: E.ElevatorID,
			Master: true,
			Orders: orders,
		},
	}

	broadcastAddr := "255.255.255.255"
	destinationAddr, _ := net.ResolveUDPAddr("udp", broadcastAddr+":20007")
	conn, err := net.DialUDP("udp", nil, destinationAddr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
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
