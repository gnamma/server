package server

type CommunicationHandler func(conn *ChildConn) error

type Dispatch struct {
	H map[string]CommunicationHandler // Do not change at runtime!
}

func (d *Dispatch) Handle(cmd string, conn *ChildConn) error {
	f, ok := d.H[cmd]
	if !ok {
		return ErrHandlerNotFound
	}

	return f(conn)
}
