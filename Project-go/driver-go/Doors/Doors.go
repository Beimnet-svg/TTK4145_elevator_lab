package doors

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

var (
	isDoorOpen = false
)

func OpenDoor(targetFloor int, floor int, b []elevio.ButtonEvent, numFloors int, drv_button chan elevio.ButtonEvent) []elevio.ButtonEvent {

	isDoorOpen = true
	elevio.SetDoorOpenLamp(true)
	fmt.Printf("Open door\n")
	startTime := time.Now()

	for {
		select {
		case a := <-drv_button:
			b = elevio.AddToQueue(a.Button, a.Floor, b)
			elevio.LightButtons(b, numFloors)
		default:
		}
		elapsed := time.Since(startTime)

		if elapsed >= 3*time.Second {
			break
		}
	}

	elevio.SetDoorOpenLamp(false)
	fmt.Printf("closing door\n")

	return b
}

func CloseDoor(floor int, obstr bool) int {
	for {

		// Check if there is an obstruction
		if obstr == true {
			// Wait until the obstruction is cleared
			for obstr {
				// Keep waiting until the channel sends false
			}
		} else {
			// If no obstruction, close the door
			elevio.SetDoorOpenLamp(false)
			break
		}
	}
	return 1
}
