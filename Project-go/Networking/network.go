package networking

import (
	"Project-go/driver-go/elevio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

var mutex sync.Mutex
var Elevators []elevio.Elevator


type HeartbeatMessage struct {
	ElevID int
	Master bool
	SeqNum int // Add sequence number to prevent duplicate processing
}

type StateUpdateMessage struct {
	ElevID   int
	Master   bool
	Elevator elevio.Elevator
	SeqNum   int
}

type MasterUpdateMessage struct {
	ElevID         int
	Master         bool
	AllActiveOrders [3][4][3]bool
	SeqNum         int
}

// Acknowledgment Message
type AckMessage struct {
	AckSeqNum int // Sequence number of the received message
}

// Generic message container
type OrderMessage struct {
	Heartbeat *HeartbeatMessage
	State     *StateUpdateMessage
	Master    *MasterUpdateMessage
	Ack       *AckMessage
}


func decodeMessage(buffer []byte) (*OrderMessage, error) {
	buf := bytes.NewBuffer(buffer)
	dec := gob.NewDecoder(buf)
	var message OrderMessage
	err := dec.Decode(&message)
	return &message, err
}


func init() {
	gob.Register(HeartbeatMessage{})
	gob.Register(StateUpdateMessage{})
	gob.Register(MasterUpdateMessage{})
	gob.Register(AckMessage{})
	gob.Register(OrderMessage{})
}


func UnifiedReceiver(orderChan chan [3][4][3]bool, heartbeatChan chan HeartbeatMessage, stateUpdateChan chan elevio.Elevator) {
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

		if msg.Heartbeat != nil {
			// Only process heartbeats quickly without blocking
			select {
			case heartbeatChan <- *msg.Heartbeat:
			default:
				// Drop if the channel is full (prevents blocking)
			}
		} else if msg.State != nil {
			// Process state update
			mutex.Lock()
			stateUpdateChan <- msg.State.Elevator
			mutex.Unlock()
		} else if msg.Master != nil {
			// Process master order update
			heartbeatChan <- HeartbeatMessage{
				ElevID: msg.Master.ElevID,
				Master: msg.Master.Master,
			}
			orderChan <- msg.Master.AllActiveOrders
		} else {
			log.Println("Received unknown message type")
		}
	}
}
func sendUDPMessage(message OrderMessage) {
	serverAddr := ":20007"
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to UDP:", err)
		return
	}
	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(message)
	if err != nil {
		fmt.Println("Error encoding message:", err)
		return
	}
	conn.Write(buffer.Bytes())
}


func sendAck(conn net.PacketConn, addr net.Addr, seqNum int) {
	ackMessage := OrderMessage{
		Ack: &AckMessage{AckSeqNum: seqNum},
	}
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(ackMessage)
	if err != nil {
		fmt.Println("Error encoding ACK:", err)
		return
	}
	conn.WriteTo(buffer.Bytes(), addr)
}


func sendReliableUDPMessage(message OrderMessage) {
	serverAddr := ":20007"
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to UDP:", err)
		return
	}
	defer conn.Close()

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(message)
	if err != nil {
		fmt.Println("Error encoding message:", err)
		return
	}

	seqNum := rand.Intn(100000) 
	retries := 0
	maxRetries := 1000
	ackChan := make(chan int, 1) 

	
	for retries < maxRetries {
	
		if message.Heartbeat != nil {
			message.Heartbeat.SeqNum = seqNum
		} else if message.State != nil {
			message.State.SeqNum = seqNum
		} else if message.Master != nil {
			message.Master.SeqNum = seqNum
		}

		
		buffer.Reset()
		enc = gob.NewEncoder(&buffer)
		err = enc.Encode(message)
		if err != nil {
			fmt.Println("Error encoding message:", err)
			return
		}

		conn.Write(buffer.Bytes()) 
		select {
		case ack := <-ackChan:
			if ack == seqNum {
				fmt.Println("ACK received for seqNum:", seqNum)
				return // Stop resending if ACK received
			}
		case <-time.After(10 * time.Millisecond): 
		}
		retries++
	}

	fmt.Println("Message delivery failed after", maxRetries, "attempts")
}


func SendHeartbeat(elevID int, isMaster bool) {
	message := OrderMessage{
		Heartbeat: &HeartbeatMessage{
			ElevID: elevID,
			Master: isMaster,
		},
	}
	sendUDPMessage(message)
}

func SendStateUpdate(e elevio.Elevator) {
	message := OrderMessage{
		State: &StateUpdateMessage{
			ElevID:   e.ElevatorID,
			Master:   false,
			Elevator: e,
		},
	}
	sendReliableUDPMessage(message)
}

func SendMasterUpdate(e elevio.Elevator, orders [3][4][3]bool) {
	message := OrderMessage{
		Master: &MasterUpdateMessage{
			ElevID:         e.ElevatorID,
			Master:         true,
			AllActiveOrders: orders,
		},
	}
	sendReliableUDPMessage(message)
}

func HeartbeatSenderSlave(e *elevio.Elevator) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	fmt.Println("🟢 Starting HeartbeatSenderSlave...")

	for {
		select {
		case <-ticker.C:
			fmt.Println("Sending Slave Heartbeat at:", time.Now().Format("15:04:05.000")) // Timestamp
			SendHeartbeat(e.ElevatorID, false)
		}
	}
}

func HeartbeatSenderMaster(e *elevio.Elevator) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	fmt.Println("🟢 Starting HeartbeatSenderMaster...")

	for {
		select {
		case <-ticker.C:
			fmt.Println("Sending Master Heartbeat at:", time.Now().Format("15:04:05.000")) // Timestamp
			SendHeartbeat(e.ElevatorID, true)
		}
	}
}

