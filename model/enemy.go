package model

import (
	"github.com/solarlune/resolv"
)

type Enemy struct {
	Sprite      Sprite
	Obj         *resolv.Object
	Health      int
	MaxHealth   int
	HurtFrames  int
	DeathFrames int
	Dead        bool
}

func (e *Enemy) Hurt() {
	e.HurtFrames = 16
	e.Health -= 1
	if e.Health <= 0 && !e.Dead {
		e.Die()
	}
}

func (e *Enemy) Die() {
	e.DeathFrames = 32
	e.Dead = true
}
