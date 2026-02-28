package draw_world_engine

import (
	"fmt"
	"math"
	"raylib/playground/engines/physics-engine"
	util "raylib/playground/game/utils"
	"raylib/playground/model"
	"raylib/playground/model/draw2d"
	"raylib/playground/model/draw2d/texture-maps"
	"raylib/playground/shared/draw"
	"raylib/playground/shared/mapdata"
	"raylib/playground/shared/point"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	player            *model.Player
	enemies           *[]*model.Enemy
	currentMap        *mapdata.MapModel
	collisionMapDebug []rl.Rectangle
	frameCount        int32
)

func SetCurrentMap(_currentMap *mapdata.MapModel) {
	currentMap = _currentMap
}

func SetCollisionMapDebug(_collisionMapDebug []rl.Rectangle) {
	collisionMapDebug = _collisionMapDebug
}

func SetPlayer(_player *model.Player) {
	player = _player
}

func SetEnemies(_enemies *[]*model.Enemy) {
	enemies = _enemies
}

func DrawMapBackground() []draw.DrawParams {

	tileSrc := rl.Rectangle{
		Height: currentMap.SrcTileDimension.Height,
		Width:  currentMap.SrcTileDimension.Width,
	}
	tileDest := rl.Rectangle{
		Height: currentMap.DestTileDimension.Height,
		Width:  currentMap.DestTileDimension.Width,
	}

	var foreGroundDrawParams []draw.DrawParams
	for i, tileInt := range currentMap.TileMap {
		if tileInt == 0 {
			continue
		}
		tileDest.X = tileDest.Width * float32(i%currentMap.Width) // 6 % 5 means x column 1
		tileDest.Y = tileDest.Width * float32(i/currentMap.Width) // 6 % 5 means y row of 1
		tileMap := texturemaps.TileMapIndex[strings.ToLower(currentMap.SrcMap[i])]
		tileSrc.X = tileMap[tileInt].X
		tileSrc.Y = tileMap[tileInt].Y

		if strings.ToUpper(currentMap.SrcMap[i]) == currentMap.SrcMap[i] {
			// TODO make this fill square more sophisticated - maybe random or something
			fillTile := tileSrc
			fillTile.X = 16
			fillTile.Y = 64
			rl.DrawTexturePro(currentMap.Texture, fillTile, tileDest, rl.NewVector2(tileDest.Width, tileDest.Height), 0, rl.White)

			// draw behind player if Y is "behind" player, but skip this with Walls
			if strings.ToLower(currentMap.SrcMap[i]) == "w" || player != nil && tileDest.Y > player.Sprite.Dest.Y {
				foreGroundDrawParams = append(
					foreGroundDrawParams,
					draw.DrawParams{
						Texture:  currentMap.Texture,
						SrcRec:   tileSrc,
						DestRec:  tileDest,
						Origin:   rl.NewVector2(tileDest.Width, tileDest.Height),
						Rotation: 0,
						Tint:     rl.White,
					})
			} else {
				rl.DrawTexturePro(currentMap.Texture, tileSrc, tileDest, rl.NewVector2(tileDest.Width, tileDest.Height), 0, rl.White)
			}
		} else {
			rl.DrawTexturePro(currentMap.Texture, tileSrc, tileDest, rl.NewVector2(tileDest.Width, tileDest.Height), 0, rl.White)
		}
	}

	return foreGroundDrawParams
}

