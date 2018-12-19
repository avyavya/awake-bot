package hoge

type Message struct {
	UserId    string
	Text      string
	Timestamp string
}

func New(userId string, text string, timestamp string) *Message {
	m := Message{userId, text, timestamp}
	return &m
}
