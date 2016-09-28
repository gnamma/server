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
	RegisterNodeCmd       = "register_node"
	RegisteredNodeCmd     = "registered_node"
	UpdateNodeCmd         = "update_node"
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

type RegisterNode struct {
	Communication

	Node Node `json:"node"`
	PID  uint `json:"pid"`
}

type RegisteredNode struct {
	Communication

	NID uint `json:"nid"`
}

type UpdateNode struct {
	Communication

	PID uint
	NID uint

	Position Point
	Rotation Point
}

type Preparer interface {
	Prepare(string)
}
