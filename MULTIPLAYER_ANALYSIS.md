# golang_raylib — Repo Analysis & Multiplayer Roadmap

## 1. What's In the Repo

### Tech Stack
- **Language**: Go 1.18
- **Rendering / Input / Audio**: `github.com/gen2brain/raylib-go` (Go bindings for raylib)
- **Collision detection**: `github.com/solarlune/resolv` (spatial grid library)

### Architecture Overview

The engine is organized into five conceptual layers that flow downward:

```
main.go
  └── game/          (coordination layer)
        ├── engines/ (systems layer)
        ├── directors/ (orchestration between engines)
        ├── director-models/ (DTOs)
        └── model/   (data layer — currently mixed with rendering)
```

#### `main.go`
The entry point is deliberately minimal — three calls in a loop:
```go
game.Initialize(true)
for game.Running {
    game.ReadPlayerInputs()
    game.Update()
    game.Render()
}
game.Quit()
```

#### `game/` package
The top-level game coordinator.

| File | Responsibility |
|------|---------------|
| `game-state.go` | Global vars (Running, Enemies, DebugMode), Initialize(), Quit(), Render() |
| `game-loop.go` | Update(): player movement, attack cooldowns, projectile outcomes, audio, camera |
| `input-router.go` | ReadPlayerInputs(): maps WASD/mouse to PlayerMovement bool struct + weapon hotkeys |
| `main-player.go` | Player construction/initialization |
| `camera.go` | 2D camera following player, zoom via mouse wheel |
| `window.go` | Screen constants (1500×960) and background color |

#### `model/` package
Core data structs. **Important caveat**: rendering is currently embedded in these structs (see §3).

| File | What it holds |
|------|--------------|
| `player.go` | `Player` struct: `Sprite`, `resolv.Object`, `Weapon`, `PlayerMovement`, `Speed`, `Attacking`, `AttackCooldown`. Contains `Draw()`, `Move()`, `Attack()`, `EquipWeapon()`, `IsMoving()`. |
| `enemy.go` | `Enemy` struct: `Sprite`, `resolv.Object`, `Health`, `HurtFrames`, `DeathFrames`, `Dead`. Contains `Draw()`, `Hurt()`, `Die()`. |
| `weapon.go` | `Weapon` struct: full projectile configuration (count, spread, TTL, velocity), attack animation via `AttackRotator func`. Contains `Draw()`, `Move()`, `AnchoredMove()`. |
| `projectile.go` | `Projectile` struct: `Start`/`End` Vector2, TTL, Velocity, Trajectory (degrees), Sprite. Contains `Draw()`. |
| `sprite.go` | `Sprite` struct: Src/Dest `rl.Rectangle`, Texture, Frame, FrameCount, Rotation |
| `armory/` | Weapon factories: swords (RegularSword, BowShooter, Key), bows (RegularBow, SwordShooter), staves (Keytar, PizzaShooter), cannon (PeopleShooter) |

#### `engines/` package

| Engine | Responsibility |
|--------|---------------|
| `physics-engine/collision.go` | `WorldCollisionSpace` (global `resolv.Space`), builds collision objects from map data, handles "env", "nav" (door), and "enemy" tags |
| `physics-engine/player-movement.go` | `CalculatePlayerMovement()`: translates moving bools → dx/dy, applies AABB collision, returns nav-tile trigger for map transitions |
| `physics-engine/projectile.go` | Global `Projectiles []Projectile` slice, `CalculatePlayerProjectileOutcome()` (line-based hit detection against enemy AABB), `FireProjects()` |
| `draw-world-engine/draw-world-engine.go` | `DrawScene()`: map bg, player, enemies, projectiles, foreground tiles, debug overlays. `DrawUI()`: debug HUD |
| `map-engine/map-engine.go` | Parses custom `.map` text format (width, height, 3 tile layers: tileMap, srcMap, collisionMap), creates `resolv.Space` |
| `audio-engine/` | Init/unload audio, play/toggle music, play sword sound |
| `spawn-engine/enemy-spawner.go` | `NewEnemy()` factory (hardcoded orc_warrior sprite for now) |
| `draw-ui-engine/` | Separate UI draw package (currently unused beyond draw-world-engine) |

#### `directors/`
Orchestrators that coordinate multiple engines.

