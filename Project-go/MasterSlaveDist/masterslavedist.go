package masterslavedist

import (
	config "Project-go/Config"
	elevio "Project-go/SingleElev/Elevio"
	"fmt"
	"time"
)

var (
	watchdogTimers         [config.NumberElev]*time.Timer
	waitForMasterMsg       *time.Timer
	aliveMasterTimer       *time.Timer
	waitForMasterMsgActive = false

	activeElev [config.NumberElev]bool
	aliveElev  [config.NumberElev]bool

	disconnected = false
	masterID     = -1 // -1 means master unknown
)

func initializeTimers() {
	for i := 0; i < len(watchdogTimers); i++ {
		watchdogTimers[i] = time.NewTimer(1 * time.Second)
		watchdogTimers[i].Stop()
	}
	waitForMasterMsg = time.NewTimer(1 * time.Second)
	waitForMasterMsg.Stop()
	aliveMasterTimer = time.NewTimer(1 * time.Second)
	aliveMasterTimer.Stop()
}

func InitializeMasterSlaveDist(localElev elevio.Elevator, activeOrdersArrived chan [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, setMaster chan bool) {

	activeElev[config.ElevID] = true
	aliveElev[config.ElevID] = true

	for ID := 0; ID < len(watchdogTimers); ID++ {
		// If different ID than our own we look at time between alive messages
		if ID != config.ElevID {
			watchdogTimers[ID] = resetTimer(watchdogTimers[ID], config.WatchdogDuration*time.Second)
		}
		// If same ID we look at time since started operation
		if ID == config.ElevID {
			watchdogTimers[ID] = resetTimer(watchdogTimers[ID], config.InactiveDuration*time.Second)
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

	if disconnected && !waitForMasterMsgActive {
		fmt.Println("Starting checkMasterTimer")
		waitForMasterMsg = time.NewTimer(config.WatchdogDuration * time.Second)
		waitForMasterMsgActive = true
	}

	if senderE.Inactive {
		activeElev[senderElevID] = false
	} else {
		activeElev[senderElevID] = true
	}

	aliveElev[senderElevID] = true
	watchdogTimers[senderElevID] = resetTimer(watchdogTimers[senderElevID], config.WatchdogDuration*time.Second)

}

func AliveRecievedFromMaster(senderElevID int, inactive bool, localElev elevio.Elevator, setMaster chan bool) {

	aliveMasterTimer = resetTimer(aliveMasterTimer, 2*config.WatchdogDuration*time.Second)

	if masterID == -1 {
		masterID = senderElevID
	}

	if inactive {
		activeElev[senderElevID] = false

	} else {
		activeElev[senderElevID] = true

	}

	aliveElev[senderElevID] = true
	watchdogTimers[senderElevID] = resetTimer(watchdogTimers[senderElevID], config.WatchdogDuration*time.Second)

	if localElev.Master {
		resolveMasterConflict(senderElevID, setMaster)
	}

}

func resolveMasterConflict(senderElevID int, setMaster chan bool) {

	if disconnected {
		setMaster <- false
		setMaster <- false
		disconnected = false

		waitForMasterMsg.Stop()
		waitForMasterMsgActive = false

		fmt.Println("Received heartbeat from elevator", senderElevID, "— clearing disconnected flag.")
		masterID = senderElevID
	}

}

func CheckTimerTimout(setMaster chan bool, elevDied chan int, elevInactive chan bool) {
	for {
		if waitForMasterMsg == nil {
			initializeTimers()
		}
		select {
		case <-waitForMasterMsg.C:
			disconnected = false
		case <-aliveMasterTimer.C:
			applyMaster(setMaster)
		default:
			for i := 0; i < len(watchdogTimers); i++ {
				select {
				case <-watchdogTimers[i].C:
					if i != config.ElevID {
						activeElev[i] = false
						aliveElev[i] = false
						elevDied <- i
						fmt.Print("Elevator disc", i, "\n")
						changeMaster(setMaster, i)
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

func resetTimer(timer *time.Timer, duration time.Duration) *time.Timer {
	if !timer.Reset(duration) {
		// Try to drain the channel if necessary.
		select {
		case <-timer.C:
		default:
		}
	}
	return timer
}

func ResetInactiveTimer(resetInactiveElev chan int, elevInactive chan bool) {
	for range resetInactiveElev {
		watchdogTimers[config.ElevID] = resetTimer(watchdogTimers[config.ElevID], config.InactiveDuration*time.Second)
		
			activeElev[config.ElevID] = true
			elevInactive <- false
			elevInactive <- false
		
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

func applyMaster(setMaster chan bool) {
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

func changeMaster(setMaster chan bool, disconnectedElevID int) {

	if numActiveElev() == 1 {
		setMaster <- true
		setMaster <- true
		masterID = config.ElevID
		disconnected = true
		return
	}

	if disconnectedElevID == masterID {
		applyMaster(setMaster)
	}
}
