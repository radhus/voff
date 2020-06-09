package watchdog

import (
	"os"
	"syscall"
	"unsafe"

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

func (d *device) ioctl(signal uintptr, data unsafe.Pointer) syscall.Errno {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		d.file.Fd(),
		signal,
		uintptr(data),
	)
	return errno
}

func (d *device) Kick() error {
	if errno := d.ioctl(unix.WDIOC_KEEPALIVE, nil); errno != 0 {
		return errno
	}
	return nil
}
