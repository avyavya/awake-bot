package timeout

import (
	"log"
	"time"
)

type Timeout struct {
	onTimeout   func(*Timeout)
	Sec         int
	RoomId      string
	userId      string
	AlertRoomId string
	Repeated    int
	canceled    bool
}

func New(f func(*Timeout), timeout int, roomId string, userId string, alertRoomId string) *Timeout {
	to := Timeout{f, timeout, roomId, userId, alertRoomId, 0, false}
	go setTimeout(&to)
	return &to
}

func NewTimeout(f func(), timeout int) {
	go invokeAfter(f, timeout)
}

func invokeAfter(f func(), timeout int) {
	time.Sleep(time.Duration(timeout) * time.Second)

	f()
}

// invoke func after timeout sec
func setTimeout(to *Timeout) {
	time.Sleep(time.Duration(to.Sec) * time.Second)

	if to.canceled {
		return
	}

	log.Printf("[info] Timed-out %d", to.Sec)
	to.onTimeout(to)
}

func (to *Timeout) GetMonitoringUserId() string {
	return to.userId
}

func (to *Timeout) Snooze() {
	to.Repeated++
	go setTimeout(to)
}

func (to *Timeout) Stop() {
	log.Printf("[info] Snooze canceled.")
	to.canceled = true
}
