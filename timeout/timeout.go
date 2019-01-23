package timeout

import (
	"log"
	"time"
)

type Timeout struct {
	Sec       int
	RoomId    string
	userId    string
	Repeated  int
	canceled  bool
	onTimeout func(*Timeout)
}

func New(f func(*Timeout), timeout int, roomId string, userId string) *Timeout {
	to := Timeout{timeout, roomId, userId, 0, false, f}
	go setTimeout(&to)
	return &to
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
