package main

import (
	"fmt"
	"net"
)

func main() {
	conn, _ := net.ListenPacket("udp", ":20007")
	defer conn.Close()

	buffer := make([]byte, 1024)
	fmt.Println("Listening for UDP on port 20007........")
	for {
		n, addr, _ := conn.ReadFrom(buffer)

		fmt.Printf("Received %s from %s\n", string(buffer[0:n]), addr)
	}
}
