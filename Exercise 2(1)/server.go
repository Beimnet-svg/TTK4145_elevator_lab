package main

import (
	"fmt"
	"net"
)

func main() {
	localAdress, _ := net.ResolveUDPAddr("udp", ":20007")
	connection, err := net.ListenUDP("udp", localAdress)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer connection.Close()

	buffer := make([]byte, 1024)
	fmt.Println("Listening for UDP on port 20007........")
	for {
		n, addr, _ := connection.ReadFrom(buffer)

		fmt.Printf("Received %s from %s\n", string(buffer[0:n]), addr)
	}
}
