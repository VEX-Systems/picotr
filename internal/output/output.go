package output

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/VEX-Systems/picotr/internal/term"
	"github.com/VEX-Systems/picotr/internal/trace"
)

type asInfo struct {
	ASN    uint32
	ASName string
}

const (
	reset = "\033[0m"
	bold  = "\033[1m"
	dim   = "\033[2m"

	fgBlack = "\033[30m"
	fgWhite = "\033[97m"
	fgGray  = "\033[90m"

	bgGray    = "\033[48;5;240m"
	bgDefault = "\033[49m"

	barGreen   = "\033[38;5;46m"
	barLime    = "\033[38;5;118m"
	barYellow  = "\033[38;5;226m"
	barOrange  = "\033[38;5;208m"
	barRed     = "\033[38;5;196m"
	barDarkRed = "\033[38;5;124m"

	lossGreen   = "\033[38;5;46m"
	lossYellow  = "\033[38;5;226m"
	lossOrange  = "\033[38;5;208m"
	lossRed     = "\033[38;5;196m"
	lossDarkRed = "\033[38;5;124m"

	barEmpty = "\033[38;5;238m"
)

const barWidth = 10

var asnColors = []int{39, 208, 213, 46, 226, 141, 203, 81, 220, 171, 196, 118, 87, 167, 229}

type Config struct {
	Target  string
	DstIP   net.IP
	LocalIP net.IP
	Engine  *trace.Engine
	Numeric bool
}

type Display struct {
	cfg Config

	dnsMu    sync.RWMutex
	dnsCache map[string]string

	asMu    sync.RWMutex
	asCache map[string]*asInfo

	dstHost     string
	dstResolved bool

	showIP    atomic.Bool
	simpleUI  atomic.Bool
	noColor   atomic.Bool
	showUnit  atomic.Bool
	floatRTT  atomic.Bool
	showAllHP atomic.Bool
	showRoute atomic.Bool
	showAS    atomic.Bool
	exportMsg atomic.Value
}

func New(cfg Config) *Display {
	return &Display{
		cfg:      cfg,
		dnsCache: make(map[string]string),
		asCache:  make(map[string]*asInfo),
	}
}

func (d *Display) ToggleIP() {
	d.showIP.Store(!d.showIP.Load())
}

func (d *Display) ToggleSimple() {
	d.simpleUI.Store(!d.simpleUI.Load())
}

func (d *Display) ToggleColor() {
	d.noColor.Store(!d.noColor.Load())
}

func (d *Display) ToggleUnit() {
	d.showUnit.Store(!d.showUnit.Load())
}

func (d *Display) ToggleFloat() {
	d.floatRTT.Store(!d.floatRTT.Load())
}

func (d *Display) ToggleShowAll() {
	d.showAllHP.Store(!d.showAllHP.Load())
}

func (d *Display) ToggleAS() {
	d.showAS.Store(!d.showAS.Load())
}

func (d *Display) ToggleRoute() {
	d.showRoute.Store(true)
}

func (d *Display) ExitRoute() {
	d.showRoute.Store(false)
}

func (d *Display) InRoute() bool {
	return d.showRoute.Load()
}

func (d *Display) Run(ctx context.Context, interval time.Duration) {
	go func() {
		names, _ := net.LookupAddr(d.cfg.DstIP.String())
		d.dnsMu.Lock()
		if len(names) > 0 {
			d.dstHost = strings.TrimSuffix(names[0], ".")
		}
		d.dstResolved = true
		d.dnsMu.Unlock()
	}()

	d.draw()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.draw()
			fmt.Print("\033[?25h")
			fmt.Println()
			return
		case <-ticker.C:
			d.draw()
		}
	}
}

