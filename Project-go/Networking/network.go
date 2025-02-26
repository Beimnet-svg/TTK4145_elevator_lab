package networking

import (
	ordermanager "Project-go/OrderManager"
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

// Define a heartbeat message structure if needed
type HeartbeatMessage struct {
    ElevID int
    Master bool
}

func UnifiedReceiver(orderChan chan [3][4][3]bool, heartbeatChan chan HeartbeatMessage) {
    conn, err := net.ListenPacket("udp", ":20007")
    if err != nil {
        log.Fatal("Error listening on port 20007:", err)
    }
    defer conn.Close()

    buffer := make([]byte, 1024)

    for {
        n, _, err := conn.ReadFrom(buffer)
        if err != nil {
            log.Println("Error reading from connection:", err)
            continue
        }

        msg, err := decodeMessage(buffer[:n])
        if err != nil {
            log.Println("Error decoding message:", err)
            continue
        }

        // Dispatch based on message type
        if msg.Slave != nil {
            // Process orders and heartbeat for a slave message
            ordermanager.UpdateOrders(msg.Slave.e, orderChan)
            heartbeatChan <- HeartbeatMessage{
                ElevID: msg.Slave.ElevID,
                Master: msg.Slave.Master,
                // fill in any additional fields if necessary
            }
        } else if msg.Master != nil {
            // Process master message for orders and heartbeat
            heartbeatChan <- HeartbeatMessage{
                ElevID: msg.Master.ElevID,
                Master: msg.Master.Master,
            }
            orderChan <- msg.Master.Orders
        } else {
            log.Println("Received message with unknown type")
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
			e: 		e,
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


func AliveM(e elevio.Elevator) {
	//Call this when we want to send a message

	// Create an instance of the struct
	message := OrderMessage{
		Slave: &OrderMessageSlave{
			ElevID: e.ElevatorID,
			Master: false,
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

func HeartBeatSender(e *elevio.Elevator){
	ticker := time.NewTicker(100*time.Millisecond)
    defer ticker.Stop()
    for {
		select {
		case <-ticker.C:
        // Send heartbeat based on current state
        AliveM(*e)
			
		}
        
    }
}

func StructM(e elevio.Elevator) {
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

