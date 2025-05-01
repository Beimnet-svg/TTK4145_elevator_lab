package elevio

import (
	config "Project-go/Config"
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = config.PollRate * time.Millisecond

var(
	_initialized bool = false
	_numFloors int = config.NumberFloors
	_mtx sync.Mutex
	_conn net.Conn
)


type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ElevatorBehaviour int

const (
	EB_Idle     ElevatorBehaviour = 0
	EB_Moving                     = 1
	EB_DoorOpen                   = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type Elevator struct {
	CurrentFloor     int
	Direction        MotorDirection
	Behaviour        ElevatorBehaviour
	Requests         [config.NumberFloors][config.NumberBtn]int
	ActiveOrders     [config.NumberFloors][config.NumberBtn]bool
	NumFloors        int
	DoorOpenDuration int
	ElevatorID       int
	Master           bool
	Obstruction      bool
	Inactive         bool
}

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}

	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func LightButtons(AllActiveOrders [config.NumberElev][config.NumberFloors][config.NumberBtn]bool, elevID int) {

	for floor := 0; floor < config.NumberFloors; floor++ {
		for button := ButtonType(0); button <= ButtonType(1); button++ {
			for elev := 0; elev < config.NumberElev; elev++ {
				if AllActiveOrders[elev][floor][button] {
					SetButtonLamp(button, floor, true)
					break
				} else {
					SetButtonLamp(button, floor, false)
				}
			}
		}
		if AllActiveOrders[elevID][floor][ButtonType(2)] {
			SetButtonLamp(ButtonType(2), floor, true)
		} else {
			SetButtonLamp(ButtonType(2), floor, false)
		}
	}
}

func AddToQueue(button ButtonType, floor int, e Elevator, orderCount int) Elevator {
	e.Requests[floor][button] = orderCount
	return e
}

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < config.NumberBtn; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
