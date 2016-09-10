package server

type Communication struct {
	Command string `json:"command"`
	SentAt  uint   `json:"sent_at"`
}

type Connection struct {
	Communication

	Username string `json:"username"`
}

type Ping struct {
	Communication
}

type Pong struct {
	Communication

	ReceivedAt uint `json:"received_at"`
}