func (d *Display) draw() {
	hops := d.cfg.Engine.Snapshot()
	maxTTL := d.cfg.Engine.MaxTTL()
	if maxTTL == 0 {
		maxTTL = len(hops)
	}

	forceIP := d.showIP.Load()
	noColor := d.noColor.Load()

	for i := 0; i < maxTTL; i++ {
		if hops[i].Addr != nil {
			ip := hops[i].Addr.String()
			if !d.cfg.Numeric && !forceIP {
				d.dnsMu.RLock()
				_, ok := d.dnsCache[ip]
				d.dnsMu.RUnlock()
				if !ok {
					go d.resolve(ip)
				}
			}
			d.asMu.RLock()
			_, ok := d.asCache[ip]
			d.asMu.RUnlock()
			if !ok {
				go d.lookupAS(ip)
			}
		}
	}

	if d.showRoute.Load() {
		d.drawRouteScreen(hops, maxTTL, noColor)
		return
	}

	simple := d.simpleUI.Load()
	units := d.showUnit.Load()
	precise := d.floatRTT.Load()
	showAS := d.showAS.Load()
	termW := term.GetWidth()

	// Calculate RTT column visible width
	rttW := 7
	if precise {
		if units {
			rttW = 8 // "%6.2fms"
		} else {
			rttW = 8 // "%8.2f"
		}
	} else if units {
		rttW = 7 // "%5.0fms"
	}

	// Calculate base width used by fixed columns (excluding host and ASN)
	var baseW int
	if simple {
		// " %-3d " + "  %5s" + "  %5d" + 4x"  %Ns" + "  %5d" = 5 + 7 + 7 + 4*(2+rttW) + 7
		baseW = 5 + 7 + 7 + 4*(2+rttW) + 7
	} else {
		// " %-3d " + 2x"  bar(12)" + 4x"  %Ns" = 5 + 2*14 + 4*(2+rttW)
		baseW = 5 + 2*14 + 4*(2+rttW)
	}

	// Dynamic host width: fill available space, min 16, max 48
	hostW := termW - baseW - 2
	if showAS {
		hostW -= 20 // reserve some space for ASN
	}
	if hostW < 16 {
		hostW = 16
	}
	if hostW > 48 {
		hostW = 48
	}

	var buf strings.Builder

	buf.WriteString("\033[?25l\033[H\033[2J")

	d.drawHeader(&buf, hops, maxTTL, noColor)

	buf.WriteString("\n")

	sepW := baseW + hostW + 2
	if sepW > termW-2 {
		sepW = termW - 2
	}

	if simple {
		hdr := fmt.Sprintf(" %-3s %-*s  %5s  %5s  %*s  %*s  %*s  %*s  %5s",
			"#", hostW, "Host", "Loss%", "Snt", rttW, "Last", rttW, "Avg", rttW, "Best", rttW, "Wrst", "Rcv")
		if noColor {
			fmt.Fprintf(&buf, "%s\n", hdr)
		} else {
			fmt.Fprintf(&buf, " %s%s%s\n", bold, hdr[1:], reset)
		}
	} else {
		hdr := fmt.Sprintf(" %-3s %-*s  %-12s  %-12s  %*s  %*s  %*s  %*s",
			"#", hostW, "Host", "Loss", "Ping", rttW, "Last", rttW, "Avg", rttW, "Best", rttW, "Wrst")
		if noColor {
			fmt.Fprintf(&buf, "%s\n", hdr)
		} else {
			fmt.Fprintf(&buf, " %s%s%s\n", bold, hdr[1:], reset)
		}
	}

	if noColor {
		fmt.Fprintf(&buf, " %s\n", strings.Repeat("-", sepW))
	} else {
		fmt.Fprintf(&buf, " %s%s%s\n", dim, strings.Repeat("\u2500", sepW), reset)
	}

	reached := false
	for i := 0; i < maxTTL; i++ {
		if hops[i].Reached {
			reached = true
			break
		}
	}

	showAll := d.showAllHP.Load()
	lastResponding := 0
	hiddenCount := 0

	if !reached {
		for i := maxTTL - 1; i >= 0; i-- {
			if hops[i].Addr != nil {
				lastResponding = i + 1
				break
			}
		}
		if lastResponding == 0 && hops[0].Sent > 0 {
			lastResponding = 1
		}
		hiddenCount = maxTTL - lastResponding
	}

	displayMax := maxTTL
	if !reached && !showAll && hiddenCount > 0 {
		displayMax = lastResponding
	}

	for i := 0; i < displayMax; i++ {
		h := hops[i]
		if simple {
			d.drawHopSimple(&buf, h, forceIP, noColor, units, precise, showAS, hostW, termW)
		} else {
			d.drawHop(&buf, h, forceIP, noColor, units, precise, showAS, hostW, termW)
		}
	}

	if !reached && hops[0].Sent > 0 && hiddenCount > 0 && !showAll {
		if noColor {
			fmt.Fprintf(&buf, "\n *** %d hops not displayed (no ICMP response) - press x to show\n", hiddenCount)
		} else {
			fmt.Fprintf(&buf, "\n %s*** %d hops not displayed (no ICMP response) - press x to show%s\n", barRed, hiddenCount, reset)
		}
	}

	if msg, ok := d.exportMsg.Load().(string); ok && msg != "" {
		if noColor {
			fmt.Fprintf(&buf, "\n %s\n", msg)
		} else {
			fmt.Fprintf(&buf, "\n %s%s%s%s\n", barGreen, bold, msg, reset)
		}
	}

	os.Stdout.WriteString(buf.String())
}

