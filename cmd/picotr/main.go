package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/VEX-Systems/picotr/internal/output"
	"github.com/VEX-Systems/picotr/internal/probe"
	"github.com/VEX-Systems/picotr/internal/term"
	"github.com/VEX-Systems/picotr/internal/trace"
)

var version = "dev"

func main() {
	var (
		maxHops  int
		timeout  time.Duration
		interval time.Duration
		numeric  bool
		showVer  bool
	)

	flag.IntVar(&maxHops, "m", 30, "max number of hops")
	flag.DurationVar(&timeout, "w", 3*time.Second, "timeout for each probe")
	flag.DurationVar(&interval, "i", 1*time.Second, "interval between probe rounds")
	flag.BoolVar(&numeric, "n", false, "do not resolve hostnames")
	flag.BoolVar(&showVer, "version", false, "show version")
	flag.Usage = usage
	flag.Parse()

	if showVer {
		fmt.Printf("picotr %s\n", version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	target := flag.Arg(0)
	ips, err := net.LookupIP(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "picotr: cannot resolve %s: %v\n", target, err)
		os.Exit(1)
	}

	var dstIP net.IP
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			dstIP = ipv4
			break
		}
	}
	if dstIP == nil {
		dstIP = ips[0]
	}

	localIP := detectLocalIP(dstIP)

	prober, err := probe.New("icmp", timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "picotr: %v\n", err)
		os.Exit(1)
	}
	defer prober.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := trace.Config{
		DstIP:    dstIP,
		MaxHops:  maxHops,
		Interval: interval,
	}

	engine := trace.New(prober, cfg)

	go engine.Run(ctx)

	display := output.New(output.Config{
		Target:  target,
		DstIP:   dstIP,
		LocalIP: localIP,
		Engine:  engine,
		Numeric: numeric,
	})

	restore, err := term.EnableRawMode()
	if err == nil {
		defer restore()
	}

	go func() {
		buf := make([]byte, 1)
		for {
			_, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			switch buf[0] {
			case 'q', 'Q', 3:
				cancel()
				return
			case 'd', 'D':
				display.ToggleIP()
			case 's', 'S':
				display.ToggleSimple()
			case 'c', 'C':
				display.ToggleColor()
			case 'u', 'U':
				display.ToggleUnit()
			case 'p', 'P':
				display.ToggleFloat()
			case 'x', 'X':
				display.ToggleShowAll()
			case 'a', 'A':
				if display.InRoute() {
					display.ExportAllHopsPNG()
				} else {
					display.ToggleAS()
				}
			case 'e', 'E':
				display.ExportPNG()
			case 't', 'T':
				display.ExportTrace()
			case 'r', 'R':
				display.ToggleRoute()
			case 127, 8:
				display.ExitRoute()
			}
		}
	}()

	display.Run(ctx, 250*time.Millisecond)
}

func detectLocalIP(dst net.IP) net.IP {
	conn, err := net.Dial("udp4", dst.String()+":80")
	if err != nil {
		return nil
	}
	defer conn.Close()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return addr.IP
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `PicoTR - PicoTraceroute: fast and efficient tracerouting tool

Usage: picotr [options] <target>

Options:
`)
	flag.PrintDefaults()
}
