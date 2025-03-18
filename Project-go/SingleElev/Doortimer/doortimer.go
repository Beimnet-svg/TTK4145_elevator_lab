package doortimer

import (
	"time"
)

var _timer *time.Timer
var _timerActive bool

func StartDoorTimer(duration int) {
	_timer = time.NewTimer(time.Duration(duration) * time.Second)
	_timerActive = true
}

func StopDoorTimer() {
	if _timer != nil {
		_timer.Stop()
	}
	_timerActive = false
}

func doorTimerTimeOut() bool {
	if !_timerActive {
		return false
	}
	select {
	case <-_timer.C:
		_timerActive = false
		return true
	default:
		return false
	}
}

func PollDoorTimer(doorTimer chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			doorTimer <- doorTimerTimeOut()
		}
	}
}
