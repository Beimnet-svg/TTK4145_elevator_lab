package main

import (
	"fmt"
	"net"
	"time"
)

func main() {

	conn, _ := net.ListenPacket("udp", ":20007")
	defer conn.Close()

	buffer := make([]byte, 1024)
	fmt.Println("Listening for UDP on port 20007 \n")
	start:= time.Now()


	for {
		

		select {
		case msg:=<-chanel:
			start=time.Now()
			
			n, _, _ := conn.ReadFrom(buffer)
			//data:=binary.BigEndian.Uint64(buffer[0:n])
			fmt.Println(buffer[0:n])

		default:
			elapsed:= time.Since(start)
			if elapsed>3 {
				//Program failed
				//Start another program (backup)
			}
	
	}
}