func (d *Display) drawHeader(buf *strings.Builder, hops []trace.HopStats, maxTTL int, noColor bool) {
	termW := term.GetWidth()

	dstDisplay := d.cfg.DstIP.String()
	d.dnsMu.RLock()
	if d.dstResolved && d.dstHost != "" {
		dstDisplay = d.dstHost
	}
	d.dnsMu.RUnlock()

	lastRTT := "-"
	for i := maxTTL - 1; i >= 0; i-- {
		if hops[i].Reached && hops[i].Received > 0 {
			lastRTT = fmt.Sprintf("%.0fms", hops[i].Last)
			break
		}
	}

	localIP := "?"
	if d.cfg.LocalIP != nil {
		localIP = d.cfg.LocalIP.String()
	}

	header := fmt.Sprintf(" PicoTR  %s (%s)  %s <-> %s  %s ",
		dstDisplay, d.cfg.DstIP, localIP, d.cfg.DstIP, lastRTT)

	padLen := termW - len([]rune(header))
	if padLen < 0 {
		padLen = 0
	}

	if noColor {
		fmt.Fprintf(buf, "%s%s\n", header, strings.Repeat(" ", padLen))
	} else {
		fmt.Fprintf(buf, "%s%s%s%s%s%s\n",
			bgGray, fgWhite, bold, header, strings.Repeat(" ", padLen), reset)
	}

	helpText := "Press q to quit | d DNS/IP | s simple | c colors | u units | p precision | a ASN | r route | t export"
	if noColor {
		fmt.Fprintf(buf, " %s\n", helpText)
	} else {
		fmt.Fprintf(buf, " %s%s%s\n", fgWhite, helpText, reset)
	}
}

func (d *Display) resolveHost(ip string, forceIP bool) string {
	if d.cfg.Numeric || forceIP {
		return ip
	}
	d.dnsMu.RLock()
	name, ok := d.dnsCache[ip]
	d.dnsMu.RUnlock()
	if ok && name != "" {
		return name
	}
	return ip
}

func (d *Display) drawHop(buf *strings.Builder, h trace.HopStats, forceIP, noColor, units, precise, showAS bool, hostW, termW int) {
	host := "???"
	asStr := ""
	if h.Addr != nil {
		host = d.resolveHost(h.Addr.String(), forceIP)
		if showAS {
			asStr = d.formatAS(h.Addr.String(), noColor)
		}
	}

	if len(host) > hostW {
		host = host[:hostW-3] + "..."
	}

	lossPercent := 0.0
	if h.Sent > 0 {
		lossPercent = float64(h.Lost) / float64(h.Sent) * 100
	}

	lossBar := d.buildLossBar(lossPercent, noColor)
	pingBar := d.buildPingBar(h.Avg, noColor)

	last := d.fmtDash(noColor)
	avg := d.fmtDash(noColor)
	best := d.fmtDash(noColor)
	worst := d.fmtDash(noColor)

	if h.Received > 0 {
		last = d.colorizeRTT(h.Last, noColor, units, precise)
		avg = d.colorizeRTT(h.Avg, noColor, units, precise)
		if h.Best < math.MaxFloat64 {
			best = d.colorizeRTT(h.Best, noColor, units, precise)
		}
		worst = d.colorizeRTT(h.Worst, noColor, units, precise)
	}

	// Base line: " TTL HOST  LOSS  PING  LAST  AVG  BEST  WRST"
	baseLineW := 5 + hostW + 2 + 12 + 2 + 12 + 4*(2+rttColW(units, precise))

	if showAS && asStr != "" {
		asVis := d.formatASVisible(h.Addr.String())
		if baseLineW+len(asVis) > termW {
			// ASN doesn't fit on same line — wrap to next line
			fmt.Fprintf(buf, " %-3d %-*s  %s  %s  %s  %s  %s  %s\n",
				h.TTL, hostW, host, lossBar, pingBar, last, avg, best, worst)
			indent := 5
			wrapMaxW := termW - indent - 1
			if wrapMaxW < 10 {
				wrapMaxW = 10
			}
			truncAS := d.formatASTruncated(h.Addr.String(), noColor, wrapMaxW)
			fmt.Fprintf(buf, "%s%s\n", strings.Repeat(" ", indent), truncAS)
		} else {
			fmt.Fprintf(buf, " %-3d %-*s  %s  %s  %s  %s  %s  %s%s\n",
				h.TTL, hostW, host, lossBar, pingBar, last, avg, best, worst, asStr)
		}
	} else {
		fmt.Fprintf(buf, " %-3d %-*s  %s  %s  %s  %s  %s  %s\n",
			h.TTL, hostW, host, lossBar, pingBar, last, avg, best, worst)
	}
}

