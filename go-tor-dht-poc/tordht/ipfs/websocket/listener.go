package websocket

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type listener struct {
	net.Listener

	closed   chan struct{}
	incoming chan *Conn
}

// Default gorilla upgrader
var upgrader = websocket.Upgrader{
	// Allow requests from *all* origins.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (l *listener) serve() {
	defer close(l.closed)
	http.Serve(l.Listener, l)
}

func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade websocket", 400)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	var cnCh <-chan bool
	if cn, ok := w.(http.CloseNotifier); ok {
		cnCh = cn.CloseNotify()
	}

	wscon := NewConn(c, cancel)
	// Just to make sure.
	defer wscon.Close()

	select {
	case l.incoming <- wscon:
	case <-l.closed:
		c.Close()
		return
	case <-cnCh:
		return
	}

	// wait until conn gets closed, otherwise the handler closes it early
	select {
	case <-ctx.Done():
	case <-l.closed:
		c.Close()
		return
	case <-cnCh:
		return
	}
}

func (l *listener) Accept() (net.Conn, error) {
	select {
	case c, ok := <-l.incoming:
		if !ok {
			return nil, fmt.Errorf("listener is closed")
		}
		return c, nil
	case <-l.closed:
		return nil, fmt.Errorf("listener is closed")
	}
}
