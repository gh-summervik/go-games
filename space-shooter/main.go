package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth     = 1280
	screenHeight    = 960
	maxPlayerLasers = 4
)

type Game struct {
	playerImg *ebiten.Image
	enemyImg  *ebiten.Image
	laserImg  *ebiten.Image

	playerX, playerY float64

	playerLasers []Laser
	enemyLasers  []Laser
	enemies      []Enemy

	shootCooldown   int
	enemySpawnTimer int

	gameOver bool

	enemiesDestroyed int
	enemiesEscaped   int
}

type Laser struct {
	x, y float64
}

type Enemy struct {
	x, y float64
}

func overlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw &&
		ax+aw > bx &&
		ay < by+bh &&
		ay+ah > by
}

func (g *Game) spawnEnemy() {
	w := g.enemyImg.Bounds().Dx()
	h := g.enemyImg.Bounds().Dy()
	e := Enemy{
		x: float64(rand.Intn(screenWidth - w)),
		y: -float64(h),
	}
	g.enemies = append(g.enemies, e)
}

func (g *Game) shootPlayer() {
	pw := g.playerImg.Bounds().Dx()
	lw := g.laserImg.Bounds().Dx()
	g.playerLasers = append(g.playerLasers, Laser{
		x: g.playerX + float64(pw/2-lw/2),
		y: g.playerY,
	})
}

func (g *Game) shootEnemy(e *Enemy) {
	ew := g.enemyImg.Bounds().Dx()
	eh := g.enemyImg.Bounds().Dy()
	g.enemyLasers = append(g.enemyLasers, Laser{
		x: e.x + float64(ew)/2,
		y: e.y + float64(eh),
	})
}

func (g *Game) Update() error {
	pw := float64(g.playerImg.Bounds().Dx())
	ph := float64(g.playerImg.Bounds().Dy())
	lw := float64(g.laserImg.Bounds().Dx())
	lh := float64(g.laserImg.Bounds().Dy())
	ew := float64(g.enemyImg.Bounds().Dx())
	eh := float64(g.enemyImg.Bounds().Dy())

	// ---- Restart game ----
	if g.gameOver && ebiten.IsKeyPressed(ebiten.KeyR) {
		*g = Game{
			playerImg:       g.playerImg,
			enemyImg:        g.enemyImg,
			laserImg:        g.laserImg,
			playerX:         screenWidth / 2,
			playerY:         400,
			enemySpawnTimer: 60,
		}
		return nil
	}

	// ---- Freeze game if over ----
	if g.gameOver {
		return nil
	}

	// ---- Player input ----
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.playerX -= 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.playerX += 4
	}
	if g.playerX < 0 {
		g.playerX = 0
	}
	if g.playerX > screenWidth-pw {
		g.playerX = screenWidth - pw
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.shootCooldown == 0 && len(g.playerLasers) < maxPlayerLasers {
		g.shootPlayer()
		g.shootCooldown = 12
	}
	if g.shootCooldown > 0 {
		g.shootCooldown--
	}

	// ---- Move player lasers ----
	for i := range g.playerLasers {
		g.playerLasers[i].y -= 6
	}

	// ---- Spawn enemies ----
	g.enemySpawnTimer--
	if g.enemySpawnTimer <= 0 {
		g.spawnEnemy()
		g.enemySpawnTimer = 60
	}

	// ---- Move enemies + enemy shooting ----
	for i := range g.enemies {
		g.enemies[i].y += 2
		if rand.Intn(90) == 0 {
			g.shootEnemy(&g.enemies[i])
		}
	}

	// ---- Move enemy lasers ----
	for i := range g.enemyLasers {
		g.enemyLasers[i].y += 4
	}

	// ---- Check collisions: player ----
	for _, e := range g.enemies {
		if overlap(e.x, e.y, ew, eh, g.playerX, g.playerY, pw, ph) {
			g.gameOver = true
			return nil
		}
	}
	for _, l := range g.enemyLasers {
		if overlap(l.x, l.y, lw, lh, g.playerX, g.playerY, pw, ph) {
			g.gameOver = true
			return nil
		}
	}

	// ---- Handle enemy hits by lasers and remove lasers ----
	newEnemies := g.enemies[:0]
	newPlayerLasers := g.playerLasers[:0]

	for _, e := range g.enemies {
		hit := false
		for _, l := range g.playerLasers {
			if overlap(l.x, l.y, lw, lh, e.x, e.y, ew, eh) {
				hit = true
				g.enemiesDestroyed++
				break
			}
		}
		if !hit {
			if e.y >= screenHeight {
				g.enemiesEscaped++
			} else {
				newEnemies = append(newEnemies, e)
			}
		}
	}

	// Rebuild playerLasers: remove lasers that hit enemies or went offscreen
	for _, l := range g.playerLasers {
		laserHit := false
		for _, e := range g.enemies {
			if overlap(l.x, l.y, lw, lh, e.x, e.y, ew, eh) {
				laserHit = true
				break
			}
		}
		if !laserHit && l.y > -10 {
			newPlayerLasers = append(newPlayerLasers, l)
		}
	}
	g.playerLasers = newPlayerLasers
	g.enemies = newEnemies

	// ---- Remove offscreen enemy lasers ----
	newEnemyLasers := g.enemyLasers[:0]
	for _, l := range g.enemyLasers {
		if l.y < screenHeight {
			newEnemyLasers = append(newEnemyLasers, l)
		}
	}
	g.enemyLasers = newEnemyLasers

	return nil
}

