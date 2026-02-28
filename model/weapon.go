package model

import (
	util "raylib/playground/game/utils"
	point "raylib/playground/shared/point"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/solarlune/resolv"
)

type Weapon struct {
	Sprite              Sprite
	SpriteFlipped       bool
	ProjectileSpriteSrc Sprite
	Obj                 *resolv.Object
	Handle              point.Point
	Reach               int
	AttackSpeed         int
	Cooldown            int
	TintColor           rl.Color
	AttackFrame         int

	IdleRotation  float32
	AttackRotator func(w Weapon) float32

	ProjectileCount         int
	Projectilelength        int
	ProjectileSpreadDegrees int
	ProjectileTTLFrames     int
	ProjectileVelocity      int
}

func (w *Weapon) Move(dx, dy float64) {
	w.Obj.X += dx
	w.Obj.Y += dy
	w.Obj.Update()
	w.Sprite.Dest.X = util.RectFromObj(w.Obj).X
	w.Sprite.Dest.Y = util.RectFromObj(w.Obj).Y
}

func (w *Weapon) AnchoredMove(x, y float64) {
	w.Sprite.Dest.X = float32(x)
	w.Sprite.Dest.Y = float32(y)
	w.Obj = util.ObjFromRect(w.Sprite.Dest)
	w.Obj.Update()
}
