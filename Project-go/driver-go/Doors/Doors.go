package doors

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

var (
	isDoorOpen =false
)

func  OpenDoor(floor int) int {
	
	if floor>=0{

		isDoorOpen=true
		elevio.SetDoorOpenLamp(true)
		fmt.Printf("Open door\n")
		time.Sleep(3* time.Second)
		// elevio.SetDoorOpenLamp(false)
		fmt.Printf("closing door\n")

	return 1
	} else {
		return 0
	}
	
}

func CloseDoor(floor int, obstr bool) int {
	for {

        // Check if there is an obstruction
        if obstr==true {
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