| Director | Does |
|----------|------|
| `map-director/map-director.go` | `LoadMap()`: calls map-engine → physics-engine collision setup → draw-world-engine state updates |

#### `director-models/`
Plain data transfer objects used between directors and engines.

- `map-model/` — `MapModel` struct (Width, Height, TileMap, SrcMap, CollisionMap, texture)
- `draw-model/` — `DrawParams` struct (for deferred foreground rendering)
- `point-model/` — `Point{X, Y float32}`

#### `dev-tools/map-maker/`
A work-in-progress GUI map editor (`gui/map-maker-gui.go`) with:
- Edit mode buttons (Pencil, Erase, Area)
- Layer selection (Background, MiddleGround, Foreground, All)
- Asset picker panel (floor tile previews from the sprite sheet)
- Save/collision editor buttons (stubs)

#### `resources/`
- `sprites/frames/` — ~100 PNG sprites: player animations, 8 enemy types (goblin, imp, ogre, masked orc, skeletons, wizard M/F, zombie, wogol), weapon sprites (15+ weapons)
- `sprites/` — Tile spritesheet (`tiles_list_v1.4`), pizza slice, key blade
- `maps/` — Custom `.map` files (`first.map`, `second.map`)

### Current Game State (What Actually Runs)
- Single-player top-down game
- Player movement (WASD/arrows), mouse-aimed attacks (left click)
- 6 equippable weapons via hotkeys 1-6: Keytar, RegularBow, SwordShooter, BowShooter, PeopleShooter, PizzaShooter
- One spawned enemy (orc warrior) with health bar, hurt flash, and tipping death animation
- Tile-based map with pixel-level AABB collision
- Door/nav tile detection triggers map transitions
- Background music + sword sound effects
- Camera follows player; mouse wheel zoom (0.75× – 2.0×)
- Backslash toggles debug overlay (hitboxes, projectile lines, FPS, aim angle)

### Open TODOs (from `TODOs` file)
- 4th map layer for gap-filling floor tiles
- Foreground depth sort by player center (not sprite top)
- Projectile/environment collision
- NPCs (non-enemy, interactable)
- Enemy spawn logic (position-based, wave-based)
- UI (health, inventory, money)
- Weapon damage attributes and healing weapons
- Weapon rotation bug with keytar at non-zero `IdleRotation`
- **In progress**: new map with navigation between zones

---

## 2. Changes Needed for Multiplayer

### The Core Problem: Rendering Is Embedded in the Model

The biggest architectural obstacle is that `Draw()` methods live inside `Player`, `Enemy`, `Weapon`, and `Projectile`. This means the model structs directly import `github.com/gen2brain/raylib-go/raylib` and call rendering functions. A server cannot run these structs headlessly — importing raylib on a server will fail or require a display.

Secondary issue: `Player.Draw()` calls `rl.GetMouseX/Y()` and `rl.GetScreenWidth/Height()` to decide facing direction, and `Player.Attack()` calls `audioengine.PlaySound()`. These are client-only concerns inside shared model logic.

The resolv dependency in models is a separate concern — resolv has no rendering dependency and can run server-side, but it may be worth replacing with a simpler AABB implementation for a server that should have zero external graphical dependencies.

---

### Phase 1: Decouple Rendering from Model (Prerequisite — Do This First)

This is purely internal refactoring, no networking yet. Estimated ~1-2 focused sessions.

**Changes:**

1. **Remove `Draw()` from `model/player.go`, `model/enemy.go`, `model/weapon.go`, `model/projectile.go`.**
   Move all draw logic into `draw-world-engine`. The draw engine already holds references to the player and enemies; it should own the rendering, not the models.

2. **Remove `rl.*` imports from the `model/` package entirely.**
   `model/player.go` uses `rl.DrawCircleLines` inside `Attack()` (debug only) and `rl.DrawTexturePro` inside `Draw()`.
   `model/enemy.go` uses `rl.DrawTexturePro` and `rl.DrawRectangle` for health bars.
   All of this moves to the draw engine.

3. **Move facing-direction logic out of `Player.Draw()`.**
   The idle direction check (`rl.GetMouseX()` vs `rl.GetScreenWidth()/2`) should be computed in `input-router.go` and stored as a field on Player (e.g., `FacingRight bool`).

4. **Move `audioengine.PlaySound()` out of `Player.Attack()`.**
   Attack() should return data only. The draw engine or game loop plays the sound when it sees an attack event.

