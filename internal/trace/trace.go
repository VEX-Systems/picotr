package trace

import (
	"context"
	"math"
	"net"
	"sync"
	"time"

	"github.com/VEX-Systems/picotr/internal/probe"
)

type HopStats struct {
	TTL      int
	Addr     net.IP
	Host     string
	Sent     int
	Received int
	Lost     int
	Last     float64
	Best     float64
	Worst    float64
	Avg      float64
	sum      float64
	Reached  bool
}

func (e *Engine) Snapshot() []HopStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]HopStats, len(e.hops))
	copy(out, e.hops)
	return out
}

func (e *Engine) MaxTTL() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.maxTTL
}

type Config struct {
	DstIP    net.IP
	MaxHops  int
	Interval time.Duration
}

type Engine struct {
	prober *probe.Prober
	cfg    Config

	mu     sync.RWMutex
	hops   []HopStats
	maxTTL int
}

func New(prober *probe.Prober, cfg Config) *Engine {
	hops := make([]HopStats, cfg.MaxHops)
	for i := range hops {
		hops[i].TTL = i + 1
		hops[i].Best = math.MaxFloat64
	}
	return &Engine{
		prober: prober,
		cfg:    cfg,
		hops:   hops,
	}
}

func (e *Engine) Run(ctx context.Context) {
	e.probeRound(ctx)

	ticker := time.NewTicker(e.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.probeRound(ctx)
		}
	}
}

func (e *Engine) probeRound(ctx context.Context) {
	e.mu.RLock()
	maxTTL := e.maxTTL
	e.mu.RUnlock()

	limit := maxTTL
	if limit == 0 {
		limit = e.cfg.MaxHops
	}

	var wg sync.WaitGroup
	for ttl := 1; ttl <= limit; ttl++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		go func(ttl int) {
			defer wg.Done()

			result, _ := e.prober.Send(e.cfg.DstIP, ttl)

			e.mu.Lock()
			h := &e.hops[ttl-1]
			h.Sent++

			if result.Addr != nil {
				h.Addr = result.Addr
				h.Received++
				h.Last = result.RTT
				if result.RTT < h.Best {
					h.Best = result.RTT
				}
				if result.RTT > h.Worst {
					h.Worst = result.RTT
				}
				h.sum += result.RTT
				h.Avg = h.sum / float64(h.Received)
			} else {
				h.Lost++
			}

			if result.Reached {
				h.Reached = true
				if e.maxTTL == 0 || ttl < e.maxTTL {
					e.maxTTL = ttl
				}
			}
			e.mu.Unlock()
		}(ttl)
	}
	wg.Wait()
}
