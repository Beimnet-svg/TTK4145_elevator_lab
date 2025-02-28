package masterslavedist

import (
	"Project-go/driver-go/elevio"
	"fmt"
	"sync"
	"time"
)

var (
	watchdogTimer    *time.Timer
	watchdogDuration = 5 * time.Second
	mu               sync.Mutex
)

func init() {
	resetWatchdogTimer()
	go Timer()
}

func FetchElevators() []elevio.Elevator {
	return []elevio.Elevator{}
}

// Implemented in the network module after recieving an alive message
func AliveRecieved(elevID int) bool {
	mu.Lock()
	defer mu.Unlock()
	e := FetchElevators()
	for i := 0; i < len(e); i++ {
		if e[i].ElevatorID == elevID {
			// If there are two masters, resolve conflict
			if e[i].Master && detectMultipleMasters(e[i]) {
				resolveMasterConflict(&e[i])
			}

			if e[i].Master {
				resetWatchdogTimer()
			}
			saveElevatorState(e[i])
			return true
		}
	}
	return false
}

func resolveMasterConflict(elevator *elevio.Elevator) {
	activeElevators := FetchElevators()
	var highestPriorityMaster *elevio.Elevator

	for i := range activeElevators {
		if activeElevators[i].Master {
			// Select the lowest ElevatorID as the master
			if highestPriorityMaster == nil || activeElevators[i].ElevatorID < highestPriorityMaster.ElevatorID {
				highestPriorityMaster = &activeElevators[i]
			}
		}
	}

	// If the current elevator is not the chosen master, step down
	if highestPriorityMaster != nil && elevator.ElevatorID != highestPriorityMaster.ElevatorID {
		elevator.Master = false
	}
}

func detectMultipleMasters(elevator elevio.Elevator) bool {
	activeElevators := FetchElevators()
	masterCount := 0

	for _, e := range activeElevators {
		if e.Master {
			masterCount++
		}
	}

	return masterCount > 1
}

func resetWatchdogTimer() {
	if watchdogTimer != nil {
		// Stop the timer because the master is alive
		watchdogTimer.Stop()
	}
	//Then master is considered dead after 5 seconds
	watchdogTimer = time.AfterFunc(watchdogDuration, func() {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("Master is considered dead")
		ChangeMaster()
	})
}

func saveElevatorState(e elevio.Elevator) {
	fmt.Printf("Saving the state of elevator %d", e.ElevatorID)
}

// Timer module to check if something dies
func Timer() {
	//Start timer
	for {
		time.Sleep(1 * time.Second)
		mu.Lock()
		if watchdogTimer != nil {
			fmt.Println("Watchdog timer is active")
		} else {
			fmt.Println("Watchdog timer is not active")
			ChangeMaster()
		}
		mu.Unlock()
	}
}

func attemptRestart(elevator *elevio.Elevator) {
	if elevator == nil {
		return
	}
	fmt.Printf("Attempting to restart elevator %d...\n", elevator.ElevatorID)

	// Simulated restart process
	go func() {
		time.Sleep(5 * time.Second) // Simulating restart delay
		mu.Lock()
		defer mu.Unlock()
		fmt.Printf("Elevator %d restarted.\n", elevator.ElevatorID)
	}()
}

// Change master function
func ChangeMaster() {
	fmt.Println("Changing master")

	mu.Lock()
	defer mu.Unlock()

	elevators := FetchElevators()
	var deadElevator *elevio.Elevator

	// Identify the dead elevator based on a missing heartbeat
	for i := range elevators {
		if !AliveRecieved(elevators[i].ElevatorID) { // Check if still alive
			deadElevator = &elevators[i]
			break
		}
	}

	if deadElevator != nil {
		fmt.Printf("Elevator %d (dead) was the master. Redistributing orders...\n", deadElevator.ElevatorID)
		//Thought of redistributing dead elevator's orders here
		// redistributeOrders(deadElevator)
		attemptRestart(deadElevator) // Try restarting dead elevators
	}

	// Elect a new master
	for i := range elevators {
		if !elevators[i].Master { // Find an elevator that was NOT master
			elevators[i].Master = true
			fmt.Printf("Elevator %d is now the new master\n", elevators[i].ElevatorID)
			break
		}
	}
}
