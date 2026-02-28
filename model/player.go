package model

import (
	"math"
	util "raylib/playground/game/utils"
	point "raylib/playground/shared/point"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/solarlune/resolv"
)

type Test struct {
	Test   string
	Sprite Sprite
}

type PlayerMovement struct {
	Up    bool
	Down  bool
	Left  bool
	Right bool
}

type Player struct {
	Sprite         Sprite
	SpriteFlipped  bool
	Obj            *resolv.Object
	Weapon         *Weapon
	Hand           point.Point
	Moving         PlayerMovement
	Speed          float32
	Attacking      bool
	AttackCooldown int
}

func (p *Player) Move(dx, dy float64) {
	p.Obj.X += dx
	p.Obj.Y += dy
	p.Obj.Update()
	p.Sprite.Dest.X = util.RectFromObj(p.Obj).X
	p.Sprite.Dest.Y = util.RectFromObj(p.Obj).Y

	ax := float64(p.Hand.X) + p.Obj.X
	ay := float64(p.Hand.Y) + p.Obj.Y
	p.Weapon.AnchoredMove(ax, ay)
}

func (p *Player) Attack() []Projectile {
	p.Weapon.AttackFrame = 0 // find a better way to trigger animation than this.
	p.AttackCooldown = p.Weapon.Cooldown

	playerCenter := point.Point{
		X: float32(p.Obj.X + p.Obj.W/2),
		Y: float32(p.Obj.Y + p.Obj.H/2),
	}
	angle := util.GetPlayerToMouseAngleDegrees()

	// TODO use weapon attributes in the future to determine this logic
	projectileCount := p.Weapon.ProjectileCount
	projectileReach := p.Weapon.Projectilelength
	projectileSpread := p.Weapon.ProjectileSpreadDegrees
	projectileTTL := p.Weapon.ProjectileTTLFrames
	projectileVelocity := p.Weapon.ProjectileVelocity
	projectileSpreadItter := int(float64(angle) - math.Floor(float64(projectileCount)/2)*float64(projectileSpread))

	var newProjectiles []Projectile
	for i := 0; i < projectileCount; i++ {
		x2 := int(float64(projectileReach) * math.Sin(util.DegreesToRadians(float64(projectileSpreadItter))))
		y2 := int(float64(projectileReach) * math.Cos(util.DegreesToRadians(float64(projectileSpreadItter))))
		var projectileTrajectory float64
		if projectileVelocity > 0 {
			projectileTrajectory = float64(projectileSpreadItter)
		}
		newProjectile := Projectile{
			Start:      rl.NewVector2(playerCenter.X, playerCenter.Y),
			End:        rl.NewVector2(playerCenter.X+float32(x2), playerCenter.Y+float32(y2)),
			Ttl:        projectileTTL,
			Velocity:   projectileVelocity,
			Trajectory: projectileTrajectory,
			Sprite: Sprite{
				Src:     p.Weapon.ProjectileSpriteSrc.Src,
				Dest:    p.Weapon.ProjectileSpriteSrc.Dest,
				Texture: p.Weapon.ProjectileSpriteSrc.Texture,
			},
		}
		newProjectiles = append(newProjectiles, newProjectile)
		projectileSpreadItter += projectileSpread
	}
	p.Attacking = false
	return newProjectiles
}

func (p *Player) EquipWeapon(w *Weapon) {
	// create new object from updated dest X/Y
	w.Sprite.Dest.X = p.Hand.X + float32(p.Obj.X)
	w.Sprite.Dest.Y = p.Hand.Y + float32(p.Obj.Y)
	w.Obj = util.ObjFromRect(w.Sprite.Dest)

	// update player weapon
	p.Weapon = w
}

func (p *Player) IsMoving() bool {
	return p.Moving.Up || p.Moving.Down || p.Moving.Left || p.Moving.Right
}
