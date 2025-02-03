package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

var (
	melding       = make(chan []byte)
	currentNumber int
)

func networking(receiver chan<- []byte, ctx context.Context) {
	conn, _ := net.ListenPacket("udp", ":20007")

	defer conn.Close()

	fmt.Println("Listening for UDP on port 20007")
	buffer := make([]byte, 1024)
	for {
		// Set a short read deadline so that ReadFrom doesn't block forever.
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			// If the error is a timeout, check if the context is done.
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				select {
				case <-ctx.Done():
					// Context cancelled; exit the goroutine.
					return
				default:
					// Timeout occurred, but the context isn't cancelled yet.
					continue
				}
			}
			continue
		}

		// Send the received data on the channel.
		receiver <- buffer[:n]
		// Optionally, reset the buffer if needed.
		buffer = make([]byte, 1024)
	}
}

func primaryFunc(number int) {
	exec.Command("gnome-terminal", "--", "go", "run", "ex4A.go").Run()

	serverAddr := "localhost:20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()
	i := number + 1
	time.Sleep(2 * time.Second)

	for {
		msg := strconv.Itoa(i)
		conn.Write([]byte(msg))
		fmt.Println("Sent:", msg)
		time.Sleep(1 * time.Second)
		i++

	}

}

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	go networking(melding, ctx)

	start := time.Now()

	for {
		select {
		case msg := <-melding:
			start = time.Now()
			fmt.Println("Received:", string(msg))
			currentNumber, _ = strconv.Atoi(string(msg))

		default:
			elapsed := time.Since(start)
			if elapsed.Seconds() > 3 {
				//Program failed
				//close goroutine
				cancel()
				primaryFunc(currentNumber)

			}

		}
	}
}