func (d *Display) drawHopSimple(buf *strings.Builder, h trace.HopStats, forceIP, noColor, units, precise, showAS bool, hostW, termW int) {
	host := "???"
	asStr := ""
	if h.Addr != nil {
		host = d.resolveHost(h.Addr.String(), forceIP)
		if showAS {
			asStr = d.formatAS(h.Addr.String(), noColor)
		}
	}

	if len(host) > hostW {
		host = host[:hostW-3] + "..."
	}

	lossPercent := 0.0
	if h.Sent > 0 {
		lossPercent = float64(h.Lost) / float64(h.Sent) * 100
	}

	var lossStr string
	if noColor {
		lossStr = fmt.Sprintf("%5.1f%%", lossPercent)
	} else {
		lossStr = fmt.Sprintf("%s%5.1f%%%s", lossColor(lossPercent), lossPercent, reset)
	}

	last := d.fmtDash(noColor)
	avg := d.fmtDash(noColor)
	best := d.fmtDash(noColor)
	worst := d.fmtDash(noColor)

	if h.Received > 0 {
		last = d.colorizeRTT(h.Last, noColor, units, precise)
		avg = d.colorizeRTT(h.Avg, noColor, units, precise)
		if h.Best < math.MaxFloat64 {
			best = d.colorizeRTT(h.Best, noColor, units, precise)
		}
		worst = d.colorizeRTT(h.Worst, noColor, units, precise)
	}

	// Base line: " TTL HOST  LOSS%  SNT  LAST  AVG  BEST  WRST  RCV"
	baseLineW := 5 + hostW + 2 + 6 + 2 + 5 + 4*(2+rttColW(units, precise)) + 2 + 5

	if showAS && asStr != "" {
		asVis := d.formatASVisible(h.Addr.String())
		if baseLineW+len(asVis) > termW {
			fmt.Fprintf(buf, " %-3d %-*s  %s  %5d  %s  %s  %s  %s  %5d\n",
				h.TTL, hostW, host, lossStr, h.Sent, last, avg, best, worst, h.Received)
			indent := 5
			wrapMaxW := termW - indent - 1
			if wrapMaxW < 10 {
				wrapMaxW = 10
			}
			truncAS := d.formatASTruncated(h.Addr.String(), noColor, wrapMaxW)
			fmt.Fprintf(buf, "%s%s\n", strings.Repeat(" ", indent), truncAS)
		} else {
			fmt.Fprintf(buf, " %-3d %-*s  %s  %5d  %s  %s  %s  %s  %5d%s\n",
				h.TTL, hostW, host, lossStr, h.Sent, last, avg, best, worst, h.Received, asStr)
		}
	} else {
		fmt.Fprintf(buf, " %-3d %-*s  %s  %5d  %s  %s  %s  %s  %5d\n",
			h.TTL, hostW, host, lossStr, h.Sent, last, avg, best, worst, h.Received)
	}
}

func rttColW(units, precise bool) int {
	if precise {
		if units {
			return 8
		}
		return 8
	}
	if units {
		return 7
	}
	return 7
}

func (d *Display) fmtDash(noColor bool) string {
	if noColor {
		return "      -"
	}
	return fmt.Sprintf("%s      -%s", dim, reset)
}

func pingFilled(avgMs float64) int {
	switch {
	case avgMs <= 0:
		return 0
	case avgMs < 5:
		return 1
	case avgMs < 10:
		return 2
	case avgMs < 20:
		return 3
	case avgMs < 40:
		return 4
	case avgMs < 60:
		return 5
	case avgMs < 100:
		return 6
	case avgMs < 150:
		return 7
	case avgMs < 200:
		return 8
	case avgMs < 300:
		return 9
	default:
		return 10
	}
}

var pingGradient = []string{barGreen, barGreen, barLime, barYellow, barYellow, barOrange, barOrange, barRed, barRed, barDarkRed}
var lossGradient = []string{lossGreen, lossGreen, lossYellow, lossYellow, lossOrange, lossOrange, lossRed, lossRed, lossDarkRed, lossDarkRed}

func (d *Display) buildPingBar(avgMs float64, noColor bool) string {
	filled := pingFilled(avgMs)

	if noColor {
		var b strings.Builder
		b.WriteByte('[')
		b.WriteString(strings.Repeat("|", filled))
		b.WriteString(strings.Repeat(".", barWidth-filled))
		b.WriteByte(']')
		b.WriteString("  ")
		return b.String()
	}

	if filled == 0 {
		return barEmpty + strings.Repeat("\u2502", barWidth) + reset + "  "
	}

	var b strings.Builder
	for i := 0; i < filled; i++ {
		b.WriteString(pingGradient[i])
		b.WriteString("\u2588")
	}
	if filled < barWidth {
		b.WriteString(barEmpty)
		b.WriteString(strings.Repeat("\u2502", barWidth-filled))
	}
	b.WriteString(reset)
	b.WriteString("  ")
	return b.String()
}

