package server

import "time"

const (
	ConnectRequestCmd     = "connect_request"
	ConnectVerdictCmd     = "connect_verdict"
	PingCmd               = "ping"
	PongCmd               = "pong"
	EnvironmentRequestCmd = "environment_request"
	EnvironmentPackageCmd = "environment_package"
	AssetRequestCmd       = "asset_request"
)

type Communication struct {
	Command string `json:"command"`
	SentAt  int64  `json:"sent_at"`
}

func (c *Communication) Prepare(cmd string) {
	c.Command = cmd
	c.SentAt = time.Now().UnixNano()
}

type ConnectRequest struct {
	Communication

	Username string `json:"username"`
}

type ConnectVerdict struct {
	Communication

	CanProceed bool   `json:"can_proceed"`
	Message    string `json:"message"`
	PlayerID   uint   `json:"player_id"`
}

type Ping struct {
	Communication
}

type Pong struct {
	Communication

	ReceivedAt int64 `json:"received_at"`
}

type Preparer interface {
	Prepare(string)
}

type EnvironmentRequest struct {
	Communication
}

type EnvironmentPackage struct {
	Communication

	AssetKeys map[string]string `json:"asset_keys"`
	Main      string            `json:"main"`
}

type AssetRequest struct {
	Communication

	Key string
}
