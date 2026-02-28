package map_director

import (
	"raylib/playground/engines/draw-world-engine"
	"raylib/playground/engines/physics-engine"
	map_loader "raylib/playground/loaders/map-loader"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func LoadMap(mapFile string, texture rl.Texture2D) {

	mapModel := map_loader.LoadMap(mapFile, texture)
	collisionMapDebug := physics_engine.SetWorldSpaceCollidables(mapModel)

	draw_world_engine.SetCurrentMap(mapModel)
	draw_world_engine.SetCollisionMapDebug(collisionMapDebug)
}