func (d *Display) buildLossBar(lossPct float64, noColor bool) string {
	filled := 0
	if lossPct > 0 {
		filled = int(lossPct / 10)
		if filled < 1 {
			filled = 1
		}
		if filled > barWidth {
			filled = barWidth
		}
	}

	if noColor {
		var b strings.Builder
		b.WriteByte('[')
		if filled == 0 {
			b.WriteString(strings.Repeat(".", barWidth))
		} else {
			b.WriteString(strings.Repeat("|", filled))
			b.WriteString(strings.Repeat(".", barWidth-filled))
		}
		b.WriteByte(']')
		b.WriteString("  ")
		return b.String()
	}

	if lossPct == 0 {
		return lossGreen + "\u2588" + barEmpty + strings.Repeat("\u2502", barWidth-1) + reset + "  "
	}

	var b strings.Builder
	for i := 0; i < filled; i++ {
		b.WriteString(lossGradient[i])
		b.WriteString("\u2588")
	}
	if filled < barWidth {
		b.WriteString(barEmpty)
		b.WriteString(strings.Repeat("\u2502", barWidth-filled))
	}
	b.WriteString(reset)
	b.WriteString("  ")
	return b.String()
}

func (d *Display) colorizeRTT(ms float64, noColor, units, precise bool) string {
	var val string
	if precise {
		if units {
			val = fmt.Sprintf("%6.2fms", ms)
		} else {
			val = fmt.Sprintf("%8.2f", ms)
		}
	} else {
		if units {
			val = fmt.Sprintf("%5.0fms", ms)
		} else {
			val = fmt.Sprintf("%7.0f", ms)
		}
	}
	if noColor {
		return val
	}
	return pingColor(ms) + val + reset
}

func pingColor(ms float64) string {
	switch {
	case ms < 20:
		return barGreen
	case ms < 50:
		return barLime
	case ms < 100:
		return barYellow
	case ms < 150:
		return barOrange
	case ms < 300:
		return barRed
	default:
		return barDarkRed
	}
}

func lossColor(pct float64) string {
	switch {
	case pct == 0:
		return lossGreen
	case pct < 10:
		return lossYellow
	case pct < 25:
		return lossOrange
	case pct < 50:
		return lossRed
	default:
		return lossDarkRed
	}
}

func (d *Display) resolve(ip string) {
	names, err := net.LookupAddr(ip)
	d.dnsMu.Lock()
	defer d.dnsMu.Unlock()
	if err != nil || len(names) == 0 {
		d.dnsCache[ip] = ip
	} else {
		d.dnsCache[ip] = strings.TrimSuffix(names[0], ".")
	}
}

func (d *Display) lookupAS(ip string) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		d.asMu.Lock()
		d.asCache[ip] = &asInfo{}
		d.asMu.Unlock()
		return
	}

	p := parsed.To4()
	if p == nil {
		d.asMu.Lock()
		d.asCache[ip] = &asInfo{}
		d.asMu.Unlock()
		return
	}

	origin := fmt.Sprintf("%d.%d.%d.%d.origin.asn.cymru.com", p[3], p[2], p[1], p[0])
	txts, err := net.LookupTXT(origin)
	if err != nil || len(txts) == 0 {
		d.asMu.Lock()
		d.asCache[ip] = &asInfo{}
		d.asMu.Unlock()
		return
	}

	fields := strings.Split(txts[0], "|")
	if len(fields) < 1 {
		d.asMu.Lock()
		d.asCache[ip] = &asInfo{}
		d.asMu.Unlock()
		return
	}

	asnStr := strings.TrimSpace(fields[0])
	asn, _ := strconv.ParseUint(asnStr, 10, 32)

	asnQuery := fmt.Sprintf("AS%s.asn.cymru.com", asnStr)
	nameTxts, err := net.LookupTXT(asnQuery)
	asName := ""
	if err == nil && len(nameTxts) > 0 {
		nameFields := strings.Split(nameTxts[0], "|")
		if len(nameFields) >= 5 {
			asName = strings.TrimSpace(nameFields[4])
		}
	}

	d.asMu.Lock()
	d.asCache[ip] = &asInfo{ASN: uint32(asn), ASName: asName}
	d.asMu.Unlock()
}

