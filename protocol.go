package server

const (
	ConnectRequestCmd = "connect_request"
	ConnectVerdictCmd = "connect_verdict"
	PingCmd           = "ping"
	PongCmd           = "pong"
)

type Communication struct {
	Command string `json:"command"`
	SentAt  uint   `json:"sent_at"`
}

type ConnectRequest struct {
	Communication

	Username string `json:"username"`
}

type ConnectVerdict struct {
	Communication

	CanProceed bool   `json:"can_proceed"`
	Message    string `json:"message"`
}

type Ping struct {
	Communication
}

type Pong struct {
	Communication

	ReceivedAt uint `json:"received_at"`
}
