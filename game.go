package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type gameState int

const (
	stateReady gameState = iota
	stateRunning
	statePaused
	stateStageClear
	stateFinish
	stateVictory
	stateOver
	stateDog
)

const (
	maxLives        = 3
	maxStages       = 2
	jumpSpeed       = 26.0
	gravity         = 60.0
	baseSpeed       = 17.0
	topSpeed        = 40.0
	catWidth        = 9
	pillowW         = 11
	pillowGap       = 3
	catHitbox       = 5.5
	roadRows        = 5
	clearBonus      = 15
	stageClearBonus = 100
	prizeBonus      = 250
)

var (
	stageGoals      = [maxStages]float64{650, 850}
	stageBaseSpeeds = [maxStages]float64{baseSpeed, 26}
	stageMaxSpeeds  = [maxStages]float64{30, topSpeed}
	stageRamps      = [maxStages]float64{0.32, 0.48}
)

type puddle struct {
	x, w   float64
	hit    bool
	passed bool
}

type particle struct {
	x, y    float64
	vx, vy  float64
	life    float64
	maxLife float64
	ch      rune
	color   int
	gravity float64
}

type floatText struct {
	x, y    float64
	text    string
	life    float64
	maxLife float64
	color   int
}

type cloud struct {
	fx, fy float64
	speed  float64
}

type game struct {
	w, h int
	st   gameState

	t       float64
	runTime float64
	overFor float64

	level         int
	stageTime     float64
	stageDistance float64
	stageGoal     float64

	catX     float64
	catY     float64
	vy       float64
	onGround bool

	speed    float64
	distance float64
	points   float64
	score    int
	best     int
	lives    int
	invuln   float64
	flash    float64
	newBest  bool
	persist  bool

	puddles   []puddle
	particles []particle
	floats    []floatText
	spawnIn   float64
	spawned   int

	clearFor    float64
	finishFor   float64
	finishDone  float64
	bites       int
	biteTimer   float64
	sausageX    float64
	sausageW    float64
	plateX      float64
	victoryFor  float64
	cat2Active  bool
	cat2Right   bool
	cat2X       float64
	pillowShown bool
	asleep      bool
	pillowX     float64
	sleepFor    float64

	dogActive bool
	catEaten  bool
	dogX      float64
	lossFor   float64
	lossDone  float64

	clouds []cloud
	rng    *rand.Rand
}