func drawPlayer(p *model.Player, fc int32) {
	if fc%8 == 1 {
		p.Sprite.Frame++
	}
	if p.Sprite.Frame > 3 {
		p.Sprite.Frame = 0
	}
	var weaponOffset float32 = 0
	if p.IsMoving() {
		p.Sprite.Src.X = 192                                                                           // pixel where run animation starts
		p.Sprite.Src.X += float32(p.Sprite.Frame) * float32(math.Abs(float64(p.Sprite.Src.Width))) // rolling the animation
		weaponOffset = -4
	} else {
		if p.SpriteFlipped {
			util.FlipLeft(&p.Sprite.Src)
		} else {
			util.FlipRight(&p.Sprite.Src)
		}
		p.Sprite.Src.X = 128                                                                           // pixel where rest idle starts
		p.Sprite.Src.X += float32(p.Sprite.Frame) * float32(math.Abs(float64(p.Sprite.Src.Width))) // rolling the animation
	}
	p.Weapon.SpriteFlipped = p.SpriteFlipped
	rl.DrawTexturePro(draw2d.Texture, p.Sprite.Src, p.Sprite.Dest, rl.NewVector2(p.Sprite.Dest.Width, p.Sprite.Dest.Height), 0, rl.White)
	updateFrame := fc%8 == 0
	drawWeapon(p.Weapon, p.Sprite.Frame, updateFrame, weaponOffset)
}

func drawWeapon(w *model.Weapon, frame int, nextFrame bool, offset float32) {
	rotation := w.IdleRotation
	if w.AttackFrame >= 0 && w.AttackRotator != nil {
		rotation = w.AttackRotator(*w)
		w.AttackFrame++

		if w.AttackFrame >= w.AttackSpeed {
			w.AttackFrame = -1 // setting to -1 to symbolize attack is finished animating
			w.Move(0, 0)       // recenter weapon after attack animation
		}

	} else if nextFrame {

		if frame == 0 || frame == 1 {
			w.Sprite.Dest.Y += 1
		} else {
			w.Sprite.Dest.Y -= 1
		}
	}

	if !w.SpriteFlipped {
		util.FlipRight(&w.Sprite.Src)
	}
	if w.SpriteFlipped {
		util.FlipLeft(&w.Sprite.Src)
		rotation *= -1
	}

	origin := rl.NewVector2(w.Handle.X, w.Handle.Y)
	dest := w.Sprite.Dest
	dest.Y += offset

	rl.DrawTexturePro(w.Sprite.Texture, w.Sprite.Src, dest,
		origin, rotation, w.TintColor)
}

func drawEnemy(e *model.Enemy, fc int32) {
	if fc%8 == 1 && !e.Dead {
		e.Sprite.Frame++
	}
	if e.Sprite.Frame > 3 {
		e.Sprite.Frame = 0
	}

	tint := rl.White
	if e.HurtFrames > 0 {
		tint = rl.Red
		e.HurtFrames--
	}
	if e.DeathFrames > 0 {
		if e.Sprite.Rotation < 90 {
			e.Sprite.Rotation = float32(math.Min(90, float64(e.Sprite.Rotation)+8))
		}
		e.DeathFrames--
	}

	e.Sprite.Src.X = 368                                                                           // pixel where rest idle starts
	e.Sprite.Src.X += float32(e.Sprite.Frame) * float32(math.Abs(float64(e.Sprite.Src.Width))) // rolling the animation

	rl.DrawTexturePro(draw2d.Texture, e.Sprite.Src, e.Sprite.Dest, rl.NewVector2(e.Sprite.Dest.Width, e.Sprite.Dest.Height), e.Sprite.Rotation, tint)

	if e.Health != e.MaxHealth && !e.Dead {
		rl.DrawRectangle(int32(e.Obj.X), int32(e.Obj.Y-10), int32(e.Obj.W), 4, rl.Red)
		rl.DrawRectangle(int32(e.Obj.X), int32(e.Obj.Y-10), int32(int(e.Obj.W)*e.Health/e.MaxHealth), 4, rl.Green)
	}
}

func drawProjectile(p *model.Projectile) {
	w := p.Sprite.Dest.Width
	h := p.Sprite.Dest.Height
	dest := rl.NewRectangle(p.Start.X, p.Start.Y, w, h)
	rl.DrawTexturePro(p.Sprite.Texture, p.Sprite.Src, dest,
		rl.NewVector2(dest.Width/2, dest.Height), float32(180-p.Trajectory), rl.White)
}

