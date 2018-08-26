package watchdog

type Device interface {
	Close() error
	Poke() error
}
