package config

const (
	NumberFloors = 4
	NumberElev   = 3
	NumberBtn    = 3

	DoorOpenDuration = 3
	WatchdogDuration = 2
	InactiveDuration = 10

	SendDelay = 50
	PollRate  = 20
)

var ElevID = -1

func SetElevID(id int) {
	ElevID = id
}
