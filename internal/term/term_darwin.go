//go:build darwin

package term

import (
	"os"

	"golang.org/x/sys/unix"
)

func GetWidth() int {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		return 120
	}
	return int(ws.Col)
}

func EnableRawMode() (func(), error) {
	fd := int(os.Stdin.Fd())

	oldState, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil, err
	}

	raw := *oldState
	raw.Lflag &^= unix.ICANON | unix.ECHO

	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, &raw); err != nil {
		return nil, err
	}

	restore := func() {
		_ = unix.IoctlSetTermios(fd, unix.TIOCSETA, oldState)
	}

	return restore, nil
}


