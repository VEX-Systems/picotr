package probe

import "net"

type Result struct {
	Addr    net.IP
	RTT     float64
	Reached bool
	Err     error
}

type Prober struct {
	impl proberImpl
}

type proberImpl interface {
	send(dst net.IP, ttl int) (Result, error)
	close() error
}

func (p *Prober) Send(dst net.IP, ttl int) (Result, error) {
	return p.impl.send(dst, ttl)
}

func (p *Prober) Close() error {
	return p.impl.close()
}
