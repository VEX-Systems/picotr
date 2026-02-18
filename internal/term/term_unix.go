//go:build !windows

package term

import (
	"os"
	"syscall"
	"unsafe"
)

type termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]byte
	Ispeed uint32
	Ospeed uint32
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

const (
	icanon    = 0x00000002
	echo      = 0x00000008
	tcgets    = 0x5401
	tcsets    = 0x5402
	tiocgwinsz = 0x5413
)

func GetWidth() int {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(tiocgwinsz), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return 120
	}
	return int(ws.Col)
}

func EnableRawMode() (func(), error) {
	fd := os.Stdin.Fd()

	var old termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(tcgets), uintptr(unsafe.Pointer(&old)))
	if errno != 0 {
		return nil, errno
	}

	raw := old
	raw.Lflag &^= icanon | echo

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(tcsets), uintptr(unsafe.Pointer(&raw)))
	if errno != 0 {
		return nil, errno
	}

	restore := func() {
		syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(tcsets), uintptr(unsafe.Pointer(&old)))
	}

	return restore, nil
}
