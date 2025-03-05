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
	localElevID      int
	Disconnected     = false
)

func InitializeMasterSlaveDist(localElev *elevio.Elevator, msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool) {
	localElevID = localElev.ElevatorID

	// Set the local elevator as active
	ActiveElev[localElevID] = true

	// Start the watchdog timers for all elevators, apart from the local one
	for i := 0; i < len(watchdogTimers); i++ {
		if i != localElev.ElevatorID {
			startWatchdogTimer(i)
		}
	}

	// On startup we set ID 0 elevator as master, unless we already have a master on the internet sending messages
	if localElevID == 0 {
		timer := time.NewTimer(config.WatchdogDuration * time.Second)

		for {
			select {
			case <-msgArrived:
				return

			case <-timer.C:
				localElev.Master = true
				return

			}

		}
	}

}

func FetchAliveElevators(ElevState [config.NumberElev]elevio.Elevator) []elevio.Elevator {
	AliveElevatorStates := []elevio.Elevator{}
	for i := 0; i < len(ActiveElev); i++ {
		if ActiveElev[i] {
			AliveElevatorStates = append(AliveElevatorStates, ElevState[i])
		}
	}
	return AliveElevatorStates

}

func AliveRecieved(elevID int, master bool, localElev *elevio.Elevator) {
	mu.Lock()
	defer mu.Unlock()

	// Set the elevator as active, need it if we have set it as inactive before
	ActiveElev[elevID] = true

	// Reset the watchdog timer
	startWatchdogTimer(elevID)

	resolveMasterConflict(master, localElev)

}

func resolveMasterConflict(master bool, localElev *elevio.Elevator) {
	// If we recieve a message from a master,
	// and we are a master with lower ID or have been disconnected, we are now slave
	if localElev.Master && master {
		if Disconnected {
			localElev.Master = false
			Disconnected = false
		}
	}
}

// Watchdog timer working along with the alive message and timer module to
// check if elevators are alive
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
					ChangeMaster(setMaster)

				}
			}
		}

	}
}

func ChangeMaster(setMaster chan bool) {
	// Count the number of active elevators
	numActiveElev := getNumActiveElev()

	// If we percieve ourselves as the only active elevator, we are disconnected
	// from the rest of the system
	if numActiveElev == 1 {
		Disconnected = true
		setMaster <- true
		return
	}

	for i := 0; i < localElevID; i++ {
		if ActiveElev[i] {
			return
		}
		setMaster <- true
	}

}

func getNumActiveElev() int {
	numActiveElev := 0
	for i := 0; i < len(ActiveElev); i++ {
		if ActiveElev[i] {
			numActiveElev++
		}
	}
	return numActiveElev
}
