
package timer

import (
	"time"
)

var _timer *time.Timer
var _timerActive bool

func StartTimer(duration int) {
    _timer = time.NewTimer(time.Duration(duration) * time.Second)
    _timerActive = true
}

func StopTimer() {
    if _timer != nil {
        _timer.Stop()
    }
    _timerActive = false
}

func TimerTimeOut() bool {
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
