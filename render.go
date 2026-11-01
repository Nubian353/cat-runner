package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ---------- canvas ----------

type cell struct {
	r    rune
	fg   int
	bg   int
	bold bool
}

type canvas struct {
	w, h  int
	cells []cell
}

func newCanvas(w, h int) *canvas {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	c := &canvas{w: w, h: h, cells: make([]cell, w*h)}
	for i := range c.cells {
		c.cells[i] = cell{r: ' ', fg: -1, bg: -1}
	}
	return c
}

func (c *canvas) put(x, y int, r rune, fg, bg int) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	c.cells[y*c.w+x] = cell{r: r, fg: fg, bg: bg}
}

// ink draws a character while keeping the background already painted there.
func (c *canvas) ink(x, y int, r rune, fg int, bold bool) {
	if r == ' ' || x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	cl := &c.cells[y*c.w+x]
	cl.r = r
	cl.fg = fg
	cl.bold = bold
}

func (c *canvas) fill(x, y, w, h int, r rune, fg, bg int) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			c.put(xx, yy, r, fg, bg)
		}
	}
}

func (c *canvas) inkText(x, y int, s string, fg int, bold bool) {
	for i, r := range []rune(s) {
		c.ink(x+i, y, r, fg, bold)
	}
}

func (c *canvas) inkTextCentered(y int, s string, fg int, bold bool) {
	c.inkText((c.w-runeLen(s))/2, y, s, fg, bold)
}

func runeLen(s string) int {
	return len([]rune(s))
}

func (c *canvas) render() string {
	var b strings.Builder
	b.Grow(c.w*c.h*4 + 128)
	b.WriteString("\x1b[H")

	curFg, curBg, curBold := -2, -2, false
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			cl := c.cells[y*c.w+x]
			if cl.fg != curFg || cl.bg != curBg || cl.bold != curBold {
				b.WriteString("\x1b[0")
				if cl.bold {
					b.WriteString(";1")
				}
				if cl.fg < 0 {
					b.WriteString(";39")
				} else {
					b.WriteString(";38;5;")
					b.WriteString(strconv.Itoa(cl.fg))
				}
				if cl.bg < 0 {
					b.WriteString(";49")
				} else {
					b.WriteString(";48;5;")
					b.WriteString(strconv.Itoa(cl.bg))
				}
				b.WriteByte('m')
				curFg, curBg, curBold = cl.fg, cl.bg, cl.bold
			}
			b.WriteRune(cl.r)
		}
		b.WriteString("\x1b[0m")
		if y != c.h-1 {
			b.WriteString("\r\n")
		}
		// The reset above means the next row needs its style emitted again.
		curFg, curBg, curBold = -2, -2, false
	}
	return b.String()
}

// ---------- scene ----------

var skyStops = []int{17, 18, 60, 61, 103, 139, 181, 223}

func skyColor(y, h int) int {
	if h <= 1 {
		return skyStops[len(skyStops)-1]
	}
	f := float64(y) / float64(h-1)
	i := int(f*float64(len(skyStops)-1) + 0.5)
	if i < 0 {
		i = 0
	}
	if i >= len(skyStops) {
		i = len(skyStops) - 1
	}
	return skyStops[i]
}

func hash2(x, y int) uint32 {
	h := uint32(x)*374761393 + uint32(y)*668265263
	h = (h ^ (h >> 13)) * 1274126177
	return h ^ (h >> 16)
}

func (g *game) drawSky(c *canvas) {
	rt := g.roadTop()
	c.fill(0, 0, c.w, rt, ' ', -1, 17)

	// stars
	for y := 2; y < rt-5; y++ {
		for x := 1; x < c.w-1; x += 2 {
			h := hash2(x+7, y*13+3)
			if h%23 != 0 {
				continue
			}
			phase := float64(h%628) / 100
			tw := 0.55 + 0.45*math.Sin(g.t*1.8+phase)
			col := 239
			if tw > 0.78 {
				col = 255
			} else if tw > 0.42 {
				col = 246
			}
			ch := '·'
			if h%7 == 0 {
				ch = '+'
			}
			c.ink(x, y, ch, col, false)
		}
	}

	g.drawMoon(c)
}

