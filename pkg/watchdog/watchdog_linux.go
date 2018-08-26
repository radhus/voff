package watchdog

import (
	"os"

	"golang.org/x/sys/unix"
)

type device struct {
	file *os.File
}

func Open(path string) (Device, error) {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}

	return &device{
		file: file,
	}, nil
}

func (d *device) Close() error {
	return d.file.Close()
}

func (d *device) Poke() error {
	// Borrowed from github.com/gokrazy/gokrazy
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		d.file.Fd(),
		unix.WDIOC_KEEPALIVE,
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}