func DrawScene(debugMode bool) {
	foreGround := DrawMapBackground()

	if player != nil {
		drawPlayer(player, frameCount)
	}
	if enemies != nil {
		for _, e := range *enemies {
			drawEnemy(e, frameCount)
		}
	}
	for i := range physics_engine.Projectiles {
		drawProjectile(&physics_engine.Projectiles[i])
	}
	// drawing foreground after player so it appears "in-front"
	for _, d := range foreGround {
		rl.DrawTexturePro(d.Texture, d.SrcRec, d.DestRec, d.Origin, d.Rotation, d.Tint)
	}

	// draw debug collision objects
	if debugMode {
		for _, o := range collisionMapDebug {
			mo := util.ObjFromRect(o)
			rl.DrawRectangleLines(int32(mo.X), int32(mo.Y), int32(mo.W), int32(mo.H), rl.White)
		}

		for _, p := range physics_engine.Projectiles {
			rl.DrawLine(int32(p.Start.X), int32(p.Start.Y), int32(p.End.X), int32(p.End.Y), rl.Pink)
		}

		// debug player collision box
		po := player.Obj
		rl.DrawRectangleLines(int32(po.X), int32(po.Y), int32(po.W), int32(po.H), rl.Orange)

		for _, e := range *enemies {
			rl.DrawRectangleLines(int32(e.Obj.X), int32(e.Obj.Y), int32(e.Obj.W), int32(e.Obj.H), rl.White)
		}

		playerCenter := point.Point{
			X: float32(player.Obj.X + player.Obj.W/2),
			Y: float32(player.Obj.Y + player.Obj.H/2),
		}
		rl.DrawCircleLines(int32(playerCenter.X), int32(playerCenter.Y), 32, rl.Green)
		angle := util.GetPlayerToMouseAngleDegrees()
		rl.DrawCircleSectorLines(rl.NewVector2(playerCenter.X, playerCenter.Y), 32, angle, angle-45, 5, rl.White)
		rl.DrawCircleSectorLines(rl.NewVector2(playerCenter.X, playerCenter.Y), 32, angle, angle+45, 5, rl.White)
	}

	frameCount = (frameCount + 1) % 256
}

func DrawUI(debugMode bool) {
	if debugMode {
		rl.DrawRectangleRounded(rl.NewRectangle(3, 3, 500, 90), .1, 10, rl.DarkGray)
		rl.DrawRectangleRoundedLines(rl.NewRectangle(3, 3, 500, 90), .1, 10, 3, rl.White)
		rl.DrawText(fmt.Sprintf("FPS: %v", rl.GetFPS()), 10, 10, 16, rl.White)
		rl.DrawText(fmt.Sprintf("player {X: %v, Y:%v}", player.Obj.X, player.Obj.Y), 10, 30, 16, rl.White)
		rl.DrawText(fmt.Sprintf("mouse  {X: %v, Y:%v}", rl.GetMouseX(), rl.GetMouseY()), 10, 50, 16, rl.White)

		// wierd thing where rise/run are opposite directions (think it has to do with x/y being negative flipped)
		rise := float64(rl.GetMouseX()) - float64(rl.GetScreenWidth()/2)
		run := float64(rl.GetMouseY()) - float64(rl.GetScreenHeight())/2

		angle := util.GetPlayerToMouseAngleDegrees()
		rl.DrawText(fmt.Sprintf("mouse->player  {X: %v, Y:%v}", rise, run), 10, 70, 16, rl.White)
		rl.DrawText(fmt.Sprintf("Atan(%v/%v) = %v degrees", rise, run, int(angle)), 250, 10, 16, rl.White)
		rl.DrawText(fmt.Sprintf("Live Projectiles: %v", len(physics_engine.Projectiles)), 250, 30, 16, rl.White)
	}
}
