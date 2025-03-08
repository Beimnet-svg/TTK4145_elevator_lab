package masterslavedist

import (
	config "Project-go/Config"
	"Project-go/driver-go/elevio"
	"fmt"
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
	masterID         = 0
)

func InitializeMasterSlaveDist(localElev elevio.Elevator, msgArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {
	localElevID = localElev.ElevatorID
	ActiveElev[localElevID] = true

	// Start the watchdog timers for all elevators except the local one.
	for i := 0; i < len(watchdogTimers); i++ {
		if i != localElev.ElevatorID {
			startWatchdogTimer(i)
		}
	}

	// All elevators start a timer to listen for an active master message.
	timer := time.NewTimer(config.WatchdogDuration * time.Second)
	select {
	case <-msgArrived:
		// A message arrived from another elevator, process it in AliveRecieved.
		return
	case <-timer.C:
		// Timer expired with no message received; if this elevator is the highest priority, elect itself as master.
		highestPriority := true
		for j := 0; j < localElevID; j++ {
			if ActiveElev[j] {
				highestPriority = false
				break
			}
		}
		if highestPriority {
			setMaster <- true
			masterID = localElevID
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

	// Reset the watchdog timer for the sender
	startWatchdogTimer(elevID)

	// If we were previously marked as disconnected and we receive a message from another elevator,
	// reset the disconnected flag.
	if Disconnected && elevID != localElev.ElevatorID {
		fmt.Println("Reconnecting: clearing disconnected flag.")
		Disconnected = false
	}

	resolveMasterConflict(master, localElev, elevID, setMaster)
}

func resolveMasterConflict(isMsgMaster bool, localElev elevio.Elevator, senderElevID int, setMaster chan bool) {
	// If the received message indicates a master and the sender has a higher priority (lower ID)
	if isMsgMaster {
		if localElev.Master && senderElevID < localElevID {
			// If we're master but a higher-priority elevator is active, step down.
			if Disconnected {
				// If we had previously considered ourselves isolated, now we acknowledge a valid master.
				setMaster <- false
				Disconnected = false
			}
			masterID = senderElevID
		} else if !localElev.Master {
			// Simply update masterID if we are not master.
			masterID = senderElevID
		}
	}
}

func startWatchdogTimer(elevID int) {
	watchdogTimers[elevID] = time.NewTimer(time.Duration(watchdogDuration) * time.Second)
}

// If we have not recieved a message from an elevator within the watchdog duration, we assume it is disconnected
func WatchdogTimer(setMaster chan bool) {
	for {
		for i := 0; i < len(watchdogTimers); i++ {
			if watchdogTimers[i] != nil && i != localElevID {
				select {
				case <-watchdogTimers[i].C:
					ActiveElev[i] = false
					fmt.Print("Elevator disc\n")

					ChangeMaster(setMaster, i)

				}
			}
		}

	}
}

func ChangeMaster(setMaster chan bool, disconnectedElevID int) {
	numActiveElev := getNumActiveElev()

	// If only this elevator is active, it should consider itself disconnected and take over.
	if numActiveElev == 1 {
		Disconnected = true
		setMaster <- true
		return
	}

	// If the disconnected elevator was the master, check if any lower-priority elevator is still active.
	if disconnectedElevID == masterID {
		for j := 0; j < localElevID; j++ {
			if ActiveElev[j] {
				// A lower ID is still active; do not trigger an election.
				return
			}
		}
		// No lower active elevator found; signal master election exactly once.
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