After this phase: `model/` has zero raylib imports and can run in a server binary.

---

### Phase 2: Create a Shared Protocol Layer

A new `shared/` package with pure data structs (no raylib, no resolv) that both server and client import.

```go
// shared/state.go
type PlayerID uint32

type PlayerInput struct {
    PlayerID   PlayerID
    Tick       uint64
    Up, Down, Left, Right bool
    Attacking  bool
    WeaponSlot int
    AimAngle   float64  // degrees, replaces mouse position
}

type PlayerState struct {
    ID            PlayerID
    X, Y          float64
    FacingRight   bool
    WeaponSlot    int
    Attacking     bool
    AttackCooldown int
    Health        int
    AnimFrame     int32
}

type EnemyState struct {
    ID          uint32
    X, Y        float64
    Health      int
    MaxHealth   int
    HurtFrames  int
    DeathFrames int
    Dead        bool
    AnimFrame   int32
}

type ProjectileState struct {
    ID         uint32
    StartX, StartY float32
    EndX, EndY     float32
    Ttl        int
    Velocity   int
    Trajectory float64
    WeaponType int
}

type WorldSnapshot struct {
    Tick       uint64
    Players    []PlayerState
    Enemies    []EnemyState
    Projectiles []ProjectileState
}
```

The `AimAngle` field is key: instead of sending mouse pixel coordinates (which are screen-relative and meaningless to the server), the client computes the world-space angle and sends that. The server uses it to spawn projectiles in the right direction.

---

### Phase 3: Server Binary

New `server/` directory, separate Go binary.

**Core loop:**
```
for each tick (e.g., every 50ms = 20 TPS):
    1. drain all pending PlayerInput messages from goroutine-safe queues
    2. apply each input to the corresponding player's state
    3. run CalculatePlayerMovement for each player
    4. run projectile simulation
    5. run enemy AI (currently stub)
    6. run CalculatePlayerProjectileOutcome
    7. serialize WorldSnapshot
    8. broadcast to all connected clients
```

**What moves to the server:**
- `engines/physics-engine/collision.go` — `WorldCollisionSpace`, collision space construction
- `engines/physics-engine/player-movement.go` — `CalculatePlayerMovement()` (after stripping `util.FlipRight/FlipLeft` sprite calls)
- `engines/physics-engine/projectile.go` — `CalculatePlayerProjectileOutcome()`, `FireProjects()`
- `engines/spawn-engine/enemy-spawner.go` — enemy creation
- Enemy AI (to be written)

**Connection management:**
One goroutine per client connection. Reads `PlayerInput` messages and pushes to a buffered channel. The tick loop drains channels each tick.

**Recommended transport for this game style:**
WebSockets (`nhooyr.io/websocket` or `gorilla/websocket`) over TCP. For a RuneScape/RotMG-style game with 20 TPS, TCP reliability is fine — you don't need sub-16ms latency guarantees. WebSockets also work through NAT and firewalls without port forwarding, which matters for early playtesting.

---

### Phase 4: Client Refactor

The existing `game/` package becomes a network client.

**`input-router.go`**: Instead of writing directly to `MainPlayer.Moving`, produce a `PlayerInput` struct and send it to the server. Still apply local input for client-side prediction (optional).

**New `client/connection.go`**: Goroutine that receives `WorldSnapshot` messages from the server and writes to a thread-safe `LatestSnapshot` variable.

**`game-loop.go` Update()**:
```
old: physics → local state
new: read LatestSnapshot → update local entity states for rendering
```

**`draw-world-engine`**: Now renders entities from snapshot state, not from live physics objects. Handles multiple remote players (other connected clients), not just `MainPlayer`.

**Client-side prediction** (optional, for responsiveness):
Apply local player movement immediately without waiting for server confirmation. When the server snapshot arrives, reconcile. This is the technique used by most action games and is what makes networked movement feel responsive. It's complex to implement correctly but not necessary for an initial demo.

---

### Phase 5: Multi-Zone/Map Support

The current map transition (`LoadMap()` triggered by nav tile collision) works fine locally. For multiplayer, zone transitions need server coordination:

- Server tracks which zone each player is in
- Map change triggers player state transfer to different zone's simulation
- Clients receive a "zone change" message and reload map assets client-side
- The `doorId-2` tagging system in `collision.go` is already a reasonable foundation for this

