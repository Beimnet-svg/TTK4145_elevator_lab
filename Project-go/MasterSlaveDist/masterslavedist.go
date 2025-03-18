package masterslavedist

import (
	config "Project-go/Config"
	elevio "Project-go/SingleElev/Elevio"
	"fmt"
	"time"
)

var (
	watchdogTimers   [config.NumberElev]*time.Timer
	checkMasterTimer *time.Timer

	ActiveElev [config.NumberElev]bool
	AliveElev  [config.NumberElev]bool

	Disconnected = false
	MasterID     = -1
)

func InitializeMasterSlaveDist(localElev elevio.Elevator, activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {

	ActiveElev[config.ElevID] = true
	AliveElev[config.ElevID] = true

	for ID := 0; ID < len(watchdogTimers); ID++ {
		// If different ID than our own we look at time between alive messages
		if ID != config.ElevID {
			startWatchdogTimer(ID, config.WatchdogDuration)
		}
		// If same ID we look at time since started operation
		if ID == config.ElevID {
			startWatchdogTimer(ID, config.InactiveDuration)
		}
	}

	timer := time.NewTimer(config.WatchdogDuration * time.Second)
	select {
	case <-activeOrdersArrived:
		// Connected to running system
		return
	case <-timer.C:
		// No master on the network
		if config.ElevID == 0 {
			setMaster <- true
			setMaster <- true
			MasterID = config.ElevID
			return
		}

		for j := 0; j < config.ElevID; j++ {
			if AliveElev[j] {
				return
			}
		}

		setMaster <- true
		setMaster <- true
		MasterID = config.ElevID

	}
}

func SetDisconnected(setDisconnected chan bool) {
	for range setDisconnected {
		Disconnected = true
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

	if Disconnected && checkMasterTimer == nil {
		fmt.Println("Starting checkMasterTimer")
		checkMasterTimer = time.NewTimer(config.WatchdogDuration * time.Second)
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

	if Disconnected {
		// If we had previously considered ourselves isolated, now we acknowledge a valid master.
		setMaster <- false
		setMaster <- false
		Disconnected = false
		checkMasterTimer = nil
		fmt.Println("Received heartbeat from elevator", senderElevID, "— clearing disconnected flag.")
		MasterID = senderElevID
	}

}

func CheckMasterTimerTimeout() {

	for {
		if checkMasterTimer == nil {
			continue
		}
		select {
		case <-checkMasterTimer.C:
			Disconnected = false
			checkMasterTimer = nil
		}
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
			startWatchdogTimer(config.ElevID, config.InactiveDuration)
			ActiveElev[config.ElevID] = true
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
					if i != config.ElevID {
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

	// If only this elevator is active, it should consider itself disconnected and take over.
	if numActiveElev == 1 {
		setMaster <- true
		setMaster <- true
		MasterID = config.ElevID
		return
	}

	// If the disconnected elevator was the master, check if any lower-priority elevator is still active.
	if disconnectedElevID == MasterID {
		if config.ElevID == 0 {
			setMaster <- true
			setMaster <- true
			MasterID = config.ElevID
			return
		}

		for j := 0; j < config.ElevID; j++ {
			if AliveElev[j] {
				MasterID = -1
				return
			}
		}
		// No lower active elevator found; signal master election.
		setMaster <- true
		setMaster <- true
		MasterID = config.ElevID
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
