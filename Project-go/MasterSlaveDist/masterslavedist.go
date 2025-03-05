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
	masterID		 = 0
)

func InitializeMasterSlaveDist(localElev elevio.Elevator, msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {

	localElevID = localElev.ElevatorID
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
				setMaster <- true
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

func AliveRecieved(elevID int, master bool, localElev elevio.Elevator, setMaster chan bool) {
	mu.Lock()
	defer mu.Unlock()

	ActiveElev[elevID] = true

	// Reset the watchdog timer
	startWatchdogTimer(elevID)

	resolveMasterConflict(master, localElev, elevID, setMaster)

}

func resolveMasterConflict(master bool, localElev elevio.Elevator, elevID int, setMaster chan bool) {
	// If we recieve a message from a master,
	// and we are a master that has previously been disconnected, we are now slave
	if localElev.Master && master {
		if Disconnected {
			setMaster <- false
			Disconnected = false
			masterID = elevID
		}
	} else if master {
		masterID = elevID
	}
}

func startWatchdogTimer(elevID int) {
	watchdogTimers[elevID] = time.NewTimer(time.Duration(watchdogDuration) * time.Second)
}

// If we have not recieved a message from an elevator within the watchdog duration, we assume it is disconnected
func WatchdogTimer(setMaster chan bool) {
	for {
		for i := 0; i < len(watchdogTimers); i++ {
			if watchdogTimers[i] != nil {
				select {
				case <-watchdogTimers[i].C:
					ActiveElev[i] = false

					ChangeMaster(setMaster, i)

				}
			}
		}

	}
}

func ChangeMaster(setMaster chan bool, i int) {

	numActiveElev := getNumActiveElev()

	// If we percieve ourselves as the only active elevator, we are disconnected from the rest of the system
	if numActiveElev == 1 {
		Disconnected = true
		setMaster <- true
		return
	}

	// If the disconnected elevator was the master, we need to elect a new master
	if i == masterID{
		for i := 0; i < localElevID; i++ {
			if ActiveElev[i] {
				return
			}
			setMaster <- true
		}
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