func (d *Display) getASInfo(ip string) *asInfo {
	d.asMu.RLock()
	info, ok := d.asCache[ip]
	d.asMu.RUnlock()
	if ok {
		return info
	}
	return nil
}

func (d *Display) formatAS(ip string, noColor bool) string {
	info := d.getASInfo(ip)
	if info == nil || info.ASN == 0 {
		return ""
	}
	label := fmt.Sprintf("AS%d", info.ASN)
	if info.ASName != "" {
		label += " " + info.ASName
	}
	if noColor {
		return fmt.Sprintf(" [%s]", label)
	}
	return fmt.Sprintf(" %s[%s]%s", fgGray, label, reset)
}

func (d *Display) formatASVisible(ip string) string {
	info := d.getASInfo(ip)
	if info == nil || info.ASN == 0 {
		return ""
	}
	label := fmt.Sprintf("AS%d", info.ASN)
	if info.ASName != "" {
		label += " " + info.ASName
	}
	return fmt.Sprintf(" [%s]", label)
}

func (d *Display) formatASTruncated(ip string, noColor bool, maxW int) string {
	info := d.getASInfo(ip)
	if info == nil || info.ASN == 0 {
		return ""
	}
	label := fmt.Sprintf("AS%d", info.ASN)
	if info.ASName != "" {
		label += " " + info.ASName
	}
	vis := fmt.Sprintf("[%s]", label)
	if len(vis) > maxW && maxW > 6 {
		vis = vis[:maxW-3] + "..."
	}
	if noColor {
		return vis
	}
	return fmt.Sprintf("%s%s%s", fgGray, vis, reset)
}

func asnColorIdx(asn uint32) int {
	return int(asn % uint32(len(asnColors)))
}

func asnBgFg(asn uint32) (string, string) {
	c := asnColors[asnColorIdx(asn)]
	bg := fmt.Sprintf("\033[48;5;%dm", c)
	if c == 226 || c == 229 || c == 220 || c == 118 || c == 46 {
		return bg, fgBlack
	}
	return bg, fgWhite
}


type routeSegment struct {
	asn   uint32
	name  string
	ips   []string
	hosts []string
	hops  []int
	isGap bool
}

func (d *Display) buildSegments(hops []trace.HopStats, displayMax int) []routeSegment {
	var segments []routeSegment
	var current *routeSegment

	for i := 0; i < displayMax; i++ {
		h := hops[i]
		if h.Addr == nil {
			if current != nil && current.isGap {
				current.hops = append(current.hops, h.TTL)
			} else {
				seg := routeSegment{isGap: true, hops: []int{h.TTL}}
				segments = append(segments, seg)
				current = &segments[len(segments)-1]
			}
			continue
		}

		ip := h.Addr.String()
		info := d.getASInfo(ip)
		var asn uint32
		var name string
		if info != nil && info.ASN != 0 {
			asn = info.ASN
			name = info.ASName
		}

		if current != nil && !current.isGap && current.asn == asn && asn != 0 {
			current.hops = append(current.hops, h.TTL)
			current.ips = append(current.ips, ip)
		} else {
			seg := routeSegment{asn: asn, name: name, ips: []string{ip}, hops: []int{h.TTL}}
			segments = append(segments, seg)
			current = &segments[len(segments)-1]
		}
	}
	return segments
}

func (d *Display) buildAllHopSegments(hops []trace.HopStats, displayMax int) []routeSegment {
	var segments []routeSegment
	for i := 0; i < displayMax; i++ {
		h := hops[i]
		if h.Addr == nil {
			segments = append(segments, routeSegment{isGap: true, hops: []int{h.TTL}})
			continue
		}
		ip := h.Addr.String()
		info := d.getASInfo(ip)
		var asn uint32
		var asName string
		if info != nil && info.ASN != 0 {
			asn = info.ASN
			asName = info.ASName
		}
		host := d.resolveHost(ip, false)
		seg := routeSegment{asn: asn, name: asName, ips: []string{ip}, hops: []int{h.TTL}}
		if host != ip {
			seg.hosts = []string{host}
		}
		segments = append(segments, seg)
	}
	return segments
}

func (d *Display) routeDisplayMax(hops []trace.HopStats, maxTTL int) int {
	reached := false
	for i := 0; i < maxTTL; i++ {
		if hops[i].Reached {
			reached = true
			break
		}
	}
	if reached {
		return maxTTL
	}
	lastResp := 0
	for i := maxTTL - 1; i >= 0; i-- {
		if hops[i].Addr != nil {
			lastResp = i + 1
			break
		}
	}
	if lastResp > 0 {
		return lastResp
	}
	return maxTTL
}