func (g *game) drawMoon(c *canvas) {
	rt := g.roadTop()
	mx := int(float64(c.w) * 0.80)
	my := 6
	if my > rt-9 {
		my = rt - 9
	}
	if my < 4 {
		my = 4
	}

	rows := []string{
		`   ▄▄▄▄▄   `,
		` ▄███████▄ `,
		`███████████`,
		` ▀███████▀ `,
		`   ▀▀▀▀▀   `,
	}
	for j, row := range rows {
		col := 0
		for _, r := range row {
			if r != ' ' {
				c.ink(mx-5+col, my-2+j, r, 230, false)
			}
			col++
		}
	}
}

func (g *game) drawClouds(c *canvas) {
	rt := g.roadTop()
	for _, cl := range g.clouds {
		x0 := int(math.Mod(cl.fx*float64(c.w)+g.t*cl.speed, float64(c.w+40))) - 20
		y0 := 3 + int(cl.fy*float64(maxInt(1, rt-13)))
		rows := []string{
			`  ▄▄▄▄  `,
			`▄██████▄`,
			` ▀▀▀▀▀ `,
		}
		for j, row := range rows {
			col := 0
			for _, r := range row {
				if r != ' ' {
					c.ink(x0+col, y0+j, r, 236, false)
				}
				col++
			}
		}
	}
}

func (g *game) drawSkyline(c *canvas) {
	rt := g.roadTop()
	const bw = 6
	off := int(g.distance * 0.12)
	for x := 0; x < c.w; x++ {
		wx := x + off
		bi := wx / bw
		h := 1 + int(hash2(bi, 7)%3)
		top := rt - h
		if top < 1 {
			top = 1
		}
		for yy := top; yy < rt; yy++ {
			c.put(x, yy, ' ', -1, 235)
		}
		if h >= 2 && wx%2 == 0 && hash2(bi, 31)%4 != 0 {
			c.ink(x, top+1, '·', 180, false)
		}
	}
}

func (g *game) drawRoad(c *canvas) {
	rt := g.roadTop()
	c.fill(0, rt, c.w, c.h-rt, ' ', -1, 237)

	for x := 0; x < c.w; x++ {
		c.put(x, rt, '▁', 240, 237)
	}

	off := int(g.distance)
	for x := 0; x < c.w; x++ {
		if ((x+off)%16+16)%16 < 7 {
			c.put(x, rt+2, '━', 178, 237)
		}
	}

	for y := rt + 1; y < c.h; y++ {
		if y == rt+2 {
			continue
		}
		for x := 0; x < c.w; x++ {
			h := hash2(x+int(g.distance*0.5), y*17+5)
			if h%41 == 0 {
				c.ink(x, y, '·', 240, false)
			}
		}
	}

	c.fill(0, c.h-1, c.w, 1, ' ', -1, 236)
}

func (g *game) drawPuddles(c *canvas) {
	for _, p := range g.puddles {
		g.drawPuddle(c, p)
	}
}

func (g *game) drawPuddle(c *canvas, p puddle) {
	rt := g.roadTop()
	x := int(math.Round(p.x))
	w := int(p.w)
	if w < 3 {
		w = 3
	}
	if x+w < 0 || x >= c.w {
		return
	}

	for i := 1; i < w-1; i++ {
		c.put(x+i, rt, '▄', 39, 237)
	}
	for i := 0; i < w; i++ {
		c.put(x+i, rt+1, '█', 25, 237)
	}
	c.put(x, rt+1, '▐', 31, 237)
	c.put(x+w-1, rt+1, '▌', 31, 237)

	shift := int(g.t * 9)
	for i := 1; i < w-1; i++ {
		if (i+shift)%4 == 0 {
			c.put(x+i, rt+1, '·', 51, 25)
		}
	}
	if w >= 5 {
		c.put(x+w/2, rt, '▄', 45, 237)
	}
}

var catIdle = []string{
	`  /\_/\  `,
	` ( o.o ) `,
	`  > ^ <  `,
	` /|   |\ `,
	`  ^   ^  `,
}

var catRunB = []string{
	`  /\_/\  `,
	` ( o.o ) `,
	`  > ^ <  `,
	`  |   |\ `,
	`  ^    ^ `,
}

var catJump = []string{
	`  /\_/\  `,
	` ( o.o ) `,
	`  > ^ <  `,
	`  /| |\  `,
	`  ^   ^  `,
}

