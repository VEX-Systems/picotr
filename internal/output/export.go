package output

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/VEX-Systems/picotr/internal/trace"
)

var font5x7 = map[rune][]uint8{
	'A': {0x7C, 0x12, 0x11, 0x12, 0x7C},
	'B': {0x7F, 0x49, 0x49, 0x49, 0x36},
	'C': {0x3E, 0x41, 0x41, 0x41, 0x22},
	'D': {0x7F, 0x41, 0x41, 0x22, 0x1C},
	'E': {0x7F, 0x49, 0x49, 0x49, 0x41},
	'F': {0x7F, 0x09, 0x09, 0x09, 0x01},
	'G': {0x3E, 0x41, 0x49, 0x49, 0x7A},
	'H': {0x7F, 0x08, 0x08, 0x08, 0x7F},
	'I': {0x00, 0x41, 0x7F, 0x41, 0x00},
	'J': {0x20, 0x40, 0x41, 0x3F, 0x01},
	'K': {0x7F, 0x08, 0x14, 0x22, 0x41},
	'L': {0x7F, 0x40, 0x40, 0x40, 0x40},
	'M': {0x7F, 0x02, 0x0C, 0x02, 0x7F},
	'N': {0x7F, 0x04, 0x08, 0x10, 0x7F},
	'O': {0x3E, 0x41, 0x41, 0x41, 0x3E},
	'P': {0x7F, 0x09, 0x09, 0x09, 0x06},
	'Q': {0x3E, 0x41, 0x51, 0x21, 0x5E},
	'R': {0x7F, 0x09, 0x19, 0x29, 0x46},
	'S': {0x46, 0x49, 0x49, 0x49, 0x31},
	'T': {0x01, 0x01, 0x7F, 0x01, 0x01},
	'U': {0x3F, 0x40, 0x40, 0x40, 0x3F},
	'V': {0x1F, 0x20, 0x40, 0x20, 0x1F},
	'W': {0x3F, 0x40, 0x30, 0x40, 0x3F},
	'X': {0x63, 0x14, 0x08, 0x14, 0x63},
	'Y': {0x07, 0x08, 0x70, 0x08, 0x07},
	'Z': {0x61, 0x51, 0x49, 0x45, 0x43},
	'a': {0x20, 0x54, 0x54, 0x54, 0x78},
	'b': {0x7F, 0x48, 0x44, 0x44, 0x38},
	'c': {0x38, 0x44, 0x44, 0x44, 0x20},
	'd': {0x38, 0x44, 0x44, 0x48, 0x7F},
	'e': {0x38, 0x54, 0x54, 0x54, 0x18},
	'f': {0x08, 0x7E, 0x09, 0x01, 0x02},
	'g': {0x0C, 0x52, 0x52, 0x52, 0x3E},
	'h': {0x7F, 0x08, 0x04, 0x04, 0x78},
	'i': {0x00, 0x44, 0x7D, 0x40, 0x00},
	'j': {0x20, 0x40, 0x44, 0x3D, 0x00},
	'k': {0x7F, 0x10, 0x28, 0x44, 0x00},
	'l': {0x00, 0x41, 0x7F, 0x40, 0x00},
	'm': {0x7C, 0x04, 0x18, 0x04, 0x78},
	'n': {0x7C, 0x08, 0x04, 0x04, 0x78},
	'o': {0x38, 0x44, 0x44, 0x44, 0x38},
	'p': {0x7C, 0x14, 0x14, 0x14, 0x08},
	'q': {0x08, 0x14, 0x14, 0x18, 0x7C},
	'r': {0x7C, 0x08, 0x04, 0x04, 0x08},
	's': {0x48, 0x54, 0x54, 0x54, 0x20},
	't': {0x04, 0x3F, 0x44, 0x40, 0x20},
	'u': {0x3C, 0x40, 0x40, 0x20, 0x7C},
	'v': {0x1C, 0x20, 0x40, 0x20, 0x1C},
	'w': {0x3C, 0x40, 0x30, 0x40, 0x3C},
	'x': {0x44, 0x28, 0x10, 0x28, 0x44},
	'y': {0x0C, 0x50, 0x50, 0x50, 0x3C},
	'z': {0x44, 0x64, 0x54, 0x4C, 0x44},
	'0': {0x3E, 0x51, 0x49, 0x45, 0x3E},
	'1': {0x00, 0x42, 0x7F, 0x40, 0x00},
	'2': {0x42, 0x61, 0x51, 0x49, 0x46},
	'3': {0x21, 0x41, 0x45, 0x4B, 0x31},
	'4': {0x18, 0x14, 0x12, 0x7F, 0x10},
	'5': {0x27, 0x45, 0x45, 0x45, 0x39},
	'6': {0x3C, 0x4A, 0x49, 0x49, 0x30},
	'7': {0x01, 0x71, 0x09, 0x05, 0x03},
	'8': {0x36, 0x49, 0x49, 0x49, 0x36},
	'9': {0x06, 0x49, 0x49, 0x29, 0x1E},
	'.': {0x00, 0x60, 0x60, 0x00, 0x00},
	',': {0x00, 0x80, 0x60, 0x00, 0x00},
	':': {0x00, 0x36, 0x36, 0x00, 0x00},
	'-': {0x08, 0x08, 0x08, 0x08, 0x08},
	'_': {0x40, 0x40, 0x40, 0x40, 0x40},
	' ': {0x00, 0x00, 0x00, 0x00, 0x00},
	'(': {0x00, 0x1C, 0x22, 0x41, 0x00},
	')': {0x00, 0x41, 0x22, 0x1C, 0x00},
	'[': {0x00, 0x7F, 0x41, 0x41, 0x00},
	']': {0x00, 0x41, 0x41, 0x7F, 0x00},
	'/': {0x20, 0x10, 0x08, 0x04, 0x02},
	'+': {0x08, 0x08, 0x3E, 0x08, 0x08},
	'?': {0x02, 0x01, 0x51, 0x09, 0x06},
	'!': {0x00, 0x00, 0x5F, 0x00, 0x00},
	'%': {0x23, 0x13, 0x08, 0x64, 0x62},
	'<': {0x08, 0x14, 0x22, 0x41, 0x00},
	'>': {0x00, 0x41, 0x22, 0x14, 0x08},
	'=': {0x14, 0x14, 0x14, 0x14, 0x14},
	'@': {0x3E, 0x41, 0x5D, 0x55, 0x1E},
	'#': {0x14, 0x7F, 0x14, 0x7F, 0x14},
	'&': {0x36, 0x49, 0x55, 0x22, 0x50},
	'*': {0x14, 0x08, 0x3E, 0x08, 0x14},
	'^': {0x02, 0x01, 0x02, 0x00, 0x00},
	'~': {0x04, 0x02, 0x04, 0x08, 0x04},
}

func drawChar(img *image.RGBA, x, y int, ch rune, fg color.RGBA, scale int) int {
	glyph, ok := font5x7[ch]
	if !ok {
		glyph = font5x7['?']
	}

	for col := 0; col < 5; col++ {
		bits := glyph[col]
		for row := 0; row < 7; row++ {
			if bits&(1<<uint(row)) != 0 {
				for sy := 0; sy < scale; sy++ {
					for sx := 0; sx < scale; sx++ {
						px := x + col*scale + sx
						py := y + row*scale + sy
						if px >= 0 && py >= 0 && px < img.Bounds().Max.X && py < img.Bounds().Max.Y {
							img.SetRGBA(px, py, fg)
						}
					}
				}
			}
		}
	}
	return 6 * scale
}

func drawText(img *image.RGBA, x, y int, text string, fg color.RGBA, scale int) int {
	cx := x
	for _, ch := range text {
		cx += drawChar(img, cx, y, ch, fg, scale)
	}
	return cx - x
}

func textWidth(text string, scale int) int {
	return len([]rune(text)) * 6 * scale
}

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			if px >= 0 && py >= 0 && px < img.Bounds().Max.X && py < img.Bounds().Max.Y {
				img.SetRGBA(px, py, c)
			}
		}
	}
}

func drawStrokeRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for px := x; px < x+w; px++ {
		img.SetRGBA(px, y, c)
		img.SetRGBA(px, y+h-1, c)
	}
	for py := y; py < y+h; py++ {
		img.SetRGBA(x, py, c)
		img.SetRGBA(x+w-1, py, c)
	}
}

func drawRoundedRect(img *image.RGBA, x, y, w, h, r int, fill, border color.RGBA) {
	fillRect(img, x+r, y, w-2*r, h, fill)
	fillRect(img, x, y+r, w, h-2*r, fill)

	for cy := 0; cy <= r; cy++ {
		for cx := 0; cx <= r; cx++ {
			if cx*cx+cy*cy <= r*r {
				img.SetRGBA(x+r-cx, y+r-cy, fill)
				img.SetRGBA(x+w-1-r+cx, y+r-cy, fill)
				img.SetRGBA(x+r-cx, y+h-1-r+cy, fill)
				img.SetRGBA(x+w-1-r+cx, y+h-1-r+cy, fill)
			}
		}
	}

	for px := x + r; px < x+w-r; px++ {
		for t := 0; t < 2; t++ {
			img.SetRGBA(px, y+t, border)
			img.SetRGBA(px, y+h-1-t, border)
		}
	}
	for py := y + r; py < y+h-r; py++ {
		for t := 0; t < 2; t++ {
			img.SetRGBA(x+t, py, border)
			img.SetRGBA(x+w-1-t, py, border)
		}
	}
}

func drawArrowRight(img *image.RGBA, x1, x2, cy int, c color.RGBA) {
	tipSize := 7
	shaftEnd := x2 - tipSize
	if shaftEnd < x1 {
		shaftEnd = x1
	}
	for px := x1; px < shaftEnd; px++ {
		img.SetRGBA(px, cy-1, c)
		img.SetRGBA(px, cy, c)
		img.SetRGBA(px, cy+1, c)
	}
	for i := 0; i <= tipSize; i++ {
		span := tipSize - i
		for dy := -span; dy <= span; dy++ {
			px := shaftEnd + i
			if px < img.Bounds().Max.X {
				img.SetRGBA(px, cy+dy, c)
			}
		}
	}
}

func drawArrowDown(img *image.RGBA, cx, y1, y2 int, c color.RGBA) {
	tipSize := 7
	shaftEnd := y2 - tipSize
	if shaftEnd < y1 {
		shaftEnd = y1
	}
	for py := y1; py < shaftEnd; py++ {
		img.SetRGBA(cx-1, py, c)
		img.SetRGBA(cx, py, c)
		img.SetRGBA(cx+1, py, c)
	}
	for i := 0; i <= tipSize; i++ {
		span := tipSize - i
		for dx := -span; dx <= span; dx++ {
			py := shaftEnd + i
			if py < img.Bounds().Max.Y {
				img.SetRGBA(cx+dx, py, c)
			}
		}
	}
}

var pngASNColors = [][3]uint8{
	{0, 172, 230},
	{255, 152, 0},
	{233, 30, 99},
	{76, 175, 80},
	{255, 235, 59},
	{156, 39, 176},
	{244, 67, 54},
	{0, 188, 212},
	{255, 193, 7},
	{171, 71, 188},
	{211, 47, 47},
	{139, 195, 74},
	{38, 166, 154},
	{229, 115, 115},
	{255, 241, 118},
}

func pngASNColor(asn uint32) color.RGBA {
	if asn == 0 {
		return color.RGBA{80, 80, 80, 255}
	}
	c := pngASNColors[asn%uint32(len(pngASNColors))]
	return color.RGBA{c[0], c[1], c[2], 255}
}

func pngASNFg(bg color.RGBA) color.RGBA {
	lum := int(bg.R)*299 + int(bg.G)*587 + int(bg.B)*114
	if lum > 128000 {
		return color.RGBA{0, 0, 0, 255}
	}
	return color.RGBA{255, 255, 255, 255}
}

func ExportRoutePNG(filename string, localIP, dstDisplay, dstIP string, segments []routeSegment, noColor bool) error {
	scale := 2
	padding := 40
	arrowW := 40
	textSc := scale
	lineH := 7*textSc + 4
	boxPadX := 20
	boxPadY := 10
	rowGap := 50
	maxImgW := 1800
	bw := noColor

	type nodeInfo struct {
		line1     string
		line2     string
		line3     string
		fillColor color.RGBA
		fgColor   color.RGBA
		borderC   color.RGBA
		boxW      int
		boxH      int
	}

	var nodes []nodeInfo

	var srcFill, dstFill, gapFill, unknownFill, white, grayFg color.RGBA
	if bw {
		srcFill = color.RGBA{255, 255, 255, 255}
		dstFill = color.RGBA{255, 255, 255, 255}
		gapFill = color.RGBA{255, 255, 255, 255}
		unknownFill = color.RGBA{255, 255, 255, 255}
		white = color.RGBA{0, 0, 0, 255}
		grayFg = color.RGBA{100, 100, 100, 255}
	} else {
		srcFill = color.RGBA{30, 90, 150, 255}
		dstFill = color.RGBA{30, 90, 150, 255}
		gapFill = color.RGBA{60, 60, 60, 255}
		unknownFill = color.RGBA{80, 80, 80, 255}
		white = color.RGBA{255, 255, 255, 255}
		grayFg = color.RGBA{160, 160, 160, 255}
	}

	nodes = append(nodes, nodeInfo{
		line1: localIP, line2: "Source",
		fillColor: srcFill, fgColor: white,
	})

	for _, seg := range segments {
		if seg.isGap {
			nodes = append(nodes, nodeInfo{
				line1: "???", line2: hopRangeStr(seg.hops),
				fillColor: gapFill, fgColor: grayFg,
			})
			continue
		}
		if seg.asn != 0 {
			l1 := fmt.Sprintf("AS%d", seg.asn)
			l2 := seg.name
			if len(l2) > 40 {
				l2 = l2[:37] + "..."
			}
			l3 := hopRangeStr(seg.hops)
			if bw {
				nodes = append(nodes, nodeInfo{
					line1: l1, line2: l2, line3: l3,
					fillColor: color.RGBA{255, 255, 255, 255},
					fgColor:   color.RGBA{0, 0, 0, 255},
				})
			} else {
				fc := pngASNColor(seg.asn)
				nodes = append(nodes, nodeInfo{
					line1: l1, line2: l2, line3: l3,
					fillColor: fc, fgColor: pngASNFg(fc),
				})
			}
		} else {
			ip := ""
			if len(seg.ips) > 0 {
				ip = seg.ips[0]
			}
			nodes = append(nodes, nodeInfo{
				line1: ip, line2: hopRangeStr(seg.hops),
				fillColor: unknownFill, fgColor: white,
			})
		}
	}

	nodes = append(nodes, nodeInfo{
		line1: dstDisplay, line2: dstIP,
		fillColor: dstFill, fgColor: white,
	})

	for i := range nodes {
		w1 := textWidth(nodes[i].line1, textSc)
		w2 := textWidth(nodes[i].line2, textSc)
		w3 := 0
		if nodes[i].line3 != "" {
			w3 = textWidth(nodes[i].line3, textSc)
		}
		maxW := w1
		if w2 > maxW {
			maxW = w2
		}
		if w3 > maxW {
			maxW = w3
		}
		nodes[i].boxW = maxW + boxPadX*2

		lines := 1
		if nodes[i].line2 != "" {
			lines = 2
		}
		if nodes[i].line3 != "" {
			lines = 3
		}
		nodes[i].boxH = lines*lineH + boxPadY*2
	}

	type rowInfo struct {
		nodeIdxs []int
		totalW   int
		maxH     int
	}

	var rows []rowInfo
	curRow := rowInfo{}
	contentW := maxImgW - padding*2

	for i, n := range nodes {
		needed := n.boxW
		if len(curRow.nodeIdxs) > 0 {
			needed += arrowW
		}
		if curRow.totalW+needed > contentW && len(curRow.nodeIdxs) > 0 {
			rows = append(rows, curRow)
			curRow = rowInfo{}
		}
		if len(curRow.nodeIdxs) > 0 {
			curRow.totalW += arrowW
		}
		curRow.totalW += n.boxW
		if n.boxH > curRow.maxH {
			curRow.maxH = n.boxH
		}
		curRow.nodeIdxs = append(curRow.nodeIdxs, i)
	}
	if len(curRow.nodeIdxs) > 0 {
		rows = append(rows, curRow)
	}

	imgW := 0
	for _, r := range rows {
		rw := r.totalW + padding*2
		if rw > imgW {
			imgW = rw
		}
	}
	if imgW > maxImgW {
		imgW = maxImgW
	}

	titleH := 30
	totalH := padding + titleH
	for i, r := range rows {
		totalH += r.maxH
		if i < len(rows)-1 {
			totalH += rowGap
		}
	}
	totalH += padding

	var bgColor, titleFg, arrowColor, footerFg color.RGBA
	if bw {
		bgColor = color.RGBA{255, 255, 255, 255}
		titleFg = color.RGBA{0, 0, 0, 255}
		arrowColor = color.RGBA{0, 0, 0, 255}
		footerFg = color.RGBA{160, 160, 160, 255}
	} else {
		bgColor = color.RGBA{24, 24, 32, 255}
		titleFg = color.RGBA{200, 200, 200, 255}
		arrowColor = color.RGBA{100, 100, 120, 255}
		footerFg = color.RGBA{60, 60, 70, 255}
	}

	img := image.NewRGBA(image.Rect(0, 0, imgW, totalH))
	fillRect(img, 0, 0, imgW, totalH, bgColor)

	titleText := "PicoTR Route"
	titleTW := textWidth(titleText, textSc+1)
	drawText(img, (imgW-titleTW)/2, padding/2, titleText, titleFg, textSc+1)

	curY := padding + titleH

	for ri, row := range rows {
		curX := padding

		for ji, ni := range row.nodeIdxs {
			node := nodes[ni]
			bx := curX
			by := curY + (row.maxH-node.boxH)/2

			if bw {
				// White fill with black border
				drawRoundedRect(img, bx, by, node.boxW, node.boxH, 6, node.fillColor, color.RGBA{0, 0, 0, 255})
			} else {
				borderC := node.fillColor
				borderC.R = uint8(min255(int(borderC.R) + 40))
				borderC.G = uint8(min255(int(borderC.G) + 40))
				borderC.B = uint8(min255(int(borderC.B) + 40))
				drawRoundedRect(img, bx, by, node.boxW, node.boxH, 6, node.fillColor, borderC)
			}

			ty := by + boxPadY
			tw1 := textWidth(node.line1, textSc)
			drawText(img, bx+(node.boxW-tw1)/2, ty, node.line1, node.fgColor, textSc)

			if node.line2 != "" {
				ty += lineH
				tw2 := textWidth(node.line2, textSc)
				subFg := node.fgColor
				if bw {
					subFg = color.RGBA{80, 80, 80, 255}
				} else {
					if subFg.R > 40 {
						subFg.R -= 40
					}
					if subFg.G > 40 {
						subFg.G -= 40
					}
					if subFg.B > 40 {
						subFg.B -= 40
					}
				}
				drawText(img, bx+(node.boxW-tw2)/2, ty, node.line2, subFg, textSc)
			}

			if node.line3 != "" {
				ty += lineH
				tw3 := textWidth(node.line3, textSc)
				var dimFg color.RGBA
				if bw {
					dimFg = color.RGBA{120, 120, 120, 255}
				} else {
					dimFg = color.RGBA{
						uint8(min255(int(node.fgColor.R)/2 + 60)),
						uint8(min255(int(node.fgColor.G)/2 + 60)),
						uint8(min255(int(node.fgColor.B)/2 + 60)),
						255,
					}
				}
				drawText(img, bx+(node.boxW-tw3)/2, ty, node.line3, dimFg, textSc)
			}

			curX += node.boxW

			if ji < len(row.nodeIdxs)-1 {
				acy := curY + row.maxH/2
				drawArrowRight(img, curX, curX+arrowW, acy, arrowColor)
				curX += arrowW
			}
		}

		if ri < len(rows)-1 {
			lastIdx := row.nodeIdxs[len(row.nodeIdxs)-1]
			lastNode := nodes[lastIdx]
			connX := curX - lastNode.boxW/2
			lastBoxBottom := curY + (row.maxH+lastNode.boxH)/2
			connY1 := lastBoxBottom

			nextRow := rows[ri+1]
			nextFirstNode := nodes[nextRow.nodeIdxs[0]]
			nextConnX := padding + nextFirstNode.boxW/2
			nextRowY := curY + row.maxH + rowGap
			nextBoxTop := nextRowY + (nextRow.maxH-nextFirstNode.boxH)/2
			connY2 := nextBoxTop

			midY := (connY1 + connY2) / 2

			for py := connY1; py <= midY; py++ {
				img.SetRGBA(connX-1, py, arrowColor)
				img.SetRGBA(connX, py, arrowColor)
				img.SetRGBA(connX+1, py, arrowColor)
			}

			lx, rx := nextConnX, connX
			if lx > rx {
				lx, rx = rx, lx
			}
			for px := lx; px <= rx; px++ {
				img.SetRGBA(px, midY-1, arrowColor)
				img.SetRGBA(px, midY, arrowColor)
				img.SetRGBA(px, midY+1, arrowColor)
			}

			drawArrowDown(img, nextConnX, midY, connY2, arrowColor)
		}

		curY += row.maxH
		if ri < len(rows)-1 {
			curY += rowGap
		}
	}

	footerY := totalH - 7*textSc - 6
	ghURL := "https://github.com/VEX-Systems/picotr"
	drawText(img, 10, footerY, ghURL, footerFg, textSc)
	watermark := "picotr"
	wmW := textWidth(watermark, textSc)
	drawText(img, imgW-wmW-10, footerY, watermark, footerFg, textSc)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

type traceRow struct {
	ttl     int
	host    string
	ip      string
	loss    float64
	sent    int
	rcv     int
	last    float64
	avg     float64
	best    float64
	worst   float64
	hasData bool
	asn     uint32
	asName  string
}

type traceExportOpts struct {
	showIP  bool
	precise bool
	units   bool
	showAS  bool
	noColor bool
}

func pngPingBarColor(ms float64) color.RGBA {
	switch {
	case ms < 20:
		return color.RGBA{0, 200, 0, 255}
	case ms < 50:
		return color.RGBA{100, 220, 0, 255}
	case ms < 100:
		return color.RGBA{220, 220, 0, 255}
	case ms < 150:
		return color.RGBA{220, 150, 0, 255}
	case ms < 300:
		return color.RGBA{220, 50, 0, 255}
	default:
		return color.RGBA{180, 0, 0, 255}
	}
}

func pngLossBarColor(pct float64) color.RGBA {
	switch {
	case pct == 0:
		return color.RGBA{0, 200, 0, 255}
	case pct < 10:
		return color.RGBA{220, 220, 0, 255}
	case pct < 25:
		return color.RGBA{220, 150, 0, 255}
	case pct < 50:
		return color.RGBA{220, 50, 0, 255}
	default:
		return color.RGBA{180, 0, 0, 255}
	}
}

func pngRTTColor(ms float64) color.RGBA {
	switch {
	case ms < 20:
		return color.RGBA{0, 220, 0, 255}
	case ms < 50:
		return color.RGBA{100, 220, 0, 255}
	case ms < 100:
		return color.RGBA{220, 220, 0, 255}
	case ms < 150:
		return color.RGBA{220, 150, 0, 255}
	case ms < 300:
		return color.RGBA{220, 50, 0, 255}
	default:
		return color.RGBA{180, 0, 0, 255}
	}
}

func ExportTracePNG(filename string, headerInfo string, rows []traceRow, opts traceExportOpts) error {
	sc := 2
	charW := 6 * sc
	charH := 7 * sc
	lineGap := 4
	rowH := charH*2 + lineGap + 10
	if opts.showIP {
		rowH = charH + 10
	}
	hdrRowH := charH + 10
	pad := 30
	barW := 80
	barH := charH + 2
	barGap := 6

	rttChars := 7
	if opts.precise {
		rttChars = 9
	}
	if opts.units {
		rttChars += 2
	}

	colTTL := 4 * charW
	colASN := 0
	if opts.showAS {
		colASN = 10 * charW
	}
	colHost := 38 * charW
	if opts.showIP {
		colHost = 20 * charW
	}
	colLoss := barW + barGap + 5*charW + charW
	colPing := barW + barGap + rttChars*charW + charW
	colRTT := rttChars * charW

	imgW := pad + colTTL + colASN + colHost + colLoss + colPing + colRTT*4 + pad
	titleSc := sc + 1
	titleCharH := 7 * titleSc
	titleRowH := titleCharH + 8 + charH + 8
	sepH := 4
	footerH := charH + 10
	imgH := titleRowH + hdrRowH + sepH + len(rows)*rowH + footerH

	bw := opts.noColor

	var bgColor, titleBg, rowAlt, sepColor, white, dimWhite, grayFg, barBgC, barFillC, hdrColor, footerFg color.RGBA
	if bw {
		bgColor = color.RGBA{255, 255, 255, 255}
		titleBg = color.RGBA{240, 240, 240, 255}
		rowAlt = color.RGBA{245, 245, 245, 255}
		sepColor = color.RGBA{180, 180, 180, 255}
		white = color.RGBA{0, 0, 0, 255}       // main text is black
		dimWhite = color.RGBA{80, 80, 80, 255}  // secondary text dark gray
		grayFg = color.RGBA{140, 140, 140, 255} // placeholder text
		barBgC = color.RGBA{220, 220, 220, 255}
		barFillC = color.RGBA{60, 60, 60, 255}
		hdrColor = color.RGBA{0, 0, 0, 255}
		footerFg = color.RGBA{160, 160, 160, 255}
	} else {
		bgColor = color.RGBA{24, 24, 32, 255}
		titleBg = color.RGBA{35, 35, 50, 255}
		rowAlt = color.RGBA{28, 28, 38, 255}
		sepColor = color.RGBA{50, 50, 65, 255}
		white = color.RGBA{220, 220, 220, 255}
		dimWhite = color.RGBA{140, 140, 150, 255}
		grayFg = color.RGBA{100, 100, 110, 255}
		barBgC = color.RGBA{40, 40, 50, 255}
		barFillC = color.RGBA{0, 0, 0, 255}
		hdrColor = color.RGBA{180, 180, 200, 255}
		footerFg = color.RGBA{60, 60, 70, 255}
	}

	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	fillRect(img, 0, 0, imgW, imgH, bgColor)

	fillRect(img, 0, 0, imgW, titleRowH, titleBg)
	titleText := "PicoTR Trace"
	tw := textWidth(titleText, titleSc)
	titleFg := white
	if bw {
		titleFg = color.RGBA{0, 0, 0, 255}
	} else {
		titleFg = color.RGBA{200, 200, 200, 255}
	}
	drawText(img, (imgW-tw)/2, 8, titleText, titleFg, titleSc)
	sw := textWidth(headerInfo, sc)
	drawText(img, (imgW-sw)/2, 8+titleCharH+6, headerInfo, dimWhite, sc)

	y := titleRowH
	x := pad
	hdrTextY := y + (hdrRowH-charH)/2

	drawText(img, x, hdrTextY, "#", hdrColor, sc)
	x += colTTL
	if opts.showAS {
		drawText(img, x, hdrTextY, "ASN", hdrColor, sc)
		x += colASN
	}
	drawText(img, x, hdrTextY, "Host", hdrColor, sc)
	x += colHost
	drawText(img, x, hdrTextY, "Loss", hdrColor, sc)
	x += colLoss
	drawText(img, x, hdrTextY, "Ping", hdrColor, sc)
	x += colPing
	drawText(img, x, hdrTextY, "Last", hdrColor, sc)
	x += colRTT
	drawText(img, x, hdrTextY, "Avg", hdrColor, sc)
	x += colRTT
	drawText(img, x, hdrTextY, "Best", hdrColor, sc)
	x += colRTT
	drawText(img, x, hdrTextY, "Wrst", hdrColor, sc)

	y += hdrRowH
	fillRect(img, pad, y, imgW-pad*2, 2, sepColor)
	y += sepH

	maxHostChars := (colHost - charW) / charW

	for i, row := range rows {
		if i%2 == 1 {
			fillRect(img, 0, y, imgW, rowH, rowAlt)
		}

		line1Y := y + 5
		line2Y := line1Y + charH + lineGap
		centerY := y + (rowH-charH)/2

		x = pad
		drawText(img, x, centerY, fmt.Sprintf("%d", row.ttl), dimWhite, sc)
		x += colTTL

		if opts.showAS {
			if row.asn != 0 {
				asnLabel := fmt.Sprintf("AS%d", row.asn)
				asnW := textWidth(asnLabel, sc) + charW
				asnBadgeY := y + (rowH-(charH+4))/2
				if bw {
					// outlined badge
					drawStrokeRect(img, x, asnBadgeY, asnW, charH+4, color.RGBA{0, 0, 0, 255})
					drawText(img, x+charW/2, asnBadgeY+2, asnLabel, color.RGBA{0, 0, 0, 255}, sc)
				} else {
					fc := pngASNColor(row.asn)
					fg := pngASNFg(fc)
					fillRect(img, x, asnBadgeY, asnW, charH+4, fc)
					drawText(img, x+charW/2, asnBadgeY+2, asnLabel, fg, sc)
				}
			}
			x += colASN
		}

		if row.host == "???" {
			drawText(img, x, centerY, "???", grayFg, sc)
		} else if opts.showIP {
			ipStr := row.ip
			if ipStr == "" {
				ipStr = row.host
			}
			if len(ipStr) > maxHostChars {
				ipStr = ipStr[:maxHostChars-3] + "..."
			}
			drawText(img, x, centerY, ipStr, white, sc)
		} else {
			if row.ip != "" && row.host != row.ip {
				hostDisplay := row.host
				if len(hostDisplay) > maxHostChars {
					hostDisplay = hostDisplay[:maxHostChars-3] + "..."
				}
				drawText(img, x, line1Y, hostDisplay, white, sc)

				ipStr := "(" + row.ip + ")"
				if len(ipStr) > maxHostChars {
					ipStr = ipStr[:maxHostChars-3] + "..."
				}
				drawText(img, x, line2Y, ipStr, dimWhite, sc)
			} else {
				ipStr := row.ip
				if ipStr == "" {
					ipStr = row.host
				}
				if len(ipStr) > maxHostChars {
					ipStr = ipStr[:maxHostChars-3] + "..."
				}
				drawText(img, x, centerY, ipStr, white, sc)
			}
		}
		x += colHost

		lossStr := fmt.Sprintf("%.0f%%", row.loss)
		lossFill := int(row.loss / 100 * float64(barW))
		if row.loss > 0 && lossFill < 3 {
			lossFill = 3
		}
		barY := y + (rowH-barH)/2
		fillRect(img, x, barY, barW, barH, barBgC)
		if bw {
			drawStrokeRect(img, x, barY, barW, barH, color.RGBA{0, 0, 0, 255})
			if lossFill > 0 {
				fillRect(img, x+1, barY+1, lossFill, barH-2, barFillC)
			}
			drawText(img, x+barW+barGap, centerY, lossStr, white, sc)
		} else {
			lossC := pngLossBarColor(row.loss)
			if lossFill > 0 {
				fillRect(img, x, barY, lossFill, barH, lossC)
			}
			drawText(img, x+barW+barGap, centerY, lossStr, lossC, sc)
		}
		x += colLoss

		if row.hasData {
			pingFill := pngPingFilled(row.avg)
			fillRect(img, x, barY, barW, barH, barBgC)
			if bw {
				drawStrokeRect(img, x, barY, barW, barH, color.RGBA{0, 0, 0, 255})
				if pingFill > 0 {
					fillRect(img, x+1, barY+1, pingFill, barH-2, barFillC)
				}
				drawText(img, x+barW+barGap, centerY, fmtRTT(row.avg, opts), white, sc)
			} else {
				if pingFill > 0 {
					fillRect(img, x, barY, pingFill, barH, pngPingBarColor(row.avg))
				}
				drawText(img, x+barW+barGap, centerY, fmtRTT(row.avg, opts), pngRTTColor(row.avg), sc)
			}
		} else {
			fillRect(img, x, barY, barW, barH, barBgC)
			if bw {
				drawStrokeRect(img, x, barY, barW, barH, color.RGBA{0, 0, 0, 255})
			}
			drawText(img, x+barW+barGap, centerY, "-", grayFg, sc)
		}
		x += colPing

		rttFg := func(ms float64) color.RGBA {
			if bw {
				return white
			}
			return pngRTTColor(ms)
		}

		if row.hasData {
			drawText(img, x, centerY, fmtRTT(row.last, opts), rttFg(row.last), sc)
			x += colRTT
			drawText(img, x, centerY, fmtRTT(row.avg, opts), rttFg(row.avg), sc)
			x += colRTT
			if row.best < math.MaxFloat64 {
				drawText(img, x, centerY, fmtRTT(row.best, opts), rttFg(row.best), sc)
			} else {
				drawText(img, x, centerY, "-", grayFg, sc)
			}
			x += colRTT
			drawText(img, x, centerY, fmtRTT(row.worst, opts), rttFg(row.worst), sc)
		} else {
			drawText(img, x, centerY, "-", grayFg, sc)
			x += colRTT
			drawText(img, x, centerY, "-", grayFg, sc)
			x += colRTT
			drawText(img, x, centerY, "-", grayFg, sc)
			x += colRTT
			drawText(img, x, centerY, "-", grayFg, sc)
		}

		y += rowH
	}

	footerY := imgH - footerH + (footerH-charH)/2
	ghURL := "https://github.com/VEX-Systems/picotr"
	drawText(img, 10, footerY, ghURL, footerFg, sc)
	watermark := "picotr"
	wmW := textWidth(watermark, sc)
	drawText(img, imgW-wmW-10, footerY, watermark, footerFg, sc)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func pngPingFilled(avgMs float64) int {
	switch {
	case avgMs <= 0:
		return 0
	case avgMs < 5:
		return 6
	case avgMs < 10:
		return 12
	case avgMs < 20:
		return 18
	case avgMs < 40:
		return 24
	case avgMs < 60:
		return 30
	case avgMs < 100:
		return 36
	case avgMs < 150:
		return 42
	case avgMs < 200:
		return 48
	case avgMs < 300:
		return 54
	default:
		return 60
	}
}

func buildTraceRows(hops []trace.HopStats, maxTTL int, resolveHost func(string) string, getAS func(string) *asInfo) []traceRow {
	reached := false
	for i := 0; i < maxTTL; i++ {
		if hops[i].Reached {
			reached = true
			break
		}
	}

	displayMax := maxTTL
	if !reached {
		last := 0
		for i := maxTTL - 1; i >= 0; i-- {
			if hops[i].Addr != nil {
				last = i + 1
				break
			}
		}
		if last > 0 {
			displayMax = last
		}
	}

	var rows []traceRow
	for i := 0; i < displayMax; i++ {
		h := hops[i]
		r := traceRow{
			ttl:  h.TTL,
			host: "???",
			sent: h.Sent,
			rcv:  h.Received,
		}
		if h.Addr != nil {
			r.ip = h.Addr.String()
			r.host = resolveHost(r.ip)
			info := getAS(r.ip)
			if info != nil && info.ASN != 0 {
				r.asn = info.ASN
				r.asName = info.ASName
			}
		}
		if h.Sent > 0 {
			r.loss = float64(h.Lost) / float64(h.Sent) * 100
		}
		if h.Received > 0 {
			r.hasData = true
			r.last = h.Last
			r.avg = h.Avg
			r.best = h.Best
			r.worst = h.Worst
		}
		rows = append(rows, r)
	}
	return rows
}

func fmtRTT(ms float64, opts traceExportOpts) string {
	if opts.precise {
		if opts.units {
			return fmt.Sprintf("%.2fms", ms)
		}
		return fmt.Sprintf("%.2f", ms)
	}
	if opts.units {
		return fmt.Sprintf("%.0fms", ms)
	}
	return fmt.Sprintf("%.0f", ms)
}

func min255(v int) int {
	if v > 255 {
		return 255
	}
	return v
}

// asnGroup represents a visual group of hop nodes sharing the same ASN.
type asnGroup struct {
	asn       uint32
	name      string
	nodeIdxs  []int // indices into the flat nodes slice
	isWrapper bool  // true if this group draws a container box
}

// ExportAllHopsGroupedPNG renders an "all hops" route PNG where consecutive
// hops sharing the same ASN are visually grouped inside an outlined container
// box with the ASN label on top.
func ExportAllHopsGroupedPNG(filename string, localIP, dstDisplay, dstIP string, segments []routeSegment, noColor bool) error {
	scale := 2
	padding := 40
	arrowW := 30
	textSc := scale
	lineH := 7*textSc + 4
	boxPadX := 14
	boxPadY := 8
	groupPadX := 12
	groupPadTop := lineH + 10 // space for ASN label at top
	groupPadBot := 10
	innerArrowW := 20
	rowGap := 50
	maxImgW := 1800
	bw := noColor

	type nodeInfo struct {
		line1     string
		line2     string
		fillColor color.RGBA
		fgColor   color.RGBA
		boxW      int
		boxH      int
	}

	var nodes []nodeInfo
	var srcFill, dstFill, gapFill, unknownFill, white, grayFg color.RGBA
	if bw {
		srcFill = color.RGBA{255, 255, 255, 255}
		dstFill = color.RGBA{255, 255, 255, 255}
		gapFill = color.RGBA{255, 255, 255, 255}
		unknownFill = color.RGBA{255, 255, 255, 255}
		white = color.RGBA{0, 0, 0, 255}
		grayFg = color.RGBA{100, 100, 100, 255}
	} else {
		srcFill = color.RGBA{30, 90, 150, 255}
		dstFill = color.RGBA{30, 90, 150, 255}
		gapFill = color.RGBA{60, 60, 60, 255}
		unknownFill = color.RGBA{80, 80, 80, 255}
		white = color.RGBA{255, 255, 255, 255}
		grayFg = color.RGBA{160, 160, 160, 255}
	}

	// Build source node
	nodes = append(nodes, nodeInfo{
		line1: localIP, line2: "Source",
		fillColor: srcFill, fgColor: white,
	})

	// Build per-hop nodes
	for _, seg := range segments {
		if seg.isGap {
			nodes = append(nodes, nodeInfo{
				line1: "???", line2: hopRangeStr(seg.hops),
				fillColor: gapFill, fgColor: grayFg,
			})
			continue
		}
		ip := ""
		if len(seg.ips) > 0 {
			ip = seg.ips[0]
		}
		// Use resolved hostname if available, otherwise IP
		host := ""
		if len(seg.hosts) > 0 && seg.hosts[0] != "" {
			host = seg.hosts[0]
		}
		l1 := ip
		l2 := host
		if l2 == ip {
			l2 = ""
		}
		if len(l2) > 30 {
			l2 = l2[:27] + "..."
		}
		hopLabel := hopRangeStr(seg.hops)
		if l2 == "" {
			l2 = hopLabel
		}

		if seg.asn != 0 {
			if bw {
				nodes = append(nodes, nodeInfo{
					line1: l1, line2: l2,
					fillColor: color.RGBA{255, 255, 255, 255},
					fgColor:   color.RGBA{0, 0, 0, 255},
				})
			} else {
				fc := pngASNColor(seg.asn)
				nodes = append(nodes, nodeInfo{
					line1: l1, line2: l2,
					fillColor: fc, fgColor: pngASNFg(fc),
				})
			}
		} else {
			nodes = append(nodes, nodeInfo{
				line1: l1, line2: l2,
				fillColor: unknownFill, fgColor: white,
			})
		}
	}

	// Build destination node
	nodes = append(nodes, nodeInfo{
		line1: dstDisplay, line2: dstIP,
		fillColor: dstFill, fgColor: white,
	})

	// Compute box sizes
	for i := range nodes {
		w1 := textWidth(nodes[i].line1, textSc)
		w2 := textWidth(nodes[i].line2, textSc)
		maxW := w1
		if w2 > maxW {
			maxW = w2
		}
		nodes[i].boxW = maxW + boxPadX*2
		if nodes[i].boxW < 60 {
			nodes[i].boxW = 60
		}
		lines := 1
		if nodes[i].line2 != "" {
			lines = 2
		}
		nodes[i].boxH = lines*lineH + boxPadY*2
	}

	// Group consecutive hops by ASN.
	// Node 0 = source, nodes 1..N-2 = hops, node N-1 = destination
	// Source and destination are always standalone.
	var groups []asnGroup

	// Source: standalone
	groups = append(groups, asnGroup{nodeIdxs: []int{0}})

	// Group hop nodes (indices 1 through len(nodes)-2)
	hopStart := 1
	hopEnd := len(nodes) - 1
	i := hopStart
	for i < hopEnd {
		seg := segments[i-hopStart]
		if seg.isGap || seg.asn == 0 {
			groups = append(groups, asnGroup{nodeIdxs: []int{i}})
			i++
			continue
		}
		// Collect consecutive hops with the same ASN
		asn := seg.asn
		name := seg.name
		start := i
		for i < hopEnd {
			s := segments[i-hopStart]
			if s.asn != asn {
				break
			}
			i++
		}
		idxs := make([]int, 0, i-start)
		for j := start; j < i; j++ {
			idxs = append(idxs, j)
		}
		isWrapper := len(idxs) > 1
		groups = append(groups, asnGroup{asn: asn, name: name, nodeIdxs: idxs, isWrapper: isWrapper})
	}

	// Destination: standalone
	groups = append(groups, asnGroup{nodeIdxs: []int{len(nodes) - 1}})

	// Compute visual width & height for each group
	type groupMetrics struct {
		w, h       int
		innerW     int // width of inner content
		innerH     int // height of inner content
		isWrapper  bool
		firstNodeI int // first node index (for arrow connection)
		lastNodeI  int // last node index (for arrow connection)
	}

	var gMetrics []groupMetrics
	for _, g := range groups {
		if g.isWrapper {
			// Container: inner boxes laid out horizontally with inner arrows
			innerW := 0
			maxInnerH := 0
			for ji, ni := range g.nodeIdxs {
				if ji > 0 {
					innerW += innerArrowW
				}
				innerW += nodes[ni].boxW
				if nodes[ni].boxH > maxInnerH {
					maxInnerH = nodes[ni].boxH
				}
			}
			totalW := innerW + groupPadX*2
			totalH := maxInnerH + groupPadTop + groupPadBot

			// Ensure container is wide enough for the ASN label
			asnLabel := fmt.Sprintf("AS%d", g.asn)
			if g.name != "" {
				asnLabel += " - " + g.name
			}
			if len(asnLabel) > 50 {
				asnLabel = asnLabel[:47] + "..."
			}
			labelW := textWidth(asnLabel, textSc) + groupPadX*2
			if totalW < labelW {
				totalW = labelW
			}

			gMetrics = append(gMetrics, groupMetrics{
				w: totalW, h: totalH,
				innerW: innerW, innerH: maxInnerH,
				isWrapper:  true,
				firstNodeI: g.nodeIdxs[0],
				lastNodeI:  g.nodeIdxs[len(g.nodeIdxs)-1],
			})
		} else {
			ni := g.nodeIdxs[0]
			gMetrics = append(gMetrics, groupMetrics{
				w: nodes[ni].boxW, h: nodes[ni].boxH,
				isWrapper:  false,
				firstNodeI: ni,
				lastNodeI:  ni,
			})
		}
	}

	// Row layout: wrap groups into rows
	type rowInfo struct {
		groupIdxs []int
		totalW    int
		maxH      int
	}

	var rows []rowInfo
	contentW := maxImgW - padding*2
	curRow := rowInfo{}

	for gi, gm := range gMetrics {
		needed := gm.w
		if len(curRow.groupIdxs) > 0 {
			needed += arrowW
		}
		if curRow.totalW+needed > contentW && len(curRow.groupIdxs) > 0 {
			rows = append(rows, curRow)
			curRow = rowInfo{}
		}
		if len(curRow.groupIdxs) > 0 {
			curRow.totalW += arrowW
		}
		curRow.totalW += gm.w
		if gm.h > curRow.maxH {
			curRow.maxH = gm.h
		}
		curRow.groupIdxs = append(curRow.groupIdxs, gi)
	}
	if len(curRow.groupIdxs) > 0 {
		rows = append(rows, curRow)
	}

	// Image dimensions
	imgW := 0
	for _, r := range rows {
		rw := r.totalW + padding*2
		if rw > imgW {
			imgW = rw
		}
	}
	if imgW > maxImgW {
		imgW = maxImgW
	}

	titleH := 30
	totalH := padding + titleH
	for i, r := range rows {
		totalH += r.maxH
		if i < len(rows)-1 {
			totalH += rowGap
		}
	}
	totalH += padding

	var bgColor, titleFg, arrowColor, footerFg, groupBorderC, groupLabelFg color.RGBA
	if bw {
		bgColor = color.RGBA{255, 255, 255, 255}
		titleFg = color.RGBA{0, 0, 0, 255}
		arrowColor = color.RGBA{0, 0, 0, 255}
		footerFg = color.RGBA{160, 160, 160, 255}
		groupBorderC = color.RGBA{0, 0, 0, 255}
		groupLabelFg = color.RGBA{0, 0, 0, 255}
	} else {
		bgColor = color.RGBA{24, 24, 32, 255}
		titleFg = color.RGBA{200, 200, 200, 255}
		arrowColor = color.RGBA{100, 100, 120, 255}
		footerFg = color.RGBA{60, 60, 70, 255}
		groupBorderC = color.RGBA{80, 80, 100, 255}
		groupLabelFg = color.RGBA{180, 180, 200, 255}
	}

	img := image.NewRGBA(image.Rect(0, 0, imgW, totalH))
	fillRect(img, 0, 0, imgW, totalH, bgColor)

	titleText := "PicoTR All Hops"
	titleTW := textWidth(titleText, textSc+1)
	drawText(img, (imgW-titleTW)/2, padding/2, titleText, titleFg, textSc+1)

	curY := padding + titleH

	for ri, row := range rows {
		curX := padding

		for ji, gi := range row.groupIdxs {
			g := groups[gi]
			gm := gMetrics[gi]

			if gm.isWrapper {
				// Draw container outline
				gx := curX
				gy := curY + (row.maxH-gm.h)/2

				containerFill := bgColor
				if bw {
					containerFill = color.RGBA{248, 248, 248, 255}
				} else {
					containerFill = color.RGBA{30, 30, 42, 255}
				}

				// Use ASN color for border if color is enabled
				borderC := groupBorderC
				if !bw && g.asn != 0 {
					c := pngASNColor(g.asn)
					// Dim the ASN color for the border
					borderC = color.RGBA{c.R / 2, c.G / 2, c.B / 2, 255}
				}

				// Draw rounded container
				drawRoundedRect(img, gx, gy, gm.w, gm.h, 8, containerFill, borderC)

				// Draw ASN label at top of container
				asnLabel := fmt.Sprintf("AS%d", g.asn)
				if g.name != "" {
					asnLabel += " - " + g.name
				}
				if len(asnLabel) > 50 {
					asnLabel = asnLabel[:47] + "..."
				}

				labelFg := groupLabelFg
				if !bw && g.asn != 0 {
					labelFg = pngASNColor(g.asn)
				}
				labelTW := textWidth(asnLabel, textSc)
				drawText(img, gx+(gm.w-labelTW)/2, gy+6, asnLabel, labelFg, textSc)

				// Draw inner hop boxes
				innerX := gx + groupPadX
				innerY := gy + groupPadTop

				for ki, ni := range g.nodeIdxs {
					node := nodes[ni]
					bx := innerX
					by := innerY + (gm.innerH-node.boxH)/2

					if bw {
						drawRoundedRect(img, bx, by, node.boxW, node.boxH, 5, node.fillColor, color.RGBA{0, 0, 0, 255})
					} else {
						bdrC := node.fillColor
						bdrC.R = uint8(min255(int(bdrC.R) + 40))
						bdrC.G = uint8(min255(int(bdrC.G) + 40))
						bdrC.B = uint8(min255(int(bdrC.B) + 40))
						drawRoundedRect(img, bx, by, node.boxW, node.boxH, 5, node.fillColor, bdrC)
					}

					ty := by + boxPadY
					tw1 := textWidth(node.line1, textSc)
					drawText(img, bx+(node.boxW-tw1)/2, ty, node.line1, node.fgColor, textSc)
					if node.line2 != "" {
						ty += lineH
						tw2 := textWidth(node.line2, textSc)
						subFg := node.fgColor
						if bw {
							subFg = color.RGBA{80, 80, 80, 255}
						} else {
							if subFg.R > 40 {
								subFg.R -= 40
							}
							if subFg.G > 40 {
								subFg.G -= 40
							}
							if subFg.B > 40 {
								subFg.B -= 40
							}
						}
						drawText(img, bx+(node.boxW-tw2)/2, ty, node.line2, subFg, textSc)
					}

					innerX += node.boxW

					if ki < len(g.nodeIdxs)-1 {
						acy := innerY + gm.innerH/2
						drawArrowRight(img, innerX, innerX+innerArrowW, acy, arrowColor)
						innerX += innerArrowW
					}
				}
			} else {
				// Standalone node
				ni := g.nodeIdxs[0]
				node := nodes[ni]
				bx := curX
				by := curY + (row.maxH-node.boxH)/2

				if bw {
					drawRoundedRect(img, bx, by, node.boxW, node.boxH, 6, node.fillColor, color.RGBA{0, 0, 0, 255})
				} else {
					bdrC := node.fillColor
					bdrC.R = uint8(min255(int(bdrC.R) + 40))
					bdrC.G = uint8(min255(int(bdrC.G) + 40))
					bdrC.B = uint8(min255(int(bdrC.B) + 40))
					drawRoundedRect(img, bx, by, node.boxW, node.boxH, 6, node.fillColor, bdrC)
				}

				ty := by + boxPadY
				tw1 := textWidth(node.line1, textSc)
				drawText(img, bx+(node.boxW-tw1)/2, ty, node.line1, node.fgColor, textSc)
				if node.line2 != "" {
					ty += lineH
					tw2 := textWidth(node.line2, textSc)
					subFg := node.fgColor
					if bw {
						subFg = color.RGBA{80, 80, 80, 255}
					} else {
						if subFg.R > 40 {
							subFg.R -= 40
						}
						if subFg.G > 40 {
							subFg.G -= 40
						}
						if subFg.B > 40 {
							subFg.B -= 40
						}
					}
					drawText(img, bx+(node.boxW-tw2)/2, ty, node.line2, subFg, textSc)
				}
			}

			curX += gm.w

			// Arrow to next group in same row
			if ji < len(row.groupIdxs)-1 {
				acy := curY + row.maxH/2
				drawArrowRight(img, curX, curX+arrowW, acy, arrowColor)
				curX += arrowW
			}
		}

		// L-shaped connector between rows
		if ri < len(rows)-1 {
			lastGI := row.groupIdxs[len(row.groupIdxs)-1]
			lastGM := gMetrics[lastGI]
			connX := curX - lastGM.w/2
			lastGY := curY + (row.maxH+lastGM.h)/2
			connY1 := lastGY

			nextRow := rows[ri+1]
			nextGI := nextRow.groupIdxs[0]
			nextGM := gMetrics[nextGI]
			nextConnX := padding + nextGM.w/2
			nextRowY := curY + row.maxH + rowGap
			nextGTop := nextRowY + (nextRow.maxH-nextGM.h)/2
			connY2 := nextGTop

			midY := (connY1 + connY2) / 2

			for py := connY1; py <= midY; py++ {
				img.SetRGBA(connX-1, py, arrowColor)
				img.SetRGBA(connX, py, arrowColor)
				img.SetRGBA(connX+1, py, arrowColor)
			}

			lx, rx := nextConnX, connX
			if lx > rx {
				lx, rx = rx, lx
			}
			for px := lx; px <= rx; px++ {
				img.SetRGBA(px, midY-1, arrowColor)
				img.SetRGBA(px, midY, arrowColor)
				img.SetRGBA(px, midY+1, arrowColor)
			}

			drawArrowDown(img, nextConnX, midY, connY2, arrowColor)
		}

		curY += row.maxH
		if ri < len(rows)-1 {
			curY += rowGap
		}
	}

	footerY := totalH - 7*textSc - 6
	ghURL := "https://github.com/VEX-Systems/picotr"
	drawText(img, 10, footerY, ghURL, footerFg, textSc)
	watermark := "picotr"
	wmW := textWidth(watermark, textSc)
	drawText(img, imgW-wmW-10, footerY, watermark, footerFg, textSc)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

