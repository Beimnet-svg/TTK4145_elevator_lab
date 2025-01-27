package main

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

func main() {
	exec.Command("gnome-terminal", "--", "go", "run", "ex4A.go").Run()

	serverAddr := "localhost:20007"
	conn, _ := net.Dial("udp", serverAddr)
	defer conn.Close()
	i := 0

	for {
		conn.Write([]byte(string(i)))
		fmt.Print(i)
		time.Sleep(1 * time.Second)
		i++
	}
}