var catScared = []string{
	`  /\_/\  `,
	` ( O.O ) `,
	`  > ^ <  `,
	` /|   |\ `,
	`  ^   ^  `,
}

var catOver = []string{
	`  /\_/\  `,
	` ( x.x ) `,
	`  > ^ <  `,
	` /|   |\ `,
	`  ^   ^  `,
}

func (g *game) drawSausage(c *canvas) {
	if g.st != stateFinish && g.st != stateVictory {
		return
	}
	rt := g.roadTop()

	px := int(math.Round(g.plateX))
	plateW := 16
	for i := 1; i < plateW-1; i++ {
		c.put(px+i, rt+1, '▄', 253, 237)
		c.put(px+i, rt+2, '▀', 250, 237)
	}
	c.put(px, rt+1, '▖', 253, 237)
	c.put(px+plateW-1, rt+1, '▗', 253, 237)
	c.ink(px+3, rt+1, '·', 70, false)
	c.ink(px+plateW-4, rt+1, '·', 70, false)

	w := int(math.Round(g.sausageW))
	if w < 4 {
		w = 4
	}
	x := int(math.Round(g.sausageX))

	if w > 5 && int(g.t*3)%2 == 0 {
		c.ink(x+2, rt-5, '~', 245, false)
		c.ink(x+6, rt-6, '~', 245, false)
	}

	// rounded sausage: narrow top, fat middle, narrow bottom
	for i := 2; i < w-2; i++ {
		c.ink(x+i, rt-2, '▄', 174, false)
	}
	for i := 1; i < w-1; i++ {
		c.ink(x+i, rt, '▀', 95, false)
	}
	for i := 0; i < w; i++ {
		c.ink(x+i, rt-1, '█', 131, false)
	}
	c.ink(x, rt-1, '▐', 95, false)
	c.ink(x+w-1, rt-1, '▌', 95, false)
	c.ink(x+2, rt-2, '▀', 180, false)
	c.ink(x+w-3, rt-2, '▀', 180, false)

	if g.bites > 0 {
		c.ink(x, rt-2, '▙', 180, false)
		c.ink(x, rt-1, '▛', 180, false)
		c.ink(x, rt, '▙', 180, false)
	}

	if g.st == stateVictory {
		baseX := int(math.Round(g.catX))
		for i := 0; i < 3; i++ {
			hy := rt - 7 - int(math.Round(math.Sin(g.t*2+float64(i))))
			c.ink(baseX-2+i*4, hy, '♥', 197, true)
		}
		for i := 0; i < 3; i++ {
			c.ink(px+3+i*3, rt, '·', 180, false)
		}
	}
}

type catPalette struct {
	body int
	ear  int
	eye  int
	nose int
}

var orangeCat = catPalette{body: 208, ear: 214, eye: 231, nose: 217}
var grayCat = catPalette{body: 250, ear: 249, eye: 39, nose: 218}

var catSleep = []string{
	`  /\_/\  `,
	` ( -.- ) `,
	`  >   <  `,
	` /_____\ `,
}

func (g *game) drawCatSprite(c *canvas, x, base int, rows []string, pal catPalette, blink, decor bool) {
	for j, line := range rows {
		yy := base - (len(rows) - 1) + j
		col := 0
		for _, r := range line {
			if r != ' ' {
				if r == 'o' && blink {
					r = '-'
				}
				colr := pal.body
				switch r {
				case 'o', 'O', 'x', '-':
					colr = pal.eye
				case '^':
					colr = pal.nose
				case '/', '\\', '_':
					colr = pal.ear
				}
				c.ink(x+col, yy, r, colr, true)
			}
			col++
		}
	}
	if decor {
		c.ink(x+catWidth, base-2, '~', pal.body, false)
		c.ink(x+catWidth+1, base-3, '~', pal.body, false)
		c.ink(x-1, base-3, '─', 223, false)
		c.ink(x+catWidth-1, base-3, '─', 223, false)
	}
}

