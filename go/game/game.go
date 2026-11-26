// Package game: lógica pura del minijuego (sin syscall/js, testeable en host).
package game

import (
	"math/rand"
	"time"
)

const (
	W = 480
	H = 640

	PlayerW     = 36.0
	PlayerH     = 20.0
	PlayerY     = H - 60.0
	PlayerSpeed = 280.0

	BulletW     = 4.0
	BulletH     = 12.0
	BulletSpeed = 460.0

	AlienW = 30.0
	AlienH = 22.0

	EBulletW     = 4.0
	EBulletH     = 10.0
	EBulletSpeed = 220.0

	ShootCD = 0.22
)

type Phase string

const (
	Ready   Phase = "ready"
	Playing Phase = "playing"
	Paused  Phase = "paused"
	Over    Phase = "over"
)

type Bullet struct {
	X, Y  float64
	Enemy bool
	Dead  bool
}

type Alien struct {
	X, Y  float64
	Row   int
	Alive bool
}

type State struct {
	Phase    Phase
	Level    int
	Score    int
	Lives    int
	PlayerX  float64
	Dir      float64
	Bullets  []Bullet
	Aliens   []Alien
	EBullets []Bullet
	cd       float64
	enemyCD  float64
	rng      *rand.Rand
}

func New() *State { return NewWithSeed(time.Now().UnixNano()) }

func NewWithSeed(seed int64) *State {
	s := &State{
		Phase:   Ready,
		Level:   1,
		Lives:   3,
		PlayerX: W / 2,
		Dir:     1,
		rng:     rand.New(rand.NewSource(seed)),
	}
	s.spawnWave()
	return s
}

func (s *State) aliveCount() int {
	n := 0
	for i := range s.Aliens {
		if s.Aliens[i].Alive {
			n++
		}
	}
	return n
}

func (s *State) spawnWave() {
	rows := 4
	if s.Level >= 3 {
		rows = 5
	}
	if s.Level >= 6 {
		rows = 6
	}
	cols := 8
	s.Aliens = s.Aliens[:0]
	gapX := (W - 40*2 - float64(cols)*AlienW) / float64(cols-1)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			s.Aliens = append(s.Aliens, Alien{
				X:     40 + float64(c)*(AlienW+gapX),
				Y:     70 + float64(r)*(AlienH+14),
				Row:   r,
				Alive: true,
			})
		}
	}
	s.Dir = 1
	s.EBullets = s.EBullets[:0]
	s.Bullets = s.Bullets[:0]
}

func overlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}

func alienSpeed(s *State) float64 {
	frac := 0.0
	if total := len(s.Aliens); total > 0 {
		frac = 1 - float64(s.aliveCount())/float64(total)
	}
	return 30 + float64(s.Level-1)*8 + frac*70
}

// Start (re)inicia la partida.
func (s *State) Start() {
	s.Level, s.Score, s.Lives = 1, 0, 3
	s.PlayerX = W / 2
	s.spawnWave()
	s.Phase = Playing
}

func (s *State) TogglePause() {
	if s.Phase == Playing {
		s.Phase = Paused
	} else if s.Phase == Paused {
		s.Phase = Playing
	}
}

func (s *State) nextLevel() {
	s.Level++
	s.spawnWave()
	s.PlayerX = W / 2
}

func (s *State) loseLife() {
	s.Lives--
	s.EBullets = s.EBullets[:0]
	s.Bullets = s.Bullets[:0]
	s.PlayerX = W / 2
	if s.Lives <= 0 {
		s.Phase = Over
	}
}

// Update avanza dt segundos. move ∈ [-1,1]; shoot dispara si el cooldown lo permite.
func (s *State) Update(dt, move float64, shoot bool) {
	if s.Phase != Playing {
		return
	}
	s.PlayerX += move * PlayerSpeed * dt
	half := PlayerW / 2
	if s.PlayerX < half {
		s.PlayerX = half
	}
	if s.PlayerX > W-half {
		s.PlayerX = W - half
	}
	s.cd -= dt
	if shoot && s.cd <= 0 {
		s.cd = ShootCD
		s.Bullets = append(s.Bullets, Bullet{X: s.PlayerX - BulletW/2, Y: PlayerY - PlayerH/2 - BulletH})
	}
	kept := s.Bullets[:0]
	for _, b := range s.Bullets {
		b.Y -= BulletSpeed * dt
		if b.Y+BulletH > 0 {
			kept = append(kept, b)
		}
	}
	s.Bullets = kept

	sp := alienSpeed(s) * s.Dir * dt
	hitEdge := false
	for i := range s.Aliens {
		a := &s.Aliens[i]
		if !a.Alive {
			continue
		}
		a.X += sp
		if a.X < 8 || a.X+AlienW > W-8 {
			hitEdge = true
		}
	}
	if hitEdge {
		s.Dir = -s.Dir
		for i := range s.Aliens {
			if s.Aliens[i].Alive {
				s.Aliens[i].Y += 18
			}
		}
	}

	s.enemyCD -= dt
	if s.enemyCD <= 0 {
		interval := 1.1 - float64(s.Level-1)*0.12
		if interval < 0.35 {
			interval = 0.35
		}
		s.enemyCD = interval
		byCol := map[int]int{}
		for i := range s.Aliens {
			a := &s.Aliens[i]
			if !a.Alive {
				continue
			}
			c := int((a.X + AlienW/2) / W * 12)
			if j, ok := byCol[c]; !ok || a.Y > s.Aliens[j].Y {
				byCol[c] = i
			}
		}
		if len(byCol) > 0 {
			keys := make([]int, 0, len(byCol))
			for c := range byCol {
				keys = append(keys, c)
			}
			a := &s.Aliens[byCol[keys[s.rng.Intn(len(keys))]]]
			s.EBullets = append(s.EBullets, Bullet{X: a.X + AlienW/2 - EBulletW/2, Y: a.Y + AlienH, Enemy: true})
		}
	}
	keptE := s.EBullets[:0]
	for _, b := range s.EBullets {
		b.Y += EBulletSpeed * dt
		if b.Y < H {
			keptE = append(keptE, b)
		}
	}
	s.EBullets = keptE

	for bi := range s.Bullets {
		b := &s.Bullets[bi]
		if b.Dead {
			continue
		}
		for ai := range s.Aliens {
			a := &s.Aliens[ai]
			if !a.Alive {
				continue
			}
			if overlap(b.X, b.Y, BulletW, BulletH, a.X, a.Y, AlienW, AlienH) {
				b.Dead = true
				a.Alive = false
				s.Score += (6 - a.Row) * 10
				break
			}
		}
	}
	kept2 := s.Bullets[:0]
	for _, b := range s.Bullets {
		if !b.Dead {
			kept2 = append(kept2, b)
		}
	}
	s.Bullets = kept2

	px0, py0 := s.PlayerX-half, PlayerY-PlayerH/2
	hit := false
	for _, b := range s.EBullets {
		if overlap(b.X, b.Y, EBulletW, EBulletH, px0, py0, PlayerW, PlayerH) {
			hit = true
			break
		}
	}
	if hit {
		s.loseLife()
	}
	for i := range s.Aliens {
		if s.Aliens[i].Alive && s.Aliens[i].Y+AlienH >= py0 {
			s.Phase = Over
			return
		}
	}
	if s.aliveCount() == 0 {
		s.nextLevel()
	}
}
