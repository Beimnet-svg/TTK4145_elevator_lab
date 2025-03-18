package networking

import (
	config "Project-go/Config"
	masterslavedist "Project-go/MasterSlaveDist"
	ordermanager "Project-go/OrderManager"
	elevfsm "Project-go/SingleElev/ElevFsm"
	elevio "Project-go/SingleElev/Elevio"
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
		localElev := elevfsm.GetElevator()

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

		localElev := elevfsm.GetElevator()
		//fmt.Print("Active elevators:", masterslavedist.ActiveElev, "\n")
		fmt.Print("Master:", localElev.Master, "\n")
		//fmt.Print("MasterID: ", masterslavedist.MasterID, "\n")
		//fmt.Print("Disconnected: ", masterslavedist.Disconnected, "\n")
		fmt.Print(("Ordercounter: "), ordermanager.GetOrderCounter(), "\n")
		fmt.Print("All active orders: ", ordermanager.GetAllActiveOrder(), "\n")
	}
}

func flushRecieverChannel(conn *net.UDPConn, buffer []byte) (*net.UDPConn, []byte) {

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

	return conn, buffer
}

func Receiver(activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {

	localAdress, _ := net.ResolveUDPAddr("udp", ":20007")
	conn, _ := net.ListenUDP("udp", localAdress)

	defer conn.Close()

	buffer := make([]byte, 1024)

	conn, buffer = flushRecieverChannel(conn, buffer)

	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			continue
		}

		msg, err := decodeMessage(buffer[:n])
		if err != nil {
			continue
		}

		localElev := elevfsm.GetElevator()

		if msg.Slave != nil && msg.Slave.ElevID != localElev.ElevatorID && localElev.Master {
			ordermanager.UpdateOrders(msg.Slave.E, activeOrdersArrived)
			masterslavedist.AliveRecievedFromSlave(msg.Slave.ElevID, msg.Slave.E, setMaster)

		} else if msg.Slave != nil && msg.Slave.ElevID != localElev.ElevatorID {
			masterslavedist.AliveRecievedFromSlave(msg.Slave.ElevID, msg.Slave.E, setMaster)

		} else if msg.Master != nil && msg.Master.ElevID != localElev.ElevatorID {
			masterslavedist.AliveRecievedFromMaster(msg.Master.ElevID, msg.Master.Inactive, localElev, setMaster)

			masterID := masterslavedist.GetMasterID()
			if masterID == msg.Master.ElevID || masterID == -1 {
				ordermanager.UpdateOrderCounter(msg.Master.OrderCounter)
				activeOrdersArrived <- msg.Master.Orders
			}
		}
	}

}
func SenderSlave(e elevio.Elevator, setDisconnected chan bool) {

	message := OrderMessage{
		Slave: &OrderMessageSlave{
			ElevID: e.ElevatorID,
			Master: false,
			E:      e,
		},
	}

	broadcastAddr := "255.255.255.255"
	destinationAddr, _ := net.ResolveUDPAddr("udp", broadcastAddr+":20007")

	conn, err := net.DialUDP("udp", nil, destinationAddr)
	if err != nil {
		setDisconnected <- true
		setDisconnected <- true
		fmt.Println("Error dialing UDP:", err)
		return
	}
	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	enc.Encode(message)

	content := buffer.Bytes()
	conn.Write(content)
}

func SenderMaster(e elevio.Elevator, orders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setDisconnected chan bool) {

	message := OrderMessage{
		Master: &OrderMessageMaster{
			ElevID:       e.ElevatorID,
			Master:       true,
			Orders:       orders,
			OrderCounter: ordermanager.GetOrderCounter(),
			Inactive:     e.Inactive,
		},
	}

	broadcastAddr := "255.255.255.255"
	destinationAddr, _ := net.ResolveUDPAddr("udp", broadcastAddr+":20007")

	conn, err := net.DialUDP("udp", nil, destinationAddr)
	if err != nil {
		setDisconnected <- true
		setDisconnected <- true
		fmt.Println("Error dialing UDP:", err)
		return
	}
	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	enc.Encode(message)

	content := buffer.Bytes()
	conn.Write(content)
}