func (g *game) drawCat(c *canvas) {
	if g.catEaten {
		return
	}
	if g.st != stateOver && g.invuln > 0 && int(g.t*16)%2 == 0 {
		return
	}

	frame := int(g.runTime*9) % 2
	bob := 0
	rows := catIdle

	switch {
	case g.st == stateDog:
		rows = catScared
	case g.st == stateOver:
		rows = catOver
	case g.st == stateReady:
		bob = int(math.Round((math.Sin(g.t*2.4) + 1) * 0.5))
	case !g.onGround:
		rows = catJump
	case frame == 1:
		rows = catRunB
		bob = 1
	}

	base := g.roadTop() - int(math.Round(g.catY)) - bob
	x := int(math.Round(g.catX))
	if g.st == stateDog {
		x += int(math.Round(math.Sin(g.t * 22)))
	}
	blink := g.st != stateOver && math.Mod(g.t+2.0, 4.1) < 0.1

	// Shadow on the asphalt, shrinking as the cat jumps.
	sw := int(math.Round(9 - g.catY*1.2))
	if sw < 2 {
		sw = 2
	}
	sx := x + (catWidth-sw)/2
	for i := 0; i < sw; i++ {
		if (i+int(g.t*4))%3 != 0 {
			c.ink(sx+i, g.roadTop()+1, '▁', 234, false)
		}
	}

	g.drawCatSprite(c, x, base, rows, orangeCat, blink, true)
}

func (g *game) drawCatShadow(c *canvas, x float64) {
	ix := int(math.Round(x))
	for i := 0; i < 7; i++ {
		if (i+int(g.t*4))%3 != 0 {
			c.ink(ix+1+i, g.roadTop()+1, '▁', 234, false)
		}
	}
}

func (g *game) drawSleepingCat(c *canvas, x, base int, pal catPalette) {
	for j, line := range catSleep {
		yy := base - (len(catSleep) - 1) + j
		col := 0
		for _, r := range line {
			if r != ' ' {
				colr := pal.body
				switch r {
				case '-':
					colr = pal.eye
				case '/', '\\', '_':
					colr = pal.ear
				}
				c.ink(x+col, yy, r, colr, true)
			}
			col++
		}
	}
}

// Одна подушка: овал с мягкой окантовкой, на котором лежит котик.
func (g *game) drawPillow(c *canvas, x, w int) {
	rt := g.roadTop()
	top := rt - 2

	c.put(x+1, top, '▖', 253, 237)
	c.put(x+w-2, top, '▗', 253, 237)
	for i := 2; i < w-2; i++ {
		c.put(x+i, top, '▄', 253, 237)
	}

	for i := 0; i < w; i++ {
		c.put(x+i, top+1, '█', 230, 237)
	}
	c.put(x+3, top+1, '▓', 223, 237)
	c.put(x+4, top+1, '▓', 223, 237)

	c.put(x+1, top+2, '▘', 181, 237)
	c.put(x+w-2, top+2, '▝', 181, 237)
	for i := 2; i < w-2; i++ {
		c.put(x+i, top+2, '▀', 181, 237)
	}
}

// Две подушки — по одной на котика.
func (g *game) drawPillows(c *canvas) {
	rt := g.roadTop()
	x := int(math.Round(g.pillowX))
	for _, px := range []int{x, x + pillowW + pillowGap} {
		g.drawPillow(c, px, pillowW)
		for i := 2; i < pillowW-2; i++ {
			c.ink(px+i, rt+1, '▁', 236, false)
		}
	}
}

func (g *game) drawZzz(c *canvas) {
	rt := g.roadTop()
	for _, cat := range []float64{g.catX, g.cat2X} {
		x := int(math.Round(cat))
		for i := 0; i < 2; i++ {
			zx := x + 6 + i + int(math.Round(math.Sin(g.t*1.8+float64(i))*1))
			zy := rt - 7 - i*2 - int(g.t*2)%2
			c.inkText(zx, zy, "z", 246, true)
		}
	}
	for i := 0; i < 2; i++ {
		hx := int(math.Round(g.pillowX)) + pillowW + pillowGap/2 + i*2
		hy := rt - 7 - i
		c.ink(hx, hy, '♥', 197, true)
	}
}

