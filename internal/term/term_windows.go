package term

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode          = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode          = kernel32.NewProc("SetConsoleMode")
	procGetConsoleScreenBufInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type consoleScreenBufferInfo struct {
	Size       [2]int16
	CursorPos  [2]int16
	Attributes uint16
	Window     [4]int16
	MaxSize    [2]int16
}

func GetWidth() int {
	handle := syscall.Handle(os.Stdout.Fd())
	var info consoleScreenBufferInfo
	r, _, _ := procGetConsoleScreenBufInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 120
	}
	w := int(info.Window[2]-info.Window[0]) + 1
	if w < 40 {
		return 120
	}
	return w
}

const (
	enableLineInput       = 0x0002
	enableEchoInput       = 0x0004
	enableProcessedInput  = 0x0001
)

func EnableRawMode() (func(), error) {
	handle := syscall.Handle(os.Stdin.Fd())

	var oldMode uint32
	r, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&oldMode)))
	if r == 0 {
		return nil, err
	}

	newMode := oldMode &^ (enableLineInput | enableEchoInput)
	newMode |= enableProcessedInput
	r, _, err = procSetConsoleMode.Call(uintptr(handle), uintptr(newMode))
	if r == 0 {
		return nil, err
	}

	restore := func() {
		procSetConsoleMode.Call(uintptr(handle), uintptr(oldMode))
	}

	return restore, nil
}
