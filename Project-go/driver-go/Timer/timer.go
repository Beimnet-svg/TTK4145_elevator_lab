package timer

import (
	"time"
)

var _timerEndTime int
var _timerActive bool

func StartTimer(duration int) {
	_timerEndTime = duration + time.Now().Second()
	_timerActive = true
}

func StopTimer() {
	_timerActive = false
}

func TimerTimeOut() bool {
	return _timerActive && (time.Now().Second() >= _timerEndTime)

}

func PollTimer(receiver chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			receiver <- TimerTimeOut()
		}
	}
}
