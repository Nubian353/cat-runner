package main

import (
	"regexp"
	"strings"
	"testing"
)

func newTestGame() *game {
	return newGame(80, 24, 0)
}

var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]")

func plain(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func TestJumpArcAndLanding(t *testing.T) {
	g := newTestGame()
	g.startRun()

	peak := 0.0
	landed := false
	for i := 0; i < 180; i++ {
		g.update(1.0 / 60)
		if g.catY > peak {
			peak = g.catY
		}
		if g.onGround && i > 10 {
			landed = true
			break
		}
	}
	if !landed {
		t.Fatal("котик не приземлился")
	}
	if peak < 4.5 || peak > 7 {
		t.Fatalf("высота прыжка не в норме: %.2f", peak)
	}
}

func TestPuddleHitCostsOneLife(t *testing.T) {
	g := newTestGame()
	g.startRun()
	g.spawnIn = 999
	g.puddles = []puddle{{x: g.catX + 1, w: 6}}
	g.catY = 0
	g.onGround = true
	g.invuln = 0

	g.update(1.0 / 60)

	if g.lives != maxLives-1 {
		t.Fatalf("жизней должно быть %d, получено %d", maxLives-1, g.lives)
	}
	if g.invuln <= 0 {
		t.Fatal("после попадания должна быть неуязвимость")
	}
	if !g.puddles[0].hit {
		t.Fatal("лужа должна быть помечена как задевшая котика")
	}
}

func TestJumpClearsPuddle(t *testing.T) {
	g := newTestGame()
	g.startRun()
	g.spawnIn = 999
	g.puddles = []puddle{{x: g.catX + 1, w: 6}}
	g.catY = 3
	g.vy = 0
	g.onGround = false
	g.invuln = 0

	g.update(1.0 / 60)

	if g.lives != maxLives {
		t.Fatalf("в прыжке жизнь не должна теряться, жизней: %d", g.lives)
	}
}

func TestThreeHitsEndTheGame(t *testing.T) {
	g := newTestGame()
	g.startRun()
	g.spawnIn = 999

	for i := 0; i < maxLives; i++ {
		g.invuln = 0
		g.catY = 0
		g.onGround = true
		g.puddles = []puddle{{x: g.catX + 1, w: 6}}
		g.update(1.0 / 60)
	}

	if g.lives != 0 {
		t.Fatalf("ожидалось 0 жизней, получено %d", g.lives)
	}
	if g.st != stateDog {
		t.Fatalf("после третьего попадания должна появиться собака, состояние %v", g.st)
	}
	stepFor(g, 3.2)
	if g.st != stateOver {
		t.Fatalf("собака не завершила сцену, состояние %v", g.st)
	}
	if !g.catEaten {
		t.Fatal("котик должен быть съеден")
	}
}

func TestClearedPuddleGivesBonus(t *testing.T) {
	g := newTestGame()
	g.startRun()
	g.spawnIn = 999
	g.points = 0
	g.score = 0
	g.puddles = []puddle{{x: g.catX - 15, w: 5}}

	g.update(1.0 / 60)

	if g.points < clearBonus {
		t.Fatalf("за перепрыгнутую лужу должен быть бонус, очков: %.2f", g.points)
	}
}

func TestNewRecord(t *testing.T) {
	g := newGame(80, 24, 10)
	g.startRun()
	g.spawnIn = 999
	g.points = 200
	g.score = 200
	g.lives = 1
	g.puddles = []puddle{{x: g.catX + 1, w: 6}}
	g.catY = 0
	g.onGround = true

	g.update(1.0 / 60)
	if g.st != stateDog {
		t.Fatalf("сначала должна появиться собака, состояние %v", g.st)
	}
	stepFor(g, 3.2)
	if g.st != stateOver {
		t.Fatalf("сцена с собакой не закончилась, состояние %v", g.st)
	}
	if !g.newBest {
		t.Fatal("новый рекорд должен быть зафиксирован")
	}
	if g.best < 200 {
		t.Fatalf("рекорд должен обновиться, получено %d", g.best)
	}
}

func TestPauseFreezesTheWorld(t *testing.T) {
	g := newTestGame()
	g.startRun()
	g.handleKey(keyPause)
	if g.st != statePaused {
		t.Fatal("игра должна встать на паузу")
	}
	before := g.distance
	g.update(1.0)
	if g.distance != before {
		t.Fatal("на паузе мир не должен двигаться")
	}
	g.handleKey(keyPause)
	if g.st != stateRunning {
		t.Fatal("игра должна продолжиться")
	}
}

func TestRestartAfterGameOver(t *testing.T) {
	g := newTestGame()
	g.startRun()
	g.lives = 1
	g.puddles = []puddle{{x: g.catX + 1, w: 6}}
	g.catY = 0
	g.onGround = true
	g.update(1.0 / 60)
	if g.st != stateDog {
		t.Fatalf("собака не появилась, состояние %v", g.st)
	}
	stepFor(g, 3.2)
	if g.st != stateOver {
		t.Fatalf("игра должна была закончиться, состояние %v", g.st)
	}
	g.overFor = 1.0
	g.handleKey(keyJump)
	if g.st != stateRunning || g.lives != maxLives || g.score != 0 {
		t.Fatal("после рестарта состояние должно сброситься")
	}
}

func TestRenderFrames(t *testing.T) {
	for _, size := range [][2]int{{64, 20}, {80, 24}, {120, 40}} {
		g := newGame(size[0], size[1], 123)
		frame := g.renderFrame()
		if !strings.Contains(plain(frame), "ОЧКИ") {
			t.Fatalf("в кадре %dx%d нет счёта", size[0], size[1])
		}
		if !strings.Contains(plain(frame), "КОТИК-БЕГУН") {
			t.Fatalf("в стартовом кадре %dx%d нет заголовка", size[0], size[1])
		}
		if strings.Contains(plain(frame), "Окно слишком маленькое") {
			t.Fatalf("кадр %dx%d ошибочно считается маленьким", size[0], size[1])
		}
	}
}

func TestTinyWindowMessage(t *testing.T) {
	g := newGame(40, 12, 0)
	if !strings.Contains(plain(g.renderFrame()), "Окно слишком маленькое") {
		t.Fatal("для маленького окна нужна подсказка")
	}
}

func stepFor(g *game, seconds float64) {
	steps := int(seconds * 60)
	for i := 0; i < steps; i++ {
		g.update(1.0 / 60)
	}
}

func TestTwoStagesAndSausagePrize(t *testing.T) {
	g := newGame(100, 30, 0)
	g.startRun()
	g.spawnIn = 9999
	g.puddles = nil
	g.stageDistance = g.stageGoal - 0.2
	g.update(1.0 / 60)

	if g.st != stateStageClear {
		t.Fatalf("после первого этапа ожидался переход, состояние %v", g.st)
	}
	if g.stageLabel() != "ЭТАП 2/2" {
		t.Fatalf("неверная подпись этапа: %s", g.stageLabel())
	}

	g.handleKey(keyJump) // пропустить заставку
	if g.st != stateRunning || g.level != 2 {
		t.Fatalf("второй этап не начался: state=%v level=%d", g.st, g.level)
	}

	g.spawnIn = 9999
	g.puddles = nil
	g.stageDistance = g.stageGoal - 0.2
	g.update(1.0 / 60)
	if g.st != stateFinish {
		t.Fatalf("после второго этапа ожидался финиш, состояние %v", g.st)
	}
	if g.sausageW != 12 {
		t.Fatalf("сосиска должна быть целой, ширина %.1f", g.sausageW)
	}

	stepFor(g, 9)
	if g.st != stateVictory {
		t.Fatalf("победа не наступила, состояние %v", g.st)
	}
	if g.bites != 3 {
		t.Fatalf("котики должны откусить три раза, укусов %d", g.bites)
	}
	if !g.cat2Active || !g.pillowShown || !g.asleep {
		t.Fatalf("второй котик должен уснуть на своей подушке: cat2=%v pillows=%v asleep=%v",
			g.cat2Active, g.pillowShown, g.asleep)
	}
	if g.sausageW != 6 {
		t.Fatalf("после трёх укусов должна остаться половина, ширина %.1f", g.sausageW)
	}
	if g.score <= 0 || g.best != g.score || !g.newBest {
		t.Fatalf("рекорд за победу не сохранён: score=%d best=%d new=%v", g.score, g.best, g.newBest)
	}
}

func TestStageClearAutoAdvances(t *testing.T) {
	g := newGame(100, 30, 0)
	g.startRun()
	g.spawnIn = 9999
	g.puddles = nil
	g.stageDistance = g.stageGoal - 0.2
	g.update(1.0 / 60)
	if g.st != stateStageClear {
		t.Fatalf("ожидался переход между этапами, состояние %v", g.st)
	}

	stepFor(g, 2.4)
	if g.st != stateRunning || g.level != 2 || g.stageDistance > 10 {
		t.Fatalf("второй этап не стартовал автоматически: state=%v level=%d dist=%.2f", g.st, g.level, g.stageDistance)
	}
}

func TestVictoryRestart(t *testing.T) {
	g := newGame(100, 30, 0)
	g.startRun()
	g.lives = 1
	g.stageDistance = g.stageGoal - 0.2
	g.update(1.0 / 60)
	g.handleKey(keyJump)
	g.spawnIn = 9999
	g.puddles = nil
	g.stageDistance = g.stageGoal - 0.2
	g.update(1.0 / 60)
	stepFor(g, 9)
	if g.st != stateVictory {
		t.Fatalf("победа не наступила, состояние %v", g.st)
	}

	g.victoryFor = 1.0
	g.handleKey(keyJump)
	if g.st != stateRunning || g.level != 1 || g.stageDistance != 0 || g.lives != maxLives {
		t.Fatalf("рестарт после победы не сбросил игру: state=%v level=%d dist=%.2f lives=%d",
			g.st, g.level, g.stageDistance, g.lives)
	}
}

func TestRenderFinishAndVictory(t *testing.T) {
	g := newGame(80, 24, 0)
	g.st = stateFinish
	g.level = 2
	g.stageDistance = stageGoals[1]
	g.sausageX = 58
	g.sausageW = 12
	g.plateX = 56
	g.catX = 42
	g.onGround = true
	finish := plain(g.renderFrame())
	if !strings.Contains(finish, "ФИНИШ") {
		t.Fatal("на финишном экране нет подписи ФИНИШ")
	}
	if !strings.Contains(finish, "ДО СОСИСКИ") {
		t.Fatal("на финишном экране нет цели")
	}

	g.st = stateVictory
	g.sausageW = 6
	g.newBest = true
	g.pillowShown = true
	g.asleep = true
	g.cat2Active = true
	g.pillowX = 40
	g.cat2X = g.pillowX + pillowW + pillowGap + 1
	g.catX = g.pillowX + 1
	victory := plain(g.renderFrame())
	if !strings.Contains(victory, "СОСИСКА СЪЕДЕНА!") {
		t.Fatal("на экране победы нет заголовка")
	}
	if !strings.Contains(victory, "ПРИЗ") {
		t.Fatal("на экране победы нет подписи ПРИЗ")
	}
	if !strings.Contains(victory, "два котика") && !strings.Contains(victory, "Два котика") {
		t.Fatal("на экране победы нет упоминания двух котиков")
	}

	g.st = stateOver
	g.dogActive = true
	g.catEaten = true
	g.dogX = 20
	g.overFor = 1
	loss := plain(g.renderFrame())
	if !strings.Contains(loss, "КОТИКА СЪЕЛИ") {
		t.Fatal("на экране проигрыша нет собаки и заголовка")
	}
}

func TestDogEatsCatOnLoss(t *testing.T) {
	g := newGame(100, 30, 0)
	g.startRun()
	g.spawnIn = 9999
	g.puddles = []puddle{{x: g.catX + 1, w: 6}}
	g.lives = 1
	g.catY = 0
	g.onGround = true
	g.update(1.0 / 60)

	if g.st != stateDog || !g.dogActive {
		t.Fatalf("при проигрыше должна появиться собака: state=%v dog=%v", g.st, g.dogActive)
	}
	if g.catEaten {
		t.Fatal("собака не должна съесть котика мгновенно")
	}
	stepFor(g, 1.4)
	if !g.catEaten {
		t.Fatal("собака должна догнать и съесть котика")
	}
	stepFor(g, 1.6)
	if g.st != stateOver {
		t.Fatalf("после сцены с собакой должен быть проигрыш, состояние %v", g.st)
	}
	if !strings.Contains(plain(g.renderFrame()), "КОТИКА СЪЕЛИ") {
		t.Fatal("нет заголовка про собаку на экране проигрыша")
	}
}
