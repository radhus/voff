package watchdog

import (
	"bytes"
	"log"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type device struct {
	file *os.File
}

type linux_info struct {
	_        uint32 // options
	_        uint32 // firmware_version
	identity [32]byte
}

func Open(path string) (Device, error) {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}

	d := &device{
		file: file,
	}

	info := linux_info{}
	errno := d.ioctl(unix.WDIOC_GETSUPPORT, unsafe.Pointer(&info))

	if errno != 0 {
		return nil, errno
	}

	identity := string(bytes.Trim(info.identity[:], "\x00"))
	log.Println("Watchdog identity:", identity)

	return d, nil
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
