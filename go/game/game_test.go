package game

import "testing"

func TestNewSpawnsWave(t *testing.T) {
	s := NewWithSeed(1)
	if s.Phase != Ready || s.Level != 1 || s.Lives != 3 || s.Score != 0 {
		t.Fatalf("estado inicial malo: %+v", s)
	}
	if got := s.aliveCount(); got != 32 {
		t.Fatalf("aliens vivos = %d, quiero 32", got)
	}
}

func TestStartResets(t *testing.T) {
	s := NewWithSeed(1)
	s.Score, s.Lives, s.Level = 999, 1, 5
	s.Start()
	if s.Phase != Playing || s.Score != 0 || s.Lives != 3 || s.Level != 1 {
		t.Fatalf("Start no reinicia: %+v", s)
	}
}

func TestPlayerClamp(t *testing.T) {
	s := NewWithSeed(1)
	s.Start()
	s.Update(5, -1, false)
	if s.PlayerX != PlayerW/2 {
		t.Fatalf("borde izq = %v", s.PlayerX)
	}
	s.Update(5, 1, false)
	if s.PlayerX != W-PlayerW/2 {
		t.Fatalf("borde der = %v", s.PlayerX)
	}
}

func TestShootCooldown(t *testing.T) {
	s := NewWithSeed(1)
	s.Start()
	s.Update(0.01, 0, true)
	s.Update(0.01, 0, true)
	if len(s.Bullets) != 1 {
		t.Fatalf("balas = %d, quiero 1 (cooldown)", len(s.Bullets))
	}
	s.Update(0.25, 0, true)
	if len(s.Bullets) != 2 {
		t.Fatalf("balas = %d, quiero 2", len(s.Bullets))
	}
}

func TestBulletKillsAlien(t *testing.T) {
	s := NewWithSeed(1)
	s.Start()
	s.Aliens = []Alien{
		{X: s.PlayerX - AlienW/2, Y: 100, Row: 0, Alive: true},
		{X: 10, Y: 300, Alive: true}, // testigo: evita nextLevel
	}
	s.Bullets = []Bullet{{X: s.PlayerX - BulletW/2, Y: 105}}
	s.Update(0.01, 0, false)
	if s.Aliens[0].Alive {
		t.Fatal("el alien sigue vivo")
	}
	if len(s.Bullets) != 0 {
		t.Fatal("la bala debió consumirse")
	}
	if s.Score != 60 {
		t.Fatalf("score = %d, quiero 60", s.Score)
	}
}

func TestEnemyBulletHitsPlayer(t *testing.T) {
	s := NewWithSeed(1)
	s.Start()
	py0 := PlayerY - PlayerH/2
	s.EBullets = []Bullet{{X: s.PlayerX, Y: py0, Enemy: true}}
	s.Update(0.01, 0, false)
	if s.Lives != 2 {
		t.Fatalf("vidas = %d, quiero 2", s.Lives)
	}
}

func TestWaveClearNextLevel(t *testing.T) {
	s := NewWithSeed(1)
	s.Start()
	for i := range s.Aliens {
		s.Aliens[i].Alive = false
	}
	s.Update(0.01, 0, false)
	if s.Level != 2 {
		t.Fatalf("nivel = %d, quiero 2", s.Level)
	}
	if s.aliveCount() == 0 {
		t.Fatal("debió respawnear oleada")
	}
}

func TestInvasionEndsGame(t *testing.T) {
	s := NewWithSeed(1)
	s.Start()
	s.Aliens = []Alien{{X: 10, Y: PlayerY, Alive: true}}
	s.Update(0.01, 0, false)
	if s.Phase != Over {
		t.Fatalf("fase = %s, quiero over", s.Phase)
	}
}

func TestDeterministicSeed(t *testing.T) {
	run := func() (int, int, float64) {
		s := NewWithSeed(42)
		s.Start()
		for i := 0; i < 60; i++ {
			s.Update(1.0/60, 0, i%10 == 0)
		}
		return s.Score, len(s.EBullets), s.Aliens[0].X
	}
	s1, e1, x1 := run()
	s2, e2, x2 := run()
	if s1 != s2 || e1 != e2 || x1 != x2 {
		t.Fatalf("no determinista: (%d,%d,%v) vs (%d,%d,%v)", s1, e1, x1, s2, e2, x2)
	}
}
