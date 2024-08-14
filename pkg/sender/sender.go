package sender

type Sender interface {
	SendMessage(msg string, recipient string)
}

type EmailSender struct {
	sender string
}

func NewESender(sender string) EmailSender {
	return EmailSender{sender: sender}
}

func (e EmailSender) SendMessage(msg string, recipient string) {
	return
}
