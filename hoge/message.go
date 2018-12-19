package hoge

type Message2 struct {
	UserId    string
	Text      string
	Timestamp string
}

func NewMessage(userId string, text string, timestamp string) *Message2 {
	m := Message2{userId, text, timestamp}
	return &m
}
