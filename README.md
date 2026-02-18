<p align="center">
  <img width="150" src="https://sc.vex.systems/branding/vex_n.png" />
</p>

# PicoTR

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.25%2B-blue.svg)](https://golang.org/)
[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/VEX-Systems/picotr/)

A fast, lightweight traceroute tool written in Go with a rich terminal UI, real-time statistics, ASN lookup, and PNG export.

## Features

- **Real-time tracing**: Continuous ICMP probing with live-updating terminal display.
- **Rich TUI**: Color-coded RTT bars, packet loss indicators, and dynamic layout adapting to terminal width.
- **ASN lookup**: Resolves Autonomous System Numbers via Cymru DNS for each hop.
- **Route view**: Visual AS path map showing how traffic traverses the internet.
- **PNG export**: Export the current trace table, route map, or all-hops breakdown as a PNG image.
- **DNS resolution**: Reverse DNS lookups with caching, toggleable on the fly.
- **Flexible display**: Toggle between simple/detailed mode, units, float precision, color, and more — all without restarting.

## Installation

```bash
go install github.com/VEX-Systems/picotr/cmd/picotr@latest
```

Or build from source:

```bash
git clone https://github.com/VEX-Systems/picotr
cd picotr
go build -o picotr ./cmd/picotr
```

> **Note**: Sending raw ICMP packets requires elevated privileges on most systems. Run with `sudo` on Linux/macOS.

## Usage

```bash
picotr [options] <target>
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-m` | `30` | Maximum number of hops |
| `-w` | `3s` | Timeout for each probe |
| `-i` | `1s` | Interval between probe rounds |
| `-n` | `false` | Do not resolve hostnames (numeric mode) |
| `-version` | — | Print version and exit |

### Example

```bash
picotr google.com
picotr -m 20 -w 2s -n 8.8.8.8
```

## Keybindings

| Key | Action |
|-----|--------|
| `q` / `Q` / `Ctrl+C` | Quit |
| `d` | Toggle DNS / IP display |
| `s` | Toggle simple / detailed mode |
| `c` | Toggle color |
| `u` | Toggle ms units |
| `p` | Toggle float precision |
| `x` | Show / hide unresponsive hops |
| `a` | Toggle ASN display (main view) |
| `r` | Enter route view |
| `Backspace` | Exit route view |
| `e` | Export PNG (trace or route) |
| `t` | Export trace table as PNG |
| `a` *(route view)* | Export all-hops PNG |

## Export

PicoTR can export the current state to PNG files without any external dependencies — the renderer is built in.

| Command | Output file |
|---------|-------------|
| `e` — trace table | `picotr_trace_<target>.png` |
| `e` *(route view)* — route map | `picotr_<target>.png` |
| `a` *(route view)* — all hops | `picotr_allhops_<target>.png` |

## Project Structure

```
cmd/picotr/       # Entry point, CLI flags, keyboard handling
internal/
  probe/          # ICMP prober abstraction (platform-specific implementations)
  trace/          # Tracing engine — concurrent probing, hop statistics
  output/         # Terminal display, ASN lookup, PNG export
  term/           # Terminal raw mode and width detection
```

## Requirements

- Go 1.25+
- `golang.org/x/net` (ICMP)
- `golang.org/x/sys` (platform syscalls)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.
