package masterslavedist

import (
	config "Project-go/Config"
	"Project-go/driver-go/elevio"
	"fmt"
	"time"
)

var (
	watchdogTimers [config.NumberElev]*time.Timer
	ActiveElev     [config.NumberElev]bool
	AliveElev      [config.NumberElev]bool
	localElevID    int
	Disconnected   = false
	MasterID       = -1
)

func InitializeMasterSlaveDist(localElev elevio.Elevator, activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {
	localElevID = localElev.ElevatorID
	ActiveElev[localElevID] = true
	AliveElev[localElevID] = true

	// Start the watchdog timers for all elevators except the local one.
	for i := 0; i < len(watchdogTimers); i++ {
		if i != localElevID {
			startWatchdogTimer(i, config.WatchdogDuration)
		}
		if i == localElevID {
			startWatchdogTimer(i, config.InactiveDuration)
		}
	}

	// All elevators start a timer to listen for an active master message.
	timer := time.NewTimer(config.WatchdogDuration * time.Second)
	select {
	case <-activeOrdersArrived:
		// A message arrived from another elevator, process it in AliveRecieved.
		fmt.Println("Message recieved")
		return
	case <-timer.C:
		// Timer expired with no message received; if this elevator is the highest priority, elect itself as master.
		if localElevID == 0 {
			setMaster <- true
			setMaster <- true
			MasterID = localElevID
			fmt.Printf("MasterID %d is now the master", MasterID)
			return
		}

		highestPriority := true
		for j := 0; j < localElevID; j++ {
			if AliveElev[j] {
				highestPriority = false
				break
			}
		}
		if highestPriority {
			setMaster <- true
			setMaster <- true
			MasterID = localElevID
		}
	}
}

func FetchAliveElevators(ElevState [config.NumberElev]elevio.Elevator) []elevio.Elevator {
	ActiveElevatorStates := []elevio.Elevator{}
	for i := 0; i < len(ActiveElev); i++ {
		if ActiveElev[i] {
			ActiveElevatorStates = append(ActiveElevatorStates, ElevState[i])
		}
	}
	return ActiveElevatorStates

}

func AliveRecievedFromSlave(elevID int, recievedE elevio.Elevator, setMaster chan bool) {

	if MasterID == -1 {
		MasterID = elevID
	}

	if recievedE.Inactive {
		ActiveElev[elevID] = false
	} else {
		ActiveElev[elevID] = true
	}

	AliveElev[elevID] = true
	startWatchdogTimer(elevID, config.WatchdogDuration)

}

func AliveRecievedFromMaster(elevID int, Inactive bool, localElev elevio.Elevator, setMaster chan bool) {
	if MasterID == -1 {
		MasterID = elevID
	}

	if Inactive {
		ActiveElev[elevID] = false

	} else {
		ActiveElev[elevID] = true

	}

	AliveElev[elevID] = true
	startWatchdogTimer(elevID, config.WatchdogDuration)

	if localElev.Master {
		resolveMasterConflict(elevID, setMaster)
	}

}

func resolveMasterConflict(senderElevID int, setMaster chan bool) {
	// If the received message indicates a master and the sender has a higher priority (lower ID)

	if Disconnected {
		// If we had previously considered ourselves isolated, now we acknowledge a valid master.
		setMaster <- false
		setMaster <- false
		Disconnected = false
		fmt.Println("Received heartbeat from elevator", senderElevID, "— clearing disconnected flag.")
		MasterID = senderElevID
	}

}

func startWatchdogTimer(elevID int, durationTime int) {
	duration := time.Duration(durationTime) * time.Second
	if watchdogTimers[elevID] != nil {
		// Reset the timer; if it wasn't active, drain its channel.
		if !watchdogTimers[elevID].Reset(duration) {
			// Try to drain the channel if necessary.
			select {
			case <-watchdogTimers[elevID].C:
			default:
			}
		}
	} else {
		watchdogTimers[elevID] = time.NewTimer(duration)
	}
}

func ResetInactiveTimer(resetInactiveElev chan int, elevInactive chan bool) {
	for {
		select {
		case <-resetInactiveElev:
			fmt.Println("Resetting inactive timer")
			startWatchdogTimer(localElevID, config.InactiveDuration)
			ActiveElev[localElevID] = true
			elevInactive <- false
			elevInactive <- false
		}
	}
}

// If we have not recieved a message from an elevator within the watchdog duration, we assume it is disconnected
func WatchdogTimer(setMaster chan bool, elevDied chan int, elevInactive chan bool) {
	for {

		for i := 0; i < len(watchdogTimers); i++ {
			if watchdogTimers[i] != nil {
				select {
				case <-watchdogTimers[i].C:
					if i != localElevID {
						ActiveElev[i] = false
						AliveElev[i] = false
						elevDied <- i
						fmt.Print("Elevator disc", i, "\n")
						ChangeMaster(setMaster, i)
					} else {
						fmt.Printf("Elevator %d inactive \n", i)
						ActiveElev[i] = false
						elevInactive <- true
						elevInactive <- true
					}

				default:
					// Timer hasn't fired; continue to the next timer.
				}
			}
		}
	}
}

func ChangeMaster(setMaster chan bool, disconnectedElevID int) {
	numActiveElev := getNumActiveElev()

	fmt.Println("Something is dead")

	// If only this elevator is active, it should consider itself disconnected and take over.
	if numActiveElev == 1 {
		Disconnected = true
		setMaster <- true
		setMaster <- true
		fmt.Println("Setting master")
		return
	}

	// If the disconnected elevator was the master, check if any lower-priority elevator is still active.
	if disconnectedElevID == MasterID {
		if localElevID == 0 {
			setMaster <- true
			setMaster <- true
			MasterID = localElevID
			return
		}

		for j := 0; j < localElevID; j++ {
			if AliveElev[j] {
				MasterID = -1
				return
			}
		}
		// No lower active elevator found; signal master election.
		setMaster <- true
		setMaster <- true
		MasterID = localElevID
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
