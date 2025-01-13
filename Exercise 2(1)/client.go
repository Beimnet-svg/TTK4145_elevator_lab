package main

import (
	"net"
)

func main() {
	serverAddr := "localhost:30000"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()

	conn.Write([]byte("This is from group 7!"))

}