func (d *Display) routeLocalIP() string {
	if d.cfg.LocalIP != nil {
		return d.cfg.LocalIP.String()
	}
	return "?"
}

func (d *Display) routeDstDisplay() string {
	dstDisplay := d.cfg.DstIP.String()
	d.dnsMu.RLock()
	if d.dstResolved && d.dstHost != "" {
		dstDisplay = d.dstHost
	}
	d.dnsMu.RUnlock()
	return dstDisplay
}

func (d *Display) drawRouteScreen(hops []trace.HopStats, maxTTL int, noColor bool) {
	displayMax := d.routeDisplayMax(hops, maxTTL)
	segments := d.buildSegments(hops, displayMax)

	localIP := d.routeLocalIP()
	dstDisplay := d.routeDstDisplay()

	termW := term.GetWidth()

	var buf strings.Builder
	buf.WriteString("\033[?25l\033[H\033[2J")

	headerText := " PicoTR - Route View "
	padLen := termW - len(headerText)
	if padLen < 0 {
		padLen = 0
	}
	if noColor {
		fmt.Fprintf(&buf, "%s%s\n", headerText, strings.Repeat(" ", padLen))
	} else {
		fmt.Fprintf(&buf, "%s%s%s%s%s%s\n",
			bgGray, fgWhite, bold, headerText, strings.Repeat(" ", padLen), reset)
	}

	if noColor {
		buf.WriteString(" Press Backspace to return | e export PNG | a export all hops\n")
	} else {
		fmt.Fprintf(&buf, " %sPress Backspace to return | e export PNG | a export all hops%s\n", fgGray, reset)
	}

	if msg, ok := d.exportMsg.Load().(string); ok && msg != "" {
		if noColor {
			fmt.Fprintf(&buf, " %s\n", msg)
		} else {
			fmt.Fprintf(&buf, " %s%s%s%s\n", barGreen, bold, msg, reset)
		}
	}
	buf.WriteString("\n")

	type pill struct {
		label   string
		visLen  int
		colored string
	}

	var pills []pill

	makePill := func(label string, colorType int, asn uint32) pill {
		vl := len([]rune(label)) + 2
		if noColor {
			return pill{label: "[" + label + "]", visLen: vl, colored: "[" + label + "]"}
		}
		bg := ""
		fg := ""
		switch {
		case colorType == -1 || colorType == -3:
			bg = "\033[48;5;24m"
			fg = fgWhite
		case colorType == -2:
			bg = "\033[48;5;236m"
			fg = fgGray
		case colorType == -4:
			bg = "\033[48;5;236m"
			fg = fgWhite
		default:
			bg, fg = asnBgFg(asn)
		}
		return pill{
			label:   label,
			visLen:  vl,
			colored: bg + fg + bold + " " + label + " " + reset,
		}
	}

	pills = append(pills, makePill(localIP, -1, 0))

	for _, seg := range segments {
		if seg.isGap {
			pills = append(pills, makePill("???", -2, 0))
		} else if seg.asn != 0 {
			label := fmt.Sprintf("AS%d", seg.asn)
			pills = append(pills, makePill(label, 0, seg.asn))
		} else {
			ip := ""
			if len(seg.ips) > 0 {
				ip = seg.ips[0]
			}
			pills = append(pills, makePill(ip, -4, 0))
		}
	}

	pills = append(pills, makePill(dstDisplay, -3, 0))

	arrow := " \u2500\u25B6 "
	arrowPlain := " -> "
	arrowVis := 4

	indent := 2
	lineW := indent
	first := true

	buf.WriteString(strings.Repeat(" ", indent))

	for _, p := range pills {
		needed := p.visLen
		if !first {
			needed += arrowVis
		}

		if !first && lineW+needed > termW {
			buf.WriteString("\n")
			buf.WriteString(strings.Repeat(" ", indent))
			lineW = indent
			first = true
		}

		if !first {
			if noColor {
				buf.WriteString(arrowPlain)
			} else {
				fmt.Fprintf(&buf, "%s%s%s", dim, arrow, reset)
			}
			lineW += arrowVis
		}

		buf.WriteString(p.colored)
		lineW += p.visLen
		first = false
	}
	buf.WriteString("\n")

	buf.WriteString("\n")
	sepW := termW - 2
	if sepW > 80 {
		sepW = 80
	}
	if noColor {
		buf.WriteString(" AS Path:\n")
		fmt.Fprintf(&buf, " %s\n", strings.Repeat("-", sepW))
	} else {
		fmt.Fprintf(&buf, " %sAS Path:%s\n", bold, reset)
		fmt.Fprintf(&buf, " %s%s%s\n", dim, strings.Repeat("\u2500", sepW), reset)
	}

	seen := make(map[uint32]bool)
	for _, seg := range segments {
		if seg.isGap || seg.asn == 0 {
			continue
		}
		if seen[seg.asn] {
			continue
		}
		seen[seg.asn] = true

		hopRange := hopRangeStr(seg.hops)
		asnLabel := fmt.Sprintf("AS%d", seg.asn)

		if noColor {
			fmt.Fprintf(&buf, "   %-12s %-35s %s\n", asnLabel, seg.name, hopRange)
		} else {
			bg, fg := asnBgFg(seg.asn)
			fmt.Fprintf(&buf, "   %s%s%s %s %s  %-35s %s%s%s\n",
				bg, fg, bold, asnLabel, reset, seg.name, dim, hopRange, reset)
		}
	}

	os.Stdout.WriteString(buf.String())
}


