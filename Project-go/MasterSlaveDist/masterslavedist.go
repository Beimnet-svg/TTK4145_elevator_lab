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
	masterTimer      *time.Timer

	activeElev [config.NumberElev]bool
	aliveElev  [config.NumberElev]bool

	disconnected = false
	masterID     = -1 // -1 means master unknown
)

func InitializeMasterSlaveDist(localElev elevio.Elevator, activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {

	activeElev[config.ElevID] = true
	aliveElev[config.ElevID] = true

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
			masterID = config.ElevID
			return
		}

		for j := 0; j < config.ElevID; j++ {
			if aliveElev[j] {
				return
			}
		}

		setMaster <- true
		setMaster <- true
		masterID = config.ElevID

	}
}

func GetMasterID() int {
	return masterID
}

func GetDisconnected() bool {
	return disconnected
}

func GetActiveElev() [config.NumberElev]bool {
	return activeElev
}

func SetDisconnected(setDisconnected chan bool) {
	for range setDisconnected {
		disconnected = true
	}
}

func FetchActiveElevators(elevState [config.NumberElev]elevio.Elevator) []elevio.Elevator {
	activeElevatorStates := []elevio.Elevator{}
	for i := 0; i < len(activeElev); i++ {
		if activeElev[i] {
			activeElevatorStates = append(activeElevatorStates, elevState[i])
		}
	}
	return activeElevatorStates

}

func AliveRecievedFromSlave(senderElevID int, senderE elevio.Elevator, setMaster chan bool) {

	if disconnected && checkMasterTimer == nil {
		fmt.Println("Starting checkMasterTimer")
		checkMasterTimer = time.NewTimer(config.WatchdogDuration * time.Second)
	}

	if senderE.Inactive {
		activeElev[senderElevID] = false
	} else {
		activeElev[senderElevID] = true
	}

	aliveElev[senderElevID] = true
	startWatchdogTimer(senderElevID, config.WatchdogDuration)

}

func AliveRecievedFromMaster(senderElevID int, inactive bool, localElev elevio.Elevator, setMaster chan bool) {

	resetNoMasterTimer(4)

	if masterID == -1 {
		masterID = senderElevID
	}

	if inactive {
		activeElev[senderElevID] = false

	} else {
		activeElev[senderElevID] = true

	}

	aliveElev[senderElevID] = true
	startWatchdogTimer(senderElevID, config.WatchdogDuration)

	if localElev.Master {
		resolveMasterConflict(senderElevID, setMaster)
	}

}

func resolveMasterConflict(senderElevID int, setMaster chan bool) {

	if disconnected {
		setMaster <- false
		setMaster <- false
		disconnected = false
		checkMasterTimer = nil
		fmt.Println("Received heartbeat from elevator", senderElevID, "— clearing disconnected flag.")
		masterID = senderElevID
	}

}

func CheckMasterTimerTimeout() {

	for {
		if checkMasterTimer == nil {
			continue
		}
		select {
		case <-checkMasterTimer.C:
			disconnected = false
			checkMasterTimer = nil
		}
	}
}

func startWatchdogTimer(elevID int, timerDuration int) {
	duration := time.Duration(timerDuration) * time.Second
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
	for range resetInactiveElev {
		startWatchdogTimer(config.ElevID, config.InactiveDuration)
		activeElev[config.ElevID] = true
		elevInactive <- false
		elevInactive <- false
	}

}

func resetNoMasterTimer(timerDuration int) {
	duration := time.Duration(timerDuration) * time.Second
	if masterTimer != nil {
		// Reset the timer; if it wasn't active, drain its channel.
		if !masterTimer.Reset(duration) {
			// Try to drain the channel if necessary.
			select {
			case <-masterTimer.C:
			default:
			}
		}
	} else {
		masterTimer = time.NewTimer(duration)
	}
}

// goroutine
func CheckThereAreOnlySlaves() {

	for range masterTimer.C {

		disconnected = false
		masterTimer = nil

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
						activeElev[i] = false
						aliveElev[i] = false
						elevDied <- i
						fmt.Print("Elevator disc", i, "\n")
						ChangeMaster(setMaster, i)
					} else {
						fmt.Printf("Elevator %d inactive \n", i)
						activeElev[i] = false
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

func numActiveElev() int {
	numActiveElev := 0
	for i := 0; i < len(activeElev); i++ {
		if activeElev[i] {
			numActiveElev++
		}
	}
	return numActiveElev
}

func ChangeMaster(setMaster chan bool, disconnectedElevID int) {

	if numActiveElev() == 1 {
		setMaster <- true
		setMaster <- true
		masterID = config.ElevID
		return
	}

	if disconnectedElevID == masterID {
		if config.ElevID == 0 {
			setMaster <- true
			setMaster <- true
			masterID = config.ElevID
			return
		}

		for j := 0; j < config.ElevID; j++ {
			if aliveElev[j] {
				masterID = -1
				return
			}
		}

		setMaster <- true
		setMaster <- true
		masterID = config.ElevID
	}
}
