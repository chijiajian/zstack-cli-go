//go:build linux || darwin

package config

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func tryConfigLock(file *os.File) (bool, error) {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) {
		return false, nil
	}
	return err == nil, err
}
func unlockConfig(file *os.File) { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }
