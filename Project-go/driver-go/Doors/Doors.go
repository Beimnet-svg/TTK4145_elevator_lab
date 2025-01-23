package doors

import (
	"Driver-go/elevio"
	"time"
)

var (
	isDoorOpen =false
)

func  OpenDoor(drv_floors chan int) int {
	floor:=<-drv_floors
	if floor==0{

		isDoorOpen=true
		elevio.SetDoorOpenLamp(true)
		time.Sleep(3* time.Second)

		return 1
	} else {
		return 0
	}
	
}

func CloseDoor(drv_floors chan int, drv_obstr chan bool) int {
	for {

        // Check if there is an obstruction
        if <-drv_obstr {
            // Wait until the obstruction is cleared
            for obstruction := <-drv_obstr; obstruction; obstruction = <-drv_obstr {
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

