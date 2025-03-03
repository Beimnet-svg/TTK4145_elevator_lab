package masterslavedist

import (
	config "Project-go/Config"
	"Project-go/driver-go/elevio"
	"sync"
	"time"
)

var (
	watchdogTimers   [config.NumberElev]*time.Timer
	watchdogDuration = config.WatchdogDuration
	mu               sync.Mutex
	ActiveElev       [config.NumberElev]bool
	ActiveElevState  [config.NumberElev]elevio.Elevator
	localElevID      int
)

func InitializeMasterSlaveDist(localElev *elevio.Elevator) {
	localElevID = localElev.ElevatorID
	ActiveElev[localElevID] = true
	if localElevID == 0 {
		localElev.Master = true
	}

	for i := 0; i < len(watchdogTimers); i++ {
		if i != localElev.ElevatorID {
			startWatchdogTimer(i)
		}
	}

}

func FetchAliveElevators() []elevio.Elevator {
	AliveElevators := []elevio.Elevator{}
	for i := 0; i < len(ActiveElev); i++ {
		if ActiveElev[i] {
			AliveElevators = append(AliveElevators, ActiveElevState[i])
		}
	}
	return AliveElevators

}

// Implemented in the network module after recieving an alive message
func AliveRecieved(elevID int, master bool, localElev *elevio.Elevator) {
	mu.Lock()
	defer mu.Unlock()

	// Set the elevator as active, need it if we have set it as inactive before
	ActiveElev[elevID] = true

	// Reset the watchdog timer
	startWatchdogTimer(elevID)

	resolveMasterConflict(elevID, master, localElev)

}

func resolveMasterConflict(elevID int, master bool, localElev *elevio.Elevator) {
	// If we recieve a message from a master, and we are a master with lower ID, we are now slave
	if localElev.Master && master {
		if localElev.ElevatorID > elevID {
			localElev.Master = false
		}
	}
}

// Watchdog timer working along with the alive message and timer module to
// check if the master is still alive
func startWatchdogTimer(elevID int) {
	watchdogTimers[elevID] = time.NewTimer(time.Duration(watchdogDuration) * time.Second)
}

// Timer module to check if something dies
func WatchdogTimer(setMaster chan bool) {
	//Start timer
	for {
		for i := 0; i < len(watchdogTimers); i++ {
			if watchdogTimers[i] != nil {
				select {
				case <-watchdogTimers[i].C:
					ActiveElev[i] = false
					if i < localElevID {
						ChangeMaster(setMaster)
					}
				}
			}
		}

	}
}

// Change master function
func ChangeMaster(setMaster chan bool) {
	// If no elevators with lower ID than us are active, we can become master
	for i := 0; i < localElevID; i++ {
		if ActiveElev[i] {
			return
		}
		setMaster <- true
	}

}

