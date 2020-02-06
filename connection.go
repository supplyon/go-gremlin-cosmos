package gremtune

import (
	"net/http"
	"time"

	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type dialer interface {
	connect() error
	IsConnected() bool
	IsDisposed() bool
	write([]byte) error
	read() (int, []byte, error)
	close() error
	getAuth() *auth
	ping(errs chan error)
}

// Websocket is the dialer for a WebSocket connection
type Websocket struct {
	host         string
	conn         *websocket.Conn
	auth         *auth
	disposed     bool
	connected    bool
	pingInterval time.Duration
	writingWait  time.Duration
	readingWait  time.Duration

	// timeout for the initial handshake
	timeout time.Duration

	readBufSize  int
	writeBufSize int

	// channel for quit notification
	quit chan struct{}
	sync.RWMutex
}

//Auth is the container for authentication data of dialer
type auth struct {
	username string
	password string
}

// NewDialer returns a WebSocket dialer to use when connecting to Gremlin Server
func NewDialer(host string, configs ...DialerConfig) dialer {
	websocket := &Websocket{
		timeout:      5 * time.Second,
		pingInterval: 60 * time.Second,
		writingWait:  15 * time.Second,
		readingWait:  15 * time.Second,
		connected:    false,
		quit:         make(chan struct{}),
		readBufSize:  8192,
		writeBufSize: 8192,
	}

	for _, conf := range configs {
		conf(websocket)
	}

	websocket.host = host
	return websocket
}

func (ws *Websocket) connect() (err error) {
	d := websocket.Dialer{
		WriteBufferSize:  ws.writeBufSize,
		ReadBufferSize:   ws.readBufSize,
		HandshakeTimeout: ws.timeout, // Timeout or else we'll hang forever and never fail on bad hosts.
	}
	ws.conn, _, err = d.Dial(ws.host, http.Header{})
	if err != nil {

		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		ws.host = ws.host + "/gremlin"
		ws.conn, _, err = d.Dial(ws.host, http.Header{})
	}

	if err == nil {
		ws.connected = true
		ws.conn.SetPongHandler(func(appData string) error {
			ws.connected = true
			return nil
		})
	}
	return
}

// IsConnected returns whether the underlying websocket is connected
func (ws *Websocket) IsConnected() bool {
	return ws.connected
}

// IsDisposed returns whether the underlying websocket is disposed
func (ws *Websocket) IsDisposed() bool {
	return ws.disposed
}

func (ws *Websocket) write(msg []byte) (err error) {
	err = ws.conn.WriteMessage(2, msg)
	return
}

func (ws *Websocket) read() (msgType int, msg []byte, err error) {
	msgType, msg, err = ws.conn.ReadMessage()
	return
}

func (ws *Websocket) close() (err error) {
	defer func() {
		close(ws.quit)
		ws.conn.Close()
		ws.disposed = true
	}()

	err = ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")) //Cleanly close the connection with the server
	return
}

func (ws *Websocket) getAuth() *auth {
	if ws.auth == nil {
		panic("You must create a Secure Dialer for authenticate with the server")
	}
	return ws.auth
}

func (ws *Websocket) ping(errs chan error) {
	ticker := time.NewTicker(ws.pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			connected := true
			if err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(ws.writingWait)); err != nil {
				errs <- err
				connected = false
			}
			ws.Lock()
			ws.connected = connected
			ws.Unlock()

		case <-ws.quit:
			return
		}
	}
}

func (c *Client) writeWorker(errs chan error, quit chan struct{}) { // writeWorker works on a loop and dispatches messages as soon as it receives them
	for {
		select {
		case msg := <-c.requests:
			c.Lock()
			err := c.conn.write(msg)
			if err != nil {
				errs <- err
				c.Errored = true
				c.Unlock()
				break
			}
			c.Unlock()

		case <-quit:
			return
		}
	}
}

func (c *Client) readWorker(errs chan error, quit chan struct{}) { // readWorker works on a loop and sorts messages as soon as it receives them
	for {
		msgType, msg, err := c.conn.read()
		if msgType == -1 { // msgType == -1 is noFrame (close connection)
			return
		}
		if err != nil {
			errs <- errors.Wrapf(err, "Receive message type: %d", msgType)
			c.Errored = true
			break
		}
		if msg != nil {
			// FIXME: At the moment the error returned by handle response is just ignored.
			err = c.handleResponse(msg)
			_ = err
		}

		select {
		case <-quit:
			return
		default:
			continue
		}
	}
}
