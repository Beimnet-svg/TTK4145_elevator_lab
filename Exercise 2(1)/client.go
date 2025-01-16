package main

import (
	"net"
	"time"
)

func main() {
	serverAddr := "10.100.23.204:20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	for {
		conn.Write([]byte("This is from group 7!"))
		time.Sleep(1 * time.Second)
	}
}
