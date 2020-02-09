package interfaces

type Dialer interface {
	Connect() error
	IsConnected() bool
	IsDisposed() bool
	Write([]byte) error
	Read() (int, []byte, error)
	Close() error
	GetAuth() Auth
	Ping() error
	GetQuitChannel() <-chan struct{}
}

//Auth is the container for authentication data of dialer
type Auth struct {
	Username string
	Password string
}