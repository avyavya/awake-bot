package timeout

import (
	"log"
	"time"
)

type Timeout struct {
	sec       int
	onTimeout func(roomId string, sec int)
	roomId    string
	canceled  bool
}

func (to *Timeout) Cancel() {
	log.Printf("[info] Snooze canceled.")
	to.canceled = true
}

func New(f func(string, int), timeout int, roomId string) *Timeout {
	to := Timeout{timeout, f, roomId, false}
	go setTimeout(&to)
	return &to
}

// invoke func after timeout sec
func setTimeout(to *Timeout) {
	time.Sleep(time.Duration(to.sec) * time.Second)

	if to.canceled {
		return
	}

	log.Printf("[info] Timed-out %d", to.sec)
	to.onTimeout(to.roomId, to.sec)
}