func (g *game) drawFinale(c *canvas) {
	rt := g.roadTop()
	if g.pillowShown {
		g.drawPillows(c)
		if g.asleep {
			g.drawSleepingCat(c, int(math.Round(g.cat2X)), rt-3, grayCat)
			g.drawSleepingCat(c, int(math.Round(g.catX)), rt-3, orangeCat)
			g.drawZzz(c)
		} else {
			g.drawCatShadow(c, g.cat2X)
			g.drawCatShadow(c, g.catX)
			g.drawCatSprite(c, int(math.Round(g.cat2X)), rt, catIdle, grayCat, false, false)
			g.drawCatSprite(c, int(math.Round(g.catX)), rt, catIdle, orangeCat, false, true)
		}
		return
	}

	g.drawSausage(c)
	frame := int(g.runTime*9) % 2
	base := rt
	if g.bites > 0 && g.biteTimer < 0.2 {
		base--
	}

	rows1 := catIdle
	if !g.cat2Active && frame == 1 {
		rows1 = catRunB
	}
	blink1 := math.Mod(g.t+2.0, 4.1) < 0.1
	g.drawCatShadow(c, g.catX)
	g.drawCatSprite(c, int(math.Round(g.catX)), base, rows1, orangeCat, blink1, true)

	if g.cat2Active {
		rows2 := catIdle
		if g.cat2X < g.catX-catWidth-2.5 && frame == 1 {
			rows2 = catRunB
		}
		blink2 := math.Mod(g.t+3.1, 4.1) < 0.1
		g.drawCatShadow(c, g.cat2X)
		g.drawCatSprite(c, int(math.Round(g.cat2X)), base, rows2, grayCat, blink2, false)
	}
}

var dogRunA = []string{
	`  ▐▄▄▄▄▄▌  `,
	`  ████████  `,
	`  █ o  o █  `,
	`  █ ▄██▄ █  `,
	`   ▀████▀   `,
	`   █    █   `,
}

var dogRunB = []string{
	`  ▐▄▄▄▄▄▌  `,
	`  ████████  `,
	`  █ o  o █  `,
	`  █ ▄██▄ █  `,
	`   ▀████▀   `,
	`   █   █    `,
}

var dogSit = []string{
	`  ▐▄▄▄▄▄▌  `,
	`  ████████  `,
	`  █ -  - █  `,
	`  █ ▄██▄ █  `,
	`   ▀████▀   `,
	`   █    █   `,
}

func dogInk(r rune) int {
	switch r {
	case 'o', '-', 'x':
		return 231
	case '▐', '▌':
		return 94
	default:
		return 130
	}
}

func (g *game) drawDog(c *canvas) {
	if !g.dogActive {
		return
	}
	rt := g.roadTop()
	x := int(math.Round(g.dogX))
	base := rt
	rows := dogRunA
	if g.catEaten {
		rows = dogSit
	} else {
		if int(g.lossFor*12)%2 == 1 {
			rows = dogRunB
		}
		base -= int(math.Round(math.Abs(math.Sin(g.lossFor*9)) * 0.7))
	}

	for i := 0; i < 9; i++ {
		if (i+int(g.t*4))%3 != 0 {
			c.ink(x+2+i, rt+1, '▁', 234, false)
		}
	}
	for j, line := range rows {
		yy := base - (len(rows) - 1) + j
		col := 0
		for _, r := range line {
			if r != ' ' {
				c.ink(x+col, yy, r, dogInk(r), true)
			}
			col++
		}
	}
	c.ink(x+12, base-3, '~', 130, false)
	c.ink(x+13, base-4, '~', 130, false)

	if g.catEaten {
		c.put(x+6, base-3, 'w', 203, 130)
		c.inkText(x+2, base-7, "НЯМ!", 220, true)
	} else {
		c.inkText(x+2, base-7, "ГАВ!", 220, true)
	}
}

func (g *game) drawEffects(c *canvas) {
	for _, p := range g.particles {
		if p.life <= 0 {
			continue
		}
		col := p.color
		if p.life/p.maxLife < 0.35 {
			col = 240
		}
		c.ink(int(math.Round(p.x)), int(math.Round(p.y)), p.ch, col, false)
	}

	for _, f := range g.floats {
		if f.life <= 0 {
			continue
		}
		col := f.color
		if f.life/f.maxLife < 0.3 {
			col = 247
		}
		x := int(math.Round(f.x)) - runeLen(f.text)/2
		c.inkText(x, int(math.Round(f.y)), f.text, col, true)
	}
}

