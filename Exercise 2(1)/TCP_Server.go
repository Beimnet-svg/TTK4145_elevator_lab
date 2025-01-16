package main

import (
	//"bufio"
	"fmt"
	"log"
	"net"
)

func main() {

	listner, err := net.Listen("tcp", ":34933")

	if err != nil {
		log.Fatal(err)
	}
	defer listner.Close()

	for {
		fmt.Printf("Entered loop\n")
		conn, err := listner.Accept()
		if err != nil {
			fmt.Print("This is an error", err)
			continue
		}
		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	// Read from the connection untill a new line is send
	defer conn.Close()
	for {
		buffer := make([]byte, 1024)

		n, _ := conn.Read(buffer)
		fmt.Printf("Message Received: %s", buffer[:n], "\n")

		_, err := conn.Write([]byte("Message Recieved\n"))

		if err != nil {
			fmt.Print("This is an error", err)
			break
		}

		// data, err := bufio.NewReader(conn).ReadString('\000')
		// if err != nil {
		// 	fmt.Println(err)
		// }

		// fmt.Print("> ", string(data))

	}
}
