package api

import (
	"sync"

	"github.com/gorilla/websocket"
)

// wsConn wraps a gorilla *websocket.Conn with a write mutex.
//
// gorilla/websocket allows one concurrent reader and one concurrent writer,
// but not two concurrent writers -- a second write while another is in
// flight panics with "concurrent write to websocket connection". Both hubs
// fan the same connection out to many goroutines (every handler that calls
// broadcastAll/notify), so every write has to go through this lock. Reads
// stay on the bare Conn in the handler's read loop, which is a separate
// goroutine and allowed to run alongside a write.
type wsConn struct {
	conn *websocket.Conn
	wmu  sync.Mutex
}

func newWSConn(conn *websocket.Conn) *wsConn {
	return &wsConn{conn: conn}
}

func (c *wsConn) writeJSON(v any) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *wsConn) close() error {
	return c.conn.Close()
}