func (g *game) drawHUD(c *canvas) {
	bg := 17
	if g.flash > 0.55 {
		bg = 52
	}
	c.fill(0, 0, c.w, 2, ' ', -1, bg)

	score := fmt.Sprintf("ОЧКИ %d", g.score)
	c.inkText(2, 0, score, 255, true)

	for i := 0; i < maxLives; i++ {
		ch := '♥'
		col := 197
		if i >= g.lives {
			ch = '♡'
			col = 240
		}
		c.ink(c.w/2-2+i*2, 0, ch, col, true)
	}

	stage := g.stageLabel()
	c.inkText(c.w-runeLen(stage)-2, 0, stage, 246, false)

	label := "ДО СОСИСКИ"
	progress := g.overallProgress()
	pct := fmt.Sprintf("%d%%", int(math.Round(progress*100)))

	barW := minInt(40, c.w-26)
	if barW < 10 {
		barW = 10
	}
	groupW := runeLen(label) + 2 + barW + 2 + runeLen(pct)
	x := (c.w - groupW) / 2
	if x < 2 {
		x = 2
		barW = c.w - x - runeLen(label) - 2 - runeLen(pct) - 2
		if barW < 8 {
			barW = 8
		}
	}
	barX := x + runeLen(label) + 2
	if barX+barW > c.w-2 {
		barW = c.w - 2 - barX
	}
	if barW < 1 {
		barW = 1
	}

	c.inkText(x, 1, label, 247, false)
	filled := int(math.Round(progress * float64(barW)))
	if filled > barW {
		filled = barW
	}
	for i := 0; i < barW; i++ {
		if i < filled {
			c.ink(barX+i, 1, '━', 178, false)
		} else {
			c.ink(barX+i, 1, '─', 238, false)
		}
	}
	pctX := barX + barW + 2
	if pctX+runeLen(pct) > c.w-1 {
		pctX = c.w - 1 - runeLen(pct)
	}
	c.inkText(pctX, 1, pct, 246, false)
}

// ---------- overlays ----------

type panelLine struct {
	text  string
	color int
	bold  bool
}

func drawPanel(c *canvas, title string, body []panelLine, footer string, footerColor int) {
	drawPanelAt(c, title, body, footer, footerColor, -1)
}