func newGame(w, h, best int) *game {
	g := &game{
		w:    w,
		h:    h,
		best: best,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.reset()
	return g
}

func (g *game) reset() {
	g.st = stateReady
	g.runTime = 0
	g.overFor = 0

	g.level = 1
	g.stageTime = 0
	g.stageDistance = 0
	g.stageGoal = stageGoals[0]

	g.catX = clampF(float64(g.w)*0.14, 8, 20)
	g.catY = 0
	g.vy = 0
	g.onGround = true

	g.speed = stageBaseSpeeds[0]
	g.distance = 0
	g.points = 0
	g.score = 0
	g.lives = maxLives
	g.invuln = 0
	g.flash = 0
	g.newBest = false
	g.clearFor = 0
	g.finishFor = 0
	g.finishDone = 0
	g.bites = 0
	g.biteTimer = 0
	g.sausageX = 0
	g.sausageW = 0
	g.plateX = 0
	g.victoryFor = 0
	g.cat2Active = false
	g.cat2Right = false
	g.cat2X = -catWidth - 2
	g.pillowShown = false
	g.asleep = false
	g.pillowX = 0
	g.sleepFor = 0
	g.dogActive = false
	g.catEaten = false
	g.dogX = 0
	g.lossFor = 0
	g.lossDone = 0

	g.puddles = nil
	g.particles = nil
	g.floats = nil
	g.spawnIn = 1.2
	g.spawned = 0

	g.clouds = []cloud{
		{fx: 0.18, fy: 0.24, speed: 1.1},
		{fx: 0.55, fy: 0.52, speed: 0.8},
		{fx: 0.82, fy: 0.10, speed: 1.6},
	}
}

func (g *game) startRun() {
	g.st = stateRunning
	g.runTime = 0
	g.overFor = 0
	g.victoryFor = 0

	g.level = 1
	g.stageTime = 0
	g.stageDistance = 0
	g.stageGoal = stageGoals[0]

	g.catX = clampF(float64(g.w)*0.14, 8, 20)
	g.catY = 0
	g.vy = 0
	g.onGround = true

	g.speed = stageBaseSpeeds[0]
	g.distance = 0
	g.points = 0
	g.score = 0
	g.lives = maxLives
	g.invuln = 0
	g.flash = 0
	g.newBest = false
	g.clearFor = 0
	g.finishFor = 0
	g.finishDone = 0
	g.bites = 0
	g.biteTimer = 0
	g.sausageW = 0
	g.victoryFor = 0
	g.cat2Active = false
	g.cat2Right = false
	g.cat2X = -catWidth - 2
	g.pillowShown = false
	g.asleep = false
	g.pillowX = 0
	g.sleepFor = 0
	g.dogActive = false
	g.catEaten = false
	g.dogX = 0
	g.lossFor = 0
	g.lossDone = 0

	g.puddles = g.puddles[:0]
	g.particles = g.particles[:0]
	g.floats = g.floats[:0]

	g.spawnOpeningPuddles()
	g.jump()
}

func (g *game) startStage(level int) {
	if level < 1 {
		level = 1
	}
	if level > maxStages {
		level = maxStages
	}

	g.st = stateRunning
	g.level = level
	g.stageTime = 0
	g.stageDistance = 0
	g.stageGoal = stageGoals[level-1]
	g.speed = stageBaseSpeeds[level-1]
	g.clearFor = 0

	g.catX = clampF(float64(g.w)*0.14, 8, 20)
	g.catY = 0
	g.vy = 0
	g.onGround = true
	g.invuln = 0
	g.flash = 0

	g.puddles = g.puddles[:0]
	g.particles = g.particles[:0]
	g.floats = g.floats[:0]

	g.spawnOpeningPuddles()
	g.jump()
}

func (g *game) spawnOpeningPuddles() {
	g.puddles = append(g.puddles, puddle{
		x: g.catX + g.speed*2.4,
		w: 4 + float64(g.level-1),
	})
	g.spawned = 1
	g.spawnPuddle()
	g.spawnIn = 1.6
}

func (g *game) update(dt float64) {
	if dt <= 0 {
		return
	}
	if dt > 0.06 {
		dt = 0.06
	}

	g.t += dt
	if g.flash > 0 {
		g.flash = math.Max(0, g.flash-dt*2.4)
	}

	switch g.st {
	case statePaused:
		return

	case stateReady:
		g.runTime += dt
		g.updateEffects(dt)

	case stateRunning:
		g.runTime += dt
		g.stageTime += dt
		idx := g.level - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= maxStages {
			idx = maxStages - 1
		}
		g.speed = math.Min(stageMaxSpeeds[idx], stageBaseSpeeds[idx]+g.stageTime*stageRamps[idx])
		g.stageDistance += g.speed * dt
		g.distance += g.speed * dt
		g.points += g.speed * dt * 1.25
		g.score = int(g.points)

		if g.invuln > 0 {
			g.invuln = math.Max(0, g.invuln-dt)
		}

		g.spawnIn -= dt
		if g.spawnIn <= 0 {
			g.spawnPuddle()
			g.spawnIn = g.nextSpawnDelay()
		}

		if !g.onGround {
			g.vy -= gravity * dt
			g.catY += g.vy * dt
			if g.catY <= 0 {
				g.catY = 0
				g.vy = 0
				g.onGround = true
				g.dust()
			}
		}

		for i := range g.puddles {
			g.puddles[i].x -= g.speed * dt
		}
		live := g.puddles[:0]
		for _, p := range g.puddles {
			if p.x+p.w > -6 {
				live = append(live, p)
			}
		}
		g.puddles = live

		g.collide()
		g.updateEffects(dt)

		if g.st == stateRunning && g.stageDistance >= g.stageGoal {
			if g.level < maxStages {
				g.beginStageClear()
			} else {
				g.beginFinish()
			}
		}

	case stateStageClear:
		g.clearFor += dt
		g.updateEffects(dt)
		if g.clearFor >= 2.2 {
			g.startStage(2)
		}

	case stateFinish:
		g.updateFinish(dt)

	case stateDog:
		g.updateDog(dt)

	case stateVictory:
		g.victoryFor += dt
		g.updateEffects(dt * 0.6)

	case stateOver:
		g.overFor += dt
		g.updateEffects(dt * 0.6)
	}
}

func (g *game) nextSpawnDelay() float64 {
	if g.level >= 2 {
		return 1.05 + g.rng.Float64()*0.75
	}
	return 1.2 + g.rng.Float64()*0.85
}

func (g *game) spawnPuddle() {
	w := 3 + g.rng.Float64()*3
	w += math.Min(2.2, g.stageTime*0.05)
	w += float64(g.level-1) * 0.6
	w = math.Min(9, w)

	g.puddles = append(g.puddles, puddle{
		x: float64(g.w) + 2,
		w: math.Round(w),
	})
	g.spawned++
}

func (g *game) jump() {
	if !g.onGround {
		return
	}
	g.onGround = false
	g.vy = jumpSpeed
	g.dust()
}

func (g *game) collide() {
	catL := g.catX + 2.5
	catR := catL + catHitbox

	for i := range g.puddles {
		p := &g.puddles[i]
		if p.passed {
			continue
		}
		if p.x+p.w < catL {
			p.passed = true
			if !p.hit {
				g.points += clearBonus
				g.score = int(g.points)
				g.floats = append(g.floats, floatText{
					x:       p.x + p.w/2,
					y:       float64(g.roadTop() - 3),
					text:    "+15",
					life:    0.75,
					maxLife: 0.75,
					color:   220,
				})
				g.sparks(p.x+p.w/2, float64(g.roadTop()))
			}
			continue
		}
		if p.hit || catR <= p.x || catL >= p.x+p.w {
			continue
		}
		if g.catY > 1.45 {
			continue
		}
		if g.invuln > 0 {
			p.hit = true
			continue
		}
		g.takeHit(p)
		if g.st != stateRunning {
			return
		}
	}
}

func (g *game) takeHit(p *puddle) {
	p.hit = true
	g.lives--
	g.invuln = 1.5
	g.flash = 1
	g.splash(p.x+p.w/2, float64(g.roadTop())+0.2)
	g.floats = append(g.floats, floatText{
		x:       g.catX + catWidth/2,
		y:       float64(g.roadTop() - 6),
		text:    "-1 ♥",
		life:    1.0,
		maxLife: 1.0,
		color:   203,
	})
	if g.lives <= 0 {
		g.beginDogAttack()
	}
}

func (g *game) beginStageClear() {
	g.st = stateStageClear
	g.clearFor = 0
	g.points += stageClearBonus
	g.score = int(g.points)
	g.puddles = nil
	g.particles = nil
	g.floats = nil
	g.invuln = 0
	g.flash = 0
	g.catY = 0
	g.vy = 0
	g.onGround = true
}

func (g *game) beginFinish() {
	g.st = stateFinish
	g.finishFor = 0
	g.finishDone = 0
	g.bites = 0
	g.biteTimer = 0
	g.finishDone = 0
	g.cat2Active = false
	g.cat2Right = false
	g.cat2Right = false
	g.cat2X = -catWidth - 2
	g.pillowShown = false
	g.asleep = false
	g.pillowX = 0
	g.sleepFor = 0
	g.sausageW = 12
	g.sausageX = math.Min(float64(g.w)-14, g.catX+58)
	if g.sausageX < g.catX+30 {
		g.sausageX = g.catX + 30
	}
	g.plateX = g.sausageX - 2

	g.points += prizeBonus
	g.score = int(g.points)

	g.puddles = nil
	g.particles = nil
	g.floats = nil
	g.catY = 0
	g.vy = 0
	g.onGround = true
	g.invuln = 0
	g.flash = 0
	g.spawnIn = 999

	g.floats = append(g.floats, floatText{
		x:       g.sausageX + 5,
		y:       float64(g.roadTop() - 6),
		text:    "СОСИСКА!",
		life:    1.8,
		maxLife: 1.8,
		color:   220,
	})
}

func (g *game) updateFinish(dt float64) {
	g.finishFor += dt
	g.runTime += dt
	rt := g.roadTop()

	if g.pillowShown {
		if !g.asleep {
			target1 := g.pillowX + 1
			target2 := g.pillowX + pillowW + pillowGap + 1
			g.catX = moveToward(g.catX, target1, 38*dt)
			g.cat2X = moveToward(g.cat2X, target2, 38*dt)
			if math.Abs(g.catX-target1) < 0.1 && math.Abs(g.cat2X-target2) < 0.1 {
				g.catX = target1
				g.cat2X = target2
				g.asleep = true
				g.sleepFor = 0
				g.floats = append(g.floats, floatText{
					x:       g.pillowX + pillowW + pillowGap/2,
					y:       float64(rt - 8),
					text:    "МУР-МУР!",
					life:    1.6,
					maxLife: 1.6,
					color:   220,
				})
			}
		} else {
			g.sleepFor += dt
			if g.sleepFor >= 1.8 {
				g.finishRun()
			}
		}
		g.updateEffects(dt)
		return
	}

	target := g.sausageX - catWidth

	if !g.cat2Active {
		if g.catX < target-0.05 {
			g.catX = math.Min(target, g.catX+42*dt)
			if g.rng.Float64() < 0.18 {
				g.dust()
			}
			g.updateEffects(dt)
			return
		}
		g.catX = target
		g.cat2Active = true
		g.cat2Right = g.sausageX+g.sausageW+2+catWidth <= float64(g.w)-1
		meowX := 2.0
		if g.cat2Right {
			g.cat2X = float64(g.w) + 2
			meowX = float64(g.w - 10)
		} else {
			g.cat2X = -catWidth - 2
		}
		g.floats = append(g.floats, floatText{
			x:       meowX,
			y:       float64(rt - 7),
			text:    "МЯУ!",
			life:    1.2,
			maxLife: 1.2,
			color:   203,
		})
		g.updateEffects(dt)
		return
	}

	cat2Target := g.catX - catWidth - 2
	if g.cat2Right {
		cat2Target = g.sausageX + g.sausageW + 2
	}
	if g.cat2Right {
		if g.cat2X > cat2Target+0.05 {
			g.cat2X = math.Max(cat2Target, g.cat2X-46*dt)
			if g.rng.Float64() < 0.15 {
				g.dustAt(g.cat2X)
			}
			g.updateEffects(dt)
			return
		}
	} else if g.cat2X < cat2Target-0.05 {
		g.cat2X = math.Min(cat2Target, g.cat2X+46*dt)
		if g.rng.Float64() < 0.15 {
			g.dustAt(g.cat2X)
		}
		g.updateEffects(dt)
		return
	}
	g.cat2X = cat2Target

	if g.bites < 3 {
		g.biteTimer += dt
		if g.biteTimer >= 0.45 {
			g.biteTimer = 0
			g.bites++
			g.sausageW = math.Max(4, g.sausageW-2)
			if g.cat2Right {
				g.sausageX++
				g.catX++
				g.cat2X--
			} else {
				g.sausageX += 2
				g.catX += 2
				g.cat2X += 2
			}
			g.eatCrumbs()
			tx := g.catX + 4
			if g.bites%2 == 0 {
				tx = g.cat2X + 4
			}
			g.floats = append(g.floats, floatText{
				x:       tx,
				y:       float64(rt - 4),
				text:    "НЯМ!",
				life:    0.7,
				maxLife: 0.7,
				color:   220,
			})
		}
		g.updateEffects(dt)
		return
	}

	g.finishDone += dt
	if g.finishDone >= 0.7 {
		g.pillowShown = true
		g.asleep = false
		g.sleepFor = 0
		g.pillowX = clampF((g.catX+g.cat2X)/2-13, 4, float64(g.w)-26)
		g.floats = append(g.floats, floatText{
			x:       g.pillowX + 14,
			y:       float64(rt - 7),
			text:    "СПАТЬ!",
			life:    1.2,
			maxLife: 1.2,
			color:   220,
		})
		for i := 0; i < 8; i++ {
			g.particles = append(g.particles, particle{
				x:       g.pillowX + 2 + g.rng.Float64()*22,
				y:       float64(rt - 2),
				vx:      (g.rng.Float64() - 0.5) * 3,
				vy:      -2 - g.rng.Float64()*3,
				life:    0.5 + g.rng.Float64()*0.3,
				maxLife: 0.8,
				ch:      '+',
				color:   220,
				gravity: 3,
			})
		}
	}
	g.updateEffects(dt)
}

func (g *game) finishRun() {
	g.st = stateVictory
	g.victoryFor = 0
	g.pillowShown = true
	g.asleep = true
	g.score = int(g.points)
	g.newBest = g.score > g.best
	if g.newBest {
		g.best = g.score
		if g.persist {
			saveBest(g.best)
		}
	}
	g.floats = append(g.floats, floatText{
		x:       g.catX + catWidth/2,
		y:       float64(g.roadTop() - 8),
		text:    "МУР-Р!",
		life:    2.0,
		maxLife: 2.0,
		color:   220,
	})
}

func (g *game) beginDogAttack() {
	g.st = stateDog
	g.dogActive = true
	g.catEaten = false
	g.dogX = float64(g.w) + 4
	g.lossFor = 0
	g.lossDone = 0
	g.puddles = nil
	g.particles = nil
	g.floats = nil
	g.catY = 0
	g.vy = 0
	g.onGround = true
	g.invuln = 0
	g.flash = 0.8
	g.floats = append(g.floats, floatText{
		x:       g.catX + 3,
		y:       float64(g.roadTop() - 7),
		text:    "МЯУ!",
		life:    1.2,
		maxLife: 1.2,
		color:   203,
	})
}

func (g *game) updateDog(dt float64) {
	g.lossFor += dt
	rt := float64(g.roadTop())
	target := g.catX - 3

	if !g.catEaten {
		if g.dogX > target {
			g.dogX = math.Max(target, g.dogX-72*dt)
		}
		if g.dogX <= target+0.1 {
			g.catEaten = true
			g.dogX = target
			g.flash = 1
			g.floats = append(g.floats, floatText{
				x:       g.catX + 4,
				y:       rt - 6,
				text:    "ХВАТЬ!",
				life:    1.1,
				maxLife: 1.1,
				color:   220,
			})
			g.fur()
		}
	} else {
		g.lossDone += dt
		if g.lossDone > 1.25 {
			g.gameOver()
			return
		}
	}
	g.updateEffects(dt)
}

func (g *game) fur() {
	colors := []int{208, 214, 231, 217}
	for i := 0; i < 12; i++ {
		g.particles = append(g.particles, particle{
			x:       g.catX + 1 + g.rng.Float64()*7,
			y:       float64(g.roadTop()) - 2 - g.rng.Float64()*2,
			vx:      (g.rng.Float64() - 0.5) * 12,
			vy:      -2 - g.rng.Float64()*5,
			life:    0.4 + g.rng.Float64()*0.4,
			maxLife: 0.8,
			ch:      '·',
			color:   colors[g.rng.Intn(len(colors))],
			gravity: 20,
		})
	}
}

func (g *game) gameOver() {
	g.st = stateOver
	g.overFor = 0
	g.score = int(g.points)
	g.newBest = g.score > g.best
	if g.newBest {
		g.best = g.score
		if g.persist {
			saveBest(g.best)
		}
	}
}

func (g *game) overallProgress() float64 {
	switch g.st {
	case stateFinish, stateVictory:
		return 1
	case stateStageClear:
		return stageGoals[0] / (stageGoals[0] + stageGoals[1])
	}

	done := 0.0
	if g.level >= 2 {
		done = stageGoals[0]
	}
	done += g.stageDistance
	return clampF(done/(stageGoals[0]+stageGoals[1]), 0, 1)
}

func (g *game) stageLabel() string {
	switch g.st {
	case stateFinish:
		return "ФИНИШ"
	case stateVictory:
		return "ПРИЗ"
	case stateStageClear:
		return "ЭТАП 2/2"
	case stateDog:
		return "СОБАКА"
	case stateOver:
		return "ПРОИГРЫШ"
	}
	return fmt.Sprintf("ЭТАП %d/%d", g.level, maxStages)
}

func (g *game) handleKey(k key) (quit bool) {
	switch k {
	case keyQuit:
		return true
	case keyPause:
		switch g.st {
		case stateRunning:
			g.st = statePaused
		case statePaused:
			g.st = stateRunning
		}
	case keyJump:
		switch g.st {
		case stateReady:
			g.startRun()
		case stateRunning:
			g.jump()
		case stateStageClear:
			g.startStage(2)
		case stateOver:
			if g.overFor > 0.45 {
				g.startRun()
			}
		case stateVictory:
			if g.victoryFor > 0.4 {
				g.startRun()
			}
		}
	case keyRestart:
		switch g.st {
		case stateOver:
			if g.overFor > 0.45 {
				g.startRun()
			}
		case stateVictory:
			if g.victoryFor > 0.4 {
				g.startRun()
			}
		}
	}
	return false
}

func (g *game) resize(w, h int) {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	g.w, g.h = w, h
	if g.st != stateFinish && g.st != stateVictory {
		g.catX = clampF(float64(w)*0.14, 8, 20)
		return
	}
	if g.sausageX > float64(w)-8 {
		g.sausageX = float64(w) - 8
	}
	g.plateX = g.sausageX - 2
}

func (g *game) tooSmall() bool {
	return g.w < 64 || g.h < 20
}

func (g *game) roadTop() int {
	rt := g.h - roadRows
	if rt < 1 {
		rt = 1
	}
	return rt
}

func (g *game) updateEffects(dt float64) {
	for i := range g.particles {
		p := &g.particles[i]
		p.life -= dt
		p.vy += p.gravity * dt
		p.x += p.vx * dt
		p.y += p.vy * dt
		if p.y > float64(g.h) {
			p.life = 0
		}
	}
	liveP := g.particles[:0]
	for _, p := range g.particles {
		if p.life > 0 {
			liveP = append(liveP, p)
		}
	}
	g.particles = liveP

	for i := range g.floats {
		f := &g.floats[i]
		f.life -= dt
		f.y -= 2.2 * dt
	}
	liveF := g.floats[:0]
	for _, f := range g.floats {
		if f.life > 0 {
			liveF = append(liveF, f)
		}
	}
	g.floats = liveF
}

func (g *game) dust() {
	for i := 0; i < 4; i++ {
		g.particles = append(g.particles, particle{
			x:       g.catX + 1 + g.rng.Float64()*5,
			y:       float64(g.roadTop()) - 0.2,
			vx:      -3 - g.rng.Float64()*4,
			vy:      -1 - g.rng.Float64()*2.2,
			life:    0.3 + g.rng.Float64()*0.2,
			maxLife: 0.5,
			ch:      '·',
			color:   246,
			gravity: 10,
		})
	}
}

func (g *game) dustAt(x float64) {
	for i := 0; i < 3; i++ {
		g.particles = append(g.particles, particle{
			x:       x + 1 + g.rng.Float64()*4,
			y:       float64(g.roadTop()) - 0.2,
			vx:      -2 - g.rng.Float64()*3,
			vy:      -1 - g.rng.Float64()*2,
			life:    0.3 + g.rng.Float64()*0.2,
			maxLife: 0.5,
			ch:      '·',
			color:   246,
			gravity: 10,
		})
	}
}

func (g *game) eatCrumbs() {
	chars := []rune{'·', '*', '\''}
	colors := []int{173, 180, 166, 220}
	for i := 0; i < 9; i++ {
		g.particles = append(g.particles, particle{
			x:       g.sausageX - 1 + g.rng.Float64()*4,
			y:       float64(g.roadTop()) - 1,
			vx:      2 + g.rng.Float64()*7,
			vy:      -2 - g.rng.Float64()*5,
			life:    0.4 + g.rng.Float64()*0.35,
			maxLife: 0.75,
			ch:      chars[g.rng.Intn(len(chars))],
			color:   colors[g.rng.Intn(len(colors))],
			gravity: 24,
		})
	}
}

func (g *game) splash(x, y float64) {
	chars := []rune{'*', '·', '\'', '.'}
	colors := []int{45, 51, 39, 231}
	for i := 0; i < 14; i++ {
		g.particles = append(g.particles, particle{
			x:       x - 4 + g.rng.Float64()*8,
			y:       y,
			vx:      (g.rng.Float64() - 0.5) * 15,
			vy:      -3 - g.rng.Float64()*7,
			life:    0.4 + g.rng.Float64()*0.45,
			maxLife: 0.85,
			ch:      chars[g.rng.Intn(len(chars))],
			color:   colors[g.rng.Intn(len(colors))],
			gravity: 27,
		})
	}
}

func (g *game) sparks(x, y float64) {
	for i := 0; i < 5; i++ {
		g.particles = append(g.particles, particle{
			x:       x - 3 + g.rng.Float64()*6,
			y:       y - 1,
			vx:      (g.rng.Float64() - 0.5) * 4,
			vy:      -2 - g.rng.Float64()*2,
			life:    0.3 + g.rng.Float64()*0.25,
			maxLife: 0.55,
			ch:      '+',
			color:   220,
			gravity: 6,
		})
	}
}

func moveToward(v, target, step float64) float64 {
	if v < target {
		return math.Min(target, v+step)
	}
	return math.Max(target, v-step)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
