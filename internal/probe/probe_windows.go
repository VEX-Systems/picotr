package probe

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"
)

var (
	iphlpapi            = syscall.NewLazyDLL("Iphlpapi.dll")
	procIcmpCreateFile  = iphlpapi.NewProc("IcmpCreateFile")
	procIcmpCloseHandle = iphlpapi.NewProc("IcmpCloseHandle")
	procIcmpSendEcho    = iphlpapi.NewProc("IcmpSendEcho")
)

const (
	ipSuccess             = 0
	ipReqTimedOut         = 11010
	ipTTLExpiredTransit   = 11013
	ipTTLExpiredReassem   = 11014
	ipDestNetUnreachable  = 11002
	ipDestHostUnreachable = 11003
	ipDestProtUnreachable = 11004
	ipDestPortUnreachable = 11005
)

type ipOption struct {
	TTL         uint8
	Tos         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

type icmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       ipOption
}

type windowsProber struct {
	timeoutMs uint32
}

func New(protocol string, timeout time.Duration) (*Prober, error) {
	h, _, err := procIcmpCreateFile.Call()
	if syscall.Handle(h) == syscall.InvalidHandle {
		return nil, fmt.Errorf("IcmpCreateFile: %w", err)
	}
	procIcmpCloseHandle.Call(h)

	return &Prober{
		impl: &windowsProber{
			timeoutMs: uint32(timeout.Milliseconds()),
		},
	}, nil
}

func (w *windowsProber) send(dst net.IP, ttl int) (Result, error) {
	dst4 := dst.To4()
	if dst4 == nil {
		return Result{}, fmt.Errorf("only IPv4 supported")
	}
	destAddr := binary.LittleEndian.Uint32(dst4)

	h, _, err := procIcmpCreateFile.Call()
	if syscall.Handle(h) == syscall.InvalidHandle {
		return Result{}, fmt.Errorf("IcmpCreateFile: %w", err)
	}
	defer procIcmpCloseHandle.Call(h)

	payload := []byte("PICOTR")
	opts := ipOption{TTL: uint8(ttl)}

	replySize := unsafe.Sizeof(icmpEchoReply{}) + uintptr(len(payload)) + 8 + 64
	replyBuf := make([]byte, replySize)

	ret, _, _ := procIcmpSendEcho.Call(
		h,
		uintptr(destAddr),
		uintptr(unsafe.Pointer(&payload[0])),
		uintptr(len(payload)),
		uintptr(unsafe.Pointer(&opts)),
		uintptr(unsafe.Pointer(&replyBuf[0])),
		uintptr(replySize),
		uintptr(w.timeoutMs),
	)

	if ret == 0 {
		return Result{}, nil
	}

	reply := (*icmpEchoReply)(unsafe.Pointer(&replyBuf[0]))

	addrBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(addrBytes, reply.Address)
	fromIP := net.IP(addrBytes)
	rtt := float64(reply.RoundTripTime)

	switch reply.Status {
	case ipSuccess:
		return Result{Addr: fromIP, RTT: rtt, Reached: true}, nil
	case ipTTLExpiredTransit, ipTTLExpiredReassem:
		return Result{Addr: fromIP, RTT: rtt}, nil
	case ipReqTimedOut:
		return Result{}, nil
	case ipDestNetUnreachable, ipDestHostUnreachable, ipDestProtUnreachable, ipDestPortUnreachable:
		return Result{Addr: fromIP, RTT: rtt, Reached: true}, nil
	default:
		return Result{Addr: fromIP, RTT: rtt}, nil
	}
}

func (w *windowsProber) close() error {
	return nil
}
