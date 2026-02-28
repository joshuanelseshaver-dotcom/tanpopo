package model

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Projectile struct {
	Start rl.Vector2
	End   rl.Vector2
	Ttl   int

	// for something like an arrow perhaps
	Velocity   int
	Trajectory float64 //degrees
	Sprite     Sprite

	// sender     *interface{} at somepoint this would be good to have
}