// func (g *Game) Update() error {
// 	// ---- Restart game ----
// 	if g.gameOver {
// 		// Only listen for restart
// 		if ebiten.IsKeyPressed(ebiten.KeyR) {
// 			*g = Game{
// 				playerImg:       g.playerImg,
// 				enemyImg:        g.enemyImg,
// 				laserImg:        g.laserImg,
// 				playerX:         screenWidth / 2,
// 				playerY:         890,
// 				enemySpawnTimer: 60,
// 			}
// 		}
// 		return nil // freeze everything else
// 	}

// 	pw := float64(g.playerImg.Bounds().Dx())
// 	ph := float64(g.playerImg.Bounds().Dy())
// 	lw := float64(g.laserImg.Bounds().Dx())
// 	lh := float64(g.laserImg.Bounds().Dy())
// 	ew := float64(g.enemyImg.Bounds().Dx())
// 	eh := float64(g.enemyImg.Bounds().Dy())

// 	// ---- Player input (only if alive) ----
// 	if !g.gameOver {
// 		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
// 			g.playerX -= 4
// 		}
// 		if ebiten.IsKeyPressed(ebiten.KeyRight) {
// 			g.playerX += 4
// 		}
// 		if g.playerX < 0 {
// 			g.playerX = 0
// 		}
// 		if g.playerX > screenWidth-pw {
// 			g.playerX = screenWidth - pw
// 		}

// 		if ebiten.IsKeyPressed(ebiten.KeySpace) && g.shootCooldown == 0 && len(g.playerLasers) < maxPlayerLasers {
// 			g.shootPlayer()
// 			g.shootCooldown = 12
// 		}

// 		if g.shootCooldown > 0 {
// 			g.shootCooldown--
// 		}
// 	}

// 	// ---- Move player lasers ----
// 	for i := range g.playerLasers {
// 		g.playerLasers[i].y -= 6
// 	}

// 	// ---- Spawn enemies ----
// 	g.enemySpawnTimer--
// 	if g.enemySpawnTimer <= 0 {
// 		g.spawnEnemy()
// 		g.enemySpawnTimer = 60
// 	}

// 	// ---- Move enemies + enemy shooting ----
// 	for i := range g.enemies {
// 		g.enemies[i].y += 2
// 		if rand.Intn(90) == 0 {
// 			g.shootEnemy(&g.enemies[i])
// 		}
// 	}

// 	// ---- Move enemy lasers ----
// 	for i := range g.enemyLasers {
// 		g.enemyLasers[i].y += 4
// 	}

