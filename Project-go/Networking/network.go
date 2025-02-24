package networking

import (
	masterslavedist "Project-go/MasterSlaveDist"
	ordermanager "Project-go/OrderManager"
	"Project-go/driver-go/elevio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

type OrderMessageSlave struct {
	ElevID int
	Master bool
	e      elevio.Elevator
}

type OrderMessageMaster struct {
	ElevID int
	Master bool
	Orders [3][4][3]bool
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

func Receiver(receiver chan [3][4][3]bool) {
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

		// Process the received message
		if msg.Slave != nil {
			ordermanager.UpdateOrders(msg.Slave.e, receiver)
			masterslavedist.AliveRecieved(msg.Slave.ElevID, msg.Slave.Master)
		} else if msg.Master != nil {
			masterslavedist.AliveRecieved(msg.Master.ElevID, msg.Master.Master)
			receiver <- msg.Master.Orders
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

func SenderMaster(e elevio.Elevator, orders [3][4][3]bool) {
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

/*
func SenderAlive(e elevio.Elevator) {
	//Call this when we want to send a message

	serverAddr := ":20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	message := strconv.Itoa(e.ElevatorID) + strconv.Itoa(e.Master)
	content := []byte(message)
	conn.Write(content)

}
func SenderStruct(e elevio.Elevator) {
	//Call this when we want to send a message

	serverAddr := ":20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(e)
	if err != nil {
		fmt.Println("Error encoding elevator:", err)
	}
	content := buffer.Bytes()
	conn.Write(content)

}
func SenderOrdersMaster(orders [3][4][3]int) {
	//Call this when we want to send a message

	serverAddr := ":20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	//Master sending out orders to all elevators, including which elev should take it
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(orders)
	if err != nil {
		fmt.Println("Error encoding orders:", err)
	}
	content := buffer.Bytes()
	conn.Write(content)
	//Master sending out orders to all elevators, including which elev should take it

}

func SenderNewOrderSlave(elevID int, orders [4][3][2]int) {
	//Call this when we want to send a message
	type OrderMessage struct {
        ElevID int
        Orders [4][3][2]int
    }

    // Create an instance of the struct
    message := OrderMessage{
        ElevID: elevID,
        Orders: orders,
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
*/
