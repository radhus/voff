package watchdog

type Device interface {
	Close() error
	Kick() error
}
