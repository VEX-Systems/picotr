//go:build !windows

package probe

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type pendingProbe struct {
	sent time.Time
	ch   chan Result
}

type unixProber struct {
	conn    *icmp.PacketConn
	timeout time.Duration
	id      uint16
	seq     atomic.Uint32
	mu      sync.Mutex
	pending map[uint32]*pendingProbe
	sendMu  sync.Mutex
	done    chan struct{}
}

func New(protocol string, timeout time.Duration) (*Prober, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("listen icmp: %w (run as root)", err)
	}

	u := &unixProber{
		conn:    conn,
		timeout: timeout,
		id:      0x91C0,
		pending: make(map[uint32]*pendingProbe),
		done:    make(chan struct{}),
	}
	go u.recvLoop()

	return &Prober{impl: u}, nil
}

func (u *unixProber) send(dst net.IP, ttl int) (Result, error) {
	seq := u.seq.Add(1)

	ch := make(chan Result, 1)
	u.mu.Lock()
	u.pending[seq] = &pendingProbe{sent: time.Now(), ch: ch}
	u.mu.Unlock()

	defer func() {
		u.mu.Lock()
		delete(u.pending, seq)
		u.mu.Unlock()
	}()

	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(u.id),
			Seq:  int(seq),
			Data: []byte("PICOTR"),
		},
	}
	wb, err := msg.Marshal(nil)
	if err != nil {
		return Result{}, err
	}

	u.sendMu.Lock()
	if err := u.conn.IPv4PacketConn().SetTTL(ttl); err != nil {
		u.sendMu.Unlock()
		return Result{}, fmt.Errorf("set ttl: %w", err)
	}
	_, err = u.conn.WriteTo(wb, &net.IPAddr{IP: dst})
	u.sendMu.Unlock()
	if err != nil {
		return Result{}, err
	}

	timer := time.NewTimer(u.timeout)
	defer timer.Stop()

	select {
	case r := <-ch:
		return r, nil
	case <-timer.C:
		return Result{}, nil
	case <-u.done:
		return Result{}, fmt.Errorf("prober closed")
	}
}

func (u *unixProber) close() error {
	close(u.done)
	return u.conn.Close()
}

func (u *unixProber) recvLoop() {
	buf := make([]byte, 1500)
	for {
		select {
		case <-u.done:
			return
		default:
		}

		_ = u.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, peer, err := u.conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case <-u.done:
				return
			default:
				continue
			}
		}

		recvTime := time.Now()
		if n < 4 {
			continue
		}

		var fromIP net.IP
		if addr, ok := peer.(*net.IPAddr); ok {
			fromIP = addr.IP
		}

		raw := buf[:n]
		msgType := raw[0]

		switch msgType {
		case 0:
			if n < 8 {
				continue
			}
			echoID := binary.BigEndian.Uint16(raw[4:6])
			echoSeq := binary.BigEndian.Uint16(raw[6:8])
			if echoID != u.id {
				continue
			}
			seq := uint32(echoSeq)
			u.mu.Lock()
			pp, ok := u.pending[seq]
			if ok {
				pp.ch <- Result{
					Addr:    fromIP,
					RTT:     float64(recvTime.Sub(pp.sent).Microseconds()) / 1000.0,
					Reached: true,
				}
			}
			u.mu.Unlock()

		case 11, 3:
			if n < 8+20+8 {
				continue
			}
			innerIP := raw[8:]
			ihl := int(innerIP[0]&0x0F) * 4
			if ihl < 20 || len(innerIP) < ihl+8 {
				continue
			}
			innerICMP := innerIP[ihl:]
			if innerICMP[0] != 8 {
				continue
			}
			innerID := binary.BigEndian.Uint16(innerICMP[4:6])
			innerSeq := binary.BigEndian.Uint16(innerICMP[6:8])
			if innerID != u.id {
				continue
			}
			seq := uint32(innerSeq)
			reached := msgType == 3
			u.mu.Lock()
			pp, ok := u.pending[seq]
			if ok {
				pp.ch <- Result{
					Addr:    fromIP,
					RTT:     float64(recvTime.Sub(pp.sent).Microseconds()) / 1000.0,
					Reached: reached,
				}
			}
			u.mu.Unlock()
		}
	}
}
