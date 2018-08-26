// +build !linux

package watchdog

import "errors"

func Open(path string) (Device, error) {
	return nil, errors.New("Platform not supported")
}
