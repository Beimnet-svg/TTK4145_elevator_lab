package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {

	conn, err := net.Dial("tcp", "10.100.23.204:34933")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	//buffer := make([]byte, 1024)
	message := "Connect to: 10.100.23.17:34933\n\000"
	message2 := "HELLO FROM SERVER\n\000"
	//buffer = []byte(message)
	conn.Write([]byte(message))
	for {
		//buffer = []byte(message2)
		time.Sleep(2 * time.Second)
		conn.Write([]byte(message2))
		buffer := make([]byte, 1024)
		n, _ := conn.Read(buffer)
		fmt.Println("Server Response: ", string(buffer[:n]))
	}

}
