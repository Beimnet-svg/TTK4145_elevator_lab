package main

import (
	"net"
	"time"
)

func main() {
	broadcastAddr := "255.255.255.255"
	destinationAddr, _ := net.ResolveUDPAddr("udp", broadcastAddr+":20007")
	conn, err := net.DialUDP("udp", nil, destinationAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		conn.Write([]byte("This is from group 7!"))
		time.Sleep(1 * time.Second)
	}
}