func (d *Display) ExportPNG() {
	hops := d.cfg.Engine.Snapshot()
	maxTTL := d.cfg.Engine.MaxTTL()
	if maxTTL == 0 {
		maxTTL = len(hops)
	}

	displayMax := d.routeDisplayMax(hops, maxTTL)
	segments := d.buildSegments(hops, displayMax)
	localIP := d.routeLocalIP()
	dstDisplay := d.routeDstDisplay()

	filename := fmt.Sprintf("picotr_%s.png", strings.ReplaceAll(d.cfg.Target, ".", "_"))
	err := ExportRoutePNG(filename, localIP, dstDisplay, d.cfg.DstIP.String(), segments, d.noColor.Load())
	if err != nil {
		d.exportMsg.Store(fmt.Sprintf("Export failed: %v", err))
	} else {
		d.exportMsg.Store(fmt.Sprintf("Exported to %s", filename))
	}
	go func() {
		time.Sleep(3 * time.Second)
		d.exportMsg.Store("")
	}()
}

func (d *Display) ExportAllHopsPNG() {
	hops := d.cfg.Engine.Snapshot()
	maxTTL := d.cfg.Engine.MaxTTL()
	if maxTTL == 0 {
		maxTTL = len(hops)
	}

	displayMax := d.routeDisplayMax(hops, maxTTL)
	segments := d.buildAllHopSegments(hops, displayMax)
	localIP := d.routeLocalIP()
	dstDisplay := d.routeDstDisplay()

	filename := fmt.Sprintf("picotr_allhops_%s.png", strings.ReplaceAll(d.cfg.Target, ".", "_"))
	err := ExportAllHopsGroupedPNG(filename, localIP, dstDisplay, d.cfg.DstIP.String(), segments, d.noColor.Load())
	if err != nil {
		d.exportMsg.Store(fmt.Sprintf("Export failed: %v", err))
	} else {
		d.exportMsg.Store(fmt.Sprintf("Exported to %s", filename))
	}
	go func() {
		time.Sleep(3 * time.Second)
		d.exportMsg.Store("")
	}()
}

func (d *Display) ExportTrace() {
	hops := d.cfg.Engine.Snapshot()
	maxTTL := d.cfg.Engine.MaxTTL()
	if maxTTL == 0 {
		maxTTL = len(hops)
	}

	resolveHost := func(ip string) string {
		return d.resolveHost(ip, false)
	}
	getAS := func(ip string) *asInfo {
		return d.getASInfo(ip)
	}

	rows := buildTraceRows(hops, maxTTL, resolveHost, getAS)

	dstDisplay := d.routeDstDisplay()
	headerInfo := fmt.Sprintf("%s (%s) <-> %s", dstDisplay, d.cfg.DstIP, d.routeLocalIP())

	opts := traceExportOpts{
		showIP:  d.showIP.Load(),
		precise: d.floatRTT.Load(),
		units:   d.showUnit.Load(),
		showAS:  d.showAS.Load(),
		noColor: d.noColor.Load(),
	}

	filename := fmt.Sprintf("picotr_trace_%s.png", strings.ReplaceAll(d.cfg.Target, ".", "_"))
	err := ExportTracePNG(filename, headerInfo, rows, opts)
	if err != nil {
		d.exportMsg.Store(fmt.Sprintf("Export failed: %v", err))
	} else {
		d.exportMsg.Store(fmt.Sprintf("Exported to %s", filename))
	}
	go func() {
		time.Sleep(3 * time.Second)
		d.exportMsg.Store("")
	}()
}

func hopRangeStr(hops []int) string {
	if len(hops) == 0 {
		return ""
	}
	if len(hops) == 1 {
		return fmt.Sprintf("hop %d", hops[0])
	}
	return fmt.Sprintf("hops %d-%d", hops[0], hops[len(hops)-1])
}

