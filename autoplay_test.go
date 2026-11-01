package main

import (
	"math/rand"
	"testing"
)

// TestPuddlesAreJumpable проверяет, что оба этапа проходимы, а в финале
// котик добегает до сосиски, встречает второго котика и засыпает на лежанке.
func TestPuddlesAreJumpable(t *testing.T) {
	airTime := 2 * jumpSpeed / gravity
	for seed := int64(1); seed <= 5; seed++ {
		g := newGame(100, 30, 0)
		g.rng = rand.New(rand.NewSource(seed))
		g.startRun()
		g.persist = false

		for frame := 0; frame < 60*120; frame++ {
			if g.st == stateVictory {
				break
			}
			if g.st == stateOver {
				t.Fatalf("seed %d: котик погиб на %.1f-й секунде (скорость %.1f)",
					seed, float64(frame)/60, g.speed)
			}
			if g.onGround {
				for _, p := range g.puddles {
					if p.x < g.catX+2.5+catHitbox {
						continue
					}
					clearIn := (p.x + p.w - (g.catX + 2.5)) / g.speed
					if clearIn <= airTime-0.15 {
						g.jump()
						break
					}
				}
			}
			g.update(1.0 / 60)
		}
		if g.st != stateVictory {
			t.Fatalf("seed %d: котик не дошёл до сосиски, состояние %v, этап %d", seed, g.st, g.level)
		}
		if g.bites != 3 {
			t.Fatalf("seed %d: котик не съел сосиску, укусов %d", seed, g.bites)
		}
		if !g.cat2Active || !g.pillowShown || !g.asleep {
			t.Fatalf("seed %d: второй котик не уснул на своей подушке (cat2=%v pillows=%v asleep=%v)",
				seed, g.cat2Active, g.pillowShown, g.asleep)
		}
	}
}