---

### Summary of File-Level Changes

| File | Change Type | What Changes |
|------|------------|--------------|
| `model/player.go` | Refactor | Remove `Draw()`, remove `rl.*` imports, move mouse/facing to input layer, move audio out of `Attack()` |
| `model/enemy.go` | Refactor | Remove `Draw()`, remove `rl.*` imports |
| `model/weapon.go` | Refactor | Remove `Draw()`, remove `rl.*` imports |
| `model/projectile.go` | Refactor | Remove `Draw()`, remove `rl.*` imports |
| `game/input-router.go` | Refactor | Produce serializable `PlayerInput` instead of mutating local player directly |
| `game/game-loop.go` | Split | Server version: pure physics tick. Client version: receive snapshot → update render state |
| `game/game-state.go` | Refactor | Remove local physics state management, add snapshot-driven state |
| `engines/physics-engine/*` | Refactor | Remove any remaining `rl.*` references, make headless-runnable |
| `engines/draw-world-engine` | Extend | Render from snapshot state; handle N remote players; add interpolation between ticks |
| New: `shared/state.go` | New | `PlayerInput`, `PlayerState`, `EnemyState`, `ProjectileState`, `WorldSnapshot` |
| New: `server/main.go` | New | Tick loop, WebSocket listener, per-client goroutines, authoritative simulation |
| New: `client/connection.go` | New | WebSocket client goroutine, snapshot receiver |

---

## 3. How Long Would It Take Claude to Write a Multiplayer Game Engine?

### Scoped to this project (RuneScape/RotMG-style 2D, Go + raylib)

**Writing the code** is not the bottleneck. Claude can produce correct-looking Go code quickly. The actual constraint is the human iteration loop: running it, seeing what breaks, feeding that back.

| Phase | Claude's speed | What slows it down |
|-------|---------------|-------------------|
| Phase 1 refactor (decouple rendering) | Fast — mechanical changes, clear target | You need to verify the game still runs correctly after each change |
| Phase 2 shared protocol | Very fast — simple structs | Design decisions: what state does the server actually need to send? |
| Phase 3 server tick loop | Fast — goroutines are idiomatic Go, Claude knows this pattern well | First desync bugs won't appear until two real clients connect |
| Phase 4 client refactor | Medium — touches many files | Deciding where to put prediction logic; testing feel |
| Phase 5 polish (interpolation, reconnect, lag handling) | Slow — requires real network testing | Only real network conditions reveal timing edge cases |

**Rough estimates for an engaged developer working with Claude:**

- **Runnable demo (2 clients, no prediction, 50ms tick)**: 1-2 weekends
- **Playable foundation (smooth movement, basic lobby, stable)**: 4-8 weeks of evening/weekend work
- **Game on top of the engine**: separate question — that's content work, not engine work

The honest limiting factor is that bugs in multiplayer code are only visible when multiple clients are running simultaneously under real (not localhost) network conditions. Claude can write the fix once you've identified what's wrong, but identifying it requires you to run the thing.

### Specific things Claude is genuinely good at here

- Rewriting the `Draw()` separation mechanically (Phase 1) — straightforward pattern
- Writing the WebSocket server boilerplate in Go — idiomatic, well-tested pattern
- Designing the `WorldSnapshot` protocol struct — can reason about what fields are needed
- Writing the tick loop — fixed-rate game loops are a solved problem
- Debugging specific desync logic once you can describe the symptom precisely

### Specific things that require more human involvement

- Tuning the tick rate and interpolation to feel right — empirical, must play it
- Deciding on client-side prediction architecture — architectural tradeoff with significant downstream consequences
- Stress testing connection drop/reconnect — requires scripted network chaos
- Choosing whether to stay on TCP/WebSocket or move to UDP as the player count grows

### The "unique thumbprint" point from the conversation

This is worth reinforcing: your instinct about building the engine for the game is correct. RuneScape's 600ms tick wasn't a limitation to work around — it *became* the game. The deliberate feel of abilities, movement, and combat all emerged from that constraint. If you build the server with a specific tick rate for your game's feel, that decision propagates through every interaction in the game in ways that Unity or Godot's default networking cannot replicate, because those engines make the tick-rate decision for you.

The engine is the game's skeleton. Building it for one game gives you a skeleton that fits that game exactly.
