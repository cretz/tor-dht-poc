package websocket

import (
	"net"
)

func StartNewListener(l net.Listener) (net.Listener, error) {
	malist := &listener{
		Listener: l,
		incoming: make(chan *Conn),
		closed:   make(chan struct{}),
	}
	go malist.serve()
	return malist, nil
}
