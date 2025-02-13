package timer

import "time"

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