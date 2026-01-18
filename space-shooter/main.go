package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	playerImg *ebiten.Image
	enemyImg  *ebiten.Image
	laserImg  *ebiten.Image

	playerX, playerY float64

	lasers []Laser

	shootCooldown int
}

type Laser struct {
	x, y float64
}

func (g *Game) shoot() {
	laser := Laser{
		x: g.playerX + float64(g.playerImg.Bounds().Dx()/2) - 2, // center bullet
		y: g.playerY,                                            // - float64(g.playerImg.Bounds().Dx()),
	}
	g.lasers = append(g.lasers, laser)
}

func (g *Game) removeOffscreenLasers() {
	newLasers := g.lasers[:0]
	for _, b := range g.lasers {
		if b.y > -10 { // keep a small buffer
			newLasers = append(newLasers, b)
		}
	}
	g.lasers = newLasers
}

func (g *Game) Update() error {
	const speed = 4

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.playerX -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.playerX += speed
	}

	w := g.playerImg.Bounds().Dx()

	if g.playerX < 0 {
		g.playerX = 0
	}
	if g.playerX > 640-float64(w) {
		g.playerX = 640 - float64(w)
	}

	// SHOOTING with cooldown
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if g.shootCooldown <= 0 {
			g.shoot()
			g.shootCooldown = 10
		}
	}

	if g.shootCooldown > 0 {
		g.shootCooldown--
	}

	// Move bullets upward
	for i := range g.lasers {
		g.lasers[i].y -= 6
	}

	// Remove bullets that are off-screen
	g.removeOffscreenLasers()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.playerX, g.playerY)

	screen.DrawImage(g.playerImg, op)

	for _, b := range g.lasers {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(b.x, b.y)
		screen.DrawImage(g.laserImg, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 640, 480
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
		playerImg: playerImg,
		enemyImg:  enemyImg,
		laserImg:  laserImg,
		playerX:   320,
		playerY:   400,
	}

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Space Shooter")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