func drawPanelAt(c *canvas, title string, body []panelLine, footer string, footerColor, forcedY int) {
	inner := runeLen(title)
	for _, l := range body {
		if runeLen(l.text) > inner {
			inner = runeLen(l.text)
		}
	}
	if runeLen(footer) > inner {
		inner = runeLen(footer)
	}

	pw := inner + 6
	ph := len(body) + 6
	if pw > c.w-2 {
		pw = c.w - 2
	}
	if ph > c.h-2 {
		ph = c.h - 2
	}
	if pw < 8 {
		pw = minInt(8, c.w)
	}
	if ph < 5 {
		ph = minInt(5, c.h)
	}
	x := (c.w - pw) / 2
	y := (c.h - ph) / 2
	if forcedY >= 0 {
		y = forcedY
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if y+ph > c.h {
		y = c.h - ph
	}

	c.fill(x+1, y+1, pw, ph, ' ', -1, 233)
	c.fill(x, y, pw, ph, ' ', -1, 235)

	for i := 0; i < pw; i++ {
		c.put(x+i, y, '─', 244, 235)
		c.put(x+i, y+ph-1, '─', 244, 235)
	}
	for j := 0; j < ph; j++ {
		c.put(x, y+j, '│', 244, 235)
		c.put(x+pw-1, y+j, '│', 244, 235)
	}
	c.put(x, y, '┌', 244, 235)
	c.put(x+pw-1, y, '┐', 244, 235)
	c.put(x, y+ph-1, '└', 244, 235)
	c.put(x+pw-1, y+ph-1, '┘', 244, 235)

	c.inkText(x+(pw-runeLen(title))/2, y+1, title, 220, true)

	for i, l := range body {
		yy := y + 3 + i
		if yy >= y+ph-2 {
			break
		}
		xx := x + (pw-runeLen(l.text))/2
		c.inkText(xx, yy, l.text, l.color, l.bold)
	}

	if footer != "" {
		c.inkText(x+(pw-runeLen(footer))/2, y+ph-2, footer, footerColor, true)
	}
}

func (g *game) drawOverlay(c *canvas) {
	switch g.st {
	case stateReady:
		body := []panelLine{
			{text: "Прыгай через лужи и не промокни.", color: 252},
			{text: "Два этапа, три жизни — и сосиска в финале.", color: 252},
			{text: ""},
			{text: "ПРОБЕЛ / W / ↑ — прыжок", color: 247},
			{text: "P — пауза        Q — выход", color: 247},
		}
		footer := ""
		if int(g.t*2.2)%2 == 0 {
			footer = "Нажми ПРОБЕЛ, чтобы начать"
		}
		drawPanel(c, "КОТИК-БЕГУН", body, footer, 220)

	case statePaused:
		body := []panelLine{
			{text: "P — продолжить", color: 252, bold: true},
		}
		drawPanel(c, "ПАУЗА", body, "Q — выход", 220)

	case stateStageClear:
		body := []panelLine{
			{text: "+100 очков", color: 220, bold: true},
			{text: "Остался последний рывок.", color: 252},
			{text: "Впереди — сосиска.", color: 252},
		}
		footer := ""
		if int(g.t*2.2)%2 == 0 {
			footer = "ПРОБЕЛ — дальше"
		}
		drawPanel(c, "ЭТАП 1 ПРОЙДЕН!", body, footer, 220)

	case stateVictory:
		body := []panelLine{
			{text: "Два котика наелись и уснули — каждый на своей подушке.", color: 252},
			{text: fmt.Sprintf("Очки: %d   Рекорд: %d", g.score, g.best), color: 255, bold: true},
		}
		if g.newBest {
			body = append(body, panelLine{text: "НОВЫЙ РЕКОРД!", color: 220, bold: true})
		} else {
			body = append(body, panelLine{text: "Сладких снов!", color: 247})
		}
		footer := "R / ПРОБЕЛ — ещё раз     Q — выход"
		drawPanelAt(c, "СОСИСКА СЪЕДЕНА!", body, footer, 220, 2)

	case stateOver:
		title := "ИГРА ОКОНЧЕНА"
		body := []panelLine{
			{text: fmt.Sprintf("Очки: %d", g.score), color: 255, bold: true},
			{text: fmt.Sprintf("Этап: %d / %d", g.level, maxStages), color: 247},
			{text: fmt.Sprintf("Рекорд: %d", g.best), color: 247},
		}
		if g.dogActive {
			title = "КОТИКА СЪЕЛИ"
			body = []panelLine{
				{text: "Собака догнала котика и съела его.", color: 252},
				{text: fmt.Sprintf("Очки: %d   Этап: %d/%d   Рекорд: %d", g.score, g.level, maxStages, g.best), color: 255, bold: true},
			}
		}
		if g.newBest {
			body = append(body, panelLine{text: "НОВЫЙ РЕКОРД!", color: 220, bold: true})
		} else if g.dogActive {
			body = append(body, panelLine{text: "Собака довольно облизывается.", color: 247})
		} else {
			body = append(body, panelLine{text: "Ещё одна попытка?", color: 247})
		}
		footer := ""
		if g.overFor > 0.45 {
			footer = "R / ПРОБЕЛ — заново     Q — выход"
		}
		drawPanelAt(c, title, body, footer, 203, 2)
	}
}

func (g *game) renderFrame() string {
	if g.tooSmall() {
		return renderTooSmall(g.w, g.h)
	}
	c := newCanvas(g.w, g.h)
	g.drawSky(c)
	g.drawClouds(c)
	g.drawSkyline(c)
	g.drawRoad(c)
	g.drawPuddles(c)
	switch g.st {
	case stateFinish, stateVictory:
		g.drawFinale(c)
	case stateDog:
		g.drawCat(c)
		g.drawDog(c)
	case stateOver:
		if g.dogActive {
			g.drawDog(c)
		}
		g.drawCat(c)
	default:
		g.drawCat(c)
	}
	g.drawEffects(c)
	g.drawHUD(c)
	g.drawOverlay(c)
	return c.render()
}

func renderTooSmall(w, h int) string {
	c := newCanvas(w, h)
	c.fill(0, 0, w, h, ' ', -1, 234)
	msg := "Окно слишком маленькое"
	sub := fmt.Sprintf("Нужно хотя бы 64×20, сейчас %d×%d", w, h)
	c.inkTextCentered(h/2, msg, 255, true)
	c.inkTextCentered(h/2+1, sub, 247, false)
	return c.render()
}

func snapshotFrame(w, h int, screen string) string {
	g := newGame(w, h, 980)
	g.t = 1.4
	if screen == "start" {
		g.st = stateReady
		g.score = 0
		g.best = 980
		g.t = 1.1
		return g.renderFrame()
	}
	if screen == "dog" || screen == "over" {
		g.st = stateDog
		g.dogActive = true
		g.catEaten = false
		g.score = 1240
		g.best = 980
		g.lives = 0
		g.level = 1
		g.stageDistance = 420
		g.stageGoal = stageGoals[0]
		g.dogX = g.catX + 13
		g.lossFor = 0.7
		g.catY = 0
		g.onGround = true
		if screen == "over" {
			g.st = stateOver
			g.catEaten = true
			g.dogX = g.catX - 3
			g.overFor = 1.2
			g.newBest = true
		}
		return g.renderFrame()
	}
	if screen == "stageclear" {
		g.st = stateStageClear
		g.level = 1
		g.stageDistance = stageGoals[0]
		g.score = 1520
		g.best = 2100
		g.clearFor = 0.2
		g.catY = 0
		g.onGround = true
		g.runTime = 2.0
		return g.renderFrame()
	}
	if screen == "finish" {
		g.st = stateFinish
		g.level = 2
		g.stageGoal = stageGoals[1]
		g.stageDistance = stageGoals[1]
		g.score = 2480
		g.best = 2100
		g.points = 2480
		g.lives = 3
		g.sausageW = 10
		g.sausageX = float64(w) - 26
		g.plateX = g.sausageX - 2
		g.catX = g.sausageX - catWidth
		g.cat2Active = true
		g.cat2Right = true
		g.cat2X = g.sausageX + g.sausageW + 2
		g.bites = 1
		g.biteTimer = 0.1
		g.catY = 0
		g.onGround = true
		g.runTime = 1.3
		g.floats = []floatText{
			{x: g.catX - 4, y: float64(g.roadTop() - 8), text: "НЯМ!", life: 0.7, maxLife: 0.7, color: 220},
			{x: g.sausageX + 5, y: float64(g.roadTop() - 6), text: "СОСИСКА!", life: 1.2, maxLife: 1.8, color: 220},
		}
		return g.renderFrame()
	}
	if screen == "victory" {
		g.st = stateVictory
		g.level = 2
		g.stageGoal = stageGoals[1]
		g.stageDistance = stageGoals[1]
		g.score = 2730
		g.best = 2730
		g.points = 2730
		g.lives = 3
		g.newBest = true
		g.victoryFor = 1.4
		g.pillowShown = true
		g.asleep = true
		g.cat2Active = true
		g.pillowX = float64(w) - 34
		g.cat2X = g.pillowX + pillowW + pillowGap + 1
		g.catX = g.pillowX + 1
		g.catY = 0
		g.onGround = true
		g.floats = []floatText{
			{x: g.pillowX + 14, y: float64(g.roadTop() - 9), text: "МУР-Р!", life: 1.4, maxLife: 2.0, color: 220},
		}
		return g.renderFrame()
	}

	g.st = stateRunning
	g.runTime = 12.3
	g.t = 1.4
	g.level = 2
	g.stageGoal = stageGoals[1]
	g.speed = 30
	g.stageDistance = 420
	g.distance = 1234
	g.points = 1240
	g.score = 1240
	g.lives = 3
	g.catY = 3.0
	g.onGround = false
	g.vy = 2.5
	g.spawnIn = 99
	g.puddles = []puddle{
		{x: g.catX - 2, w: 7, passed: true},
		{x: g.catX + 38, w: 6},
		{x: g.catX + 64, w: 7},
		{x: g.catX + 88, w: 5},
	}
	g.floats = []floatText{
		{x: g.catX + 16, y: float64(g.roadTop() - 7), text: "+15", life: 0.6, maxLife: 0.75, color: 220},
	}
	g.particles = []particle{
		{x: g.catX + 8, y: float64(g.roadTop()) - 1, vx: 1, vy: -1, life: 0.5, maxLife: 0.6, ch: '·', color: 51},
		{x: g.catX + 10, y: float64(g.roadTop()) - 2, vx: 1, vy: -1, life: 0.5, maxLife: 0.6, ch: '*', color: 231},
	}
	return g.renderFrame()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