// 	// ---- Player collisions with enemies ----
// 	for _, e := range g.enemies {
// 		if overlap(e.x, e.y, ew, eh, g.playerX, g.playerY, pw, ph) {
// 			g.gameOver = true
// 			break
// 		}
// 	}

// 	// ---- Player collisions with enemy lasers ----
// 	for _, l := range g.enemyLasers {
// 		if overlap(l.x, l.y, lw, lh, g.playerX, g.playerY, pw, ph) {
// 			g.gameOver = true
// 			break
// 		}
// 	}

// 	// ---- Remove hit enemies ----
// 	newEnemies := g.enemies[:0]
// 	newPlayerLasers := g.playerLasers[:0]

// 	for _, e := range g.enemies {
// 		hit := false
// 		for i, l := range g.playerLasers {
// 			if overlap(l.x, l.y, lw, lh, e.x, e.y, ew, eh) {
// 				hit = true
// 				g.enemiesDestroyed++
// 				g.playerLasers[i].y = -999 // move offscreen
// 				break
// 			}
// 		}
// 		if !hit {
// 			if e.y >= screenHeight {
// 				g.enemiesEscaped++ // count escaped enemies
// 			} else {
// 				newEnemies = append(newEnemies, e)
// 			}
// 		}
// 	}
// 	g.enemies = newEnemies

// 	// ---- Remove offscreen player lasers ----
// 	// newPlayerLasers := g.playerLasers[:0]
// 	for _, l := range g.playerLasers {
// 		if l.y > -10 {
// 			newPlayerLasers = append(newPlayerLasers, l)
// 		}
// 	}
// 	g.playerLasers = newPlayerLasers

// 	// ---- Remove offscreen enemy lasers ----
// 	newEnemyLasers := g.enemyLasers[:0]
// 	for _, l := range g.enemyLasers {
// 		if l.y < screenHeight {
// 			newEnemyLasers = append(newEnemyLasers, l)
// 		}
// 	}
// 	g.enemyLasers = newEnemyLasers

// 	return nil
// }

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	if g.gameOver {
		score := g.enemiesDestroyed - g.enemiesEscaped

		msg := strings.Builder{}
		msg.WriteString("GAME OVER\n")
		msg.WriteString("Enemies destroyed: ")
		msg.WriteString(fmt.Sprintf("%d", g.enemiesDestroyed))
		msg.WriteString("\nEnemies escaped: ")
		msg.WriteString(fmt.Sprintf("%d", g.enemiesEscaped))
		msg.WriteString("\nFinal score: ")
		msg.WriteString(fmt.Sprintf("%d", score))
		msg.WriteString("\nPress R to restart")

		ebitenutil.DebugPrint(screen, msg.String())
		return
	}

	// Player
	op.GeoM.Translate(g.playerX, g.playerY)
	screen.DrawImage(g.playerImg, op)

	// Player lasers
	for _, l := range g.playerLasers {
		op.GeoM.Reset()
		op.GeoM.Translate(l.x, l.y)
		screen.DrawImage(g.laserImg, op)
	}

	// Enemy lasers
	for _, l := range g.enemyLasers {
		op.GeoM.Reset()
		op.GeoM.Translate(l.x, l.y)
		screen.DrawImage(g.laserImg, op)
	}

	// Enemies
	for _, e := range g.enemies {
		op.GeoM.Reset()
		op.GeoM.Translate(e.x, e.y)
		screen.DrawImage(g.enemyImg, op)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/player.png")
	if err != nil {
		log.Fatal(err)
	}
	enemyImg, _, err := ebitenutil.NewImageFromFile("assets/enemy.png")
	if err != nil {
		log.Fatal(err)
	}
	laserImg, _, err := ebitenutil.NewImageFromFile("assets/laser.png")
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		playerImg:       playerImg,
		enemyImg:        enemyImg,
		laserImg:        laserImg,
		playerX:         screenWidth / 2,
		playerY:         890,
		enemySpawnTimer: 60,
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Shooter")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
