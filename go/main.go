// Command invasores: puente syscall/js + loop + render del minijuego.
// La lógica vive en el paquete game (pura, testeada); aquí solo pegamento DOM.
package main

import (
	"syscall/js"

	"invasores/game"
)

var (
	state  = game.New()
	canvas js.Value
	ctx    js.Value

	rafID    js.Value
	frameFn  js.Func
	keyDnFn  js.Func
	keyUpFn  js.Func
	running  bool
	lastT    float64
	heldL    bool
	heldR    bool
	touchMv  float64
	shootNow bool

	stars [][2]float64
)

func getCanvas() bool {
	doc := js.Global().Get("document")
	canvas = doc.Call("getElementById", "invasores-canvas")
	if canvas.IsNull() || canvas.IsUndefined() {
		return false
	}
	canvas.Set("width", game.W)
	canvas.Set("height", game.H)
	ctx = canvas.Call("getContext", "2d")
	return true
}

func moveInput() float64 {
	m := touchMv
	if heldL {
		m -= 1
	}
	if heldR {
		m += 1
	}
	if m < -1 {
		m = -1
	}
	if m > 1 {
		m = 1
	}
	return m
}

func frame(this js.Value, args []js.Value) any {
	ts := args[0].Float()
	dt := (ts - lastT) / 1000
	lastT = ts
	if dt > 0.05 {
		dt = 0.05
	}
	state.Update(dt, moveInput(), shootNow)
	shootNow = false
	render()
	rafID = js.Global().Call("requestAnimationFrame", frameFn)
	return nil
}

func onKey(down bool) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		code := e.Get("code").String()
		if !down {
			switch code {
			case "ArrowLeft":
				heldL = false
			case "ArrowRight":
				heldR = false
			}
			return nil
		}
		if !running {
			return nil
		}
		switch code {
		case "ArrowLeft", "ArrowRight", "Space":
			e.Call("preventDefault")
		}
		switch code {
		case "ArrowLeft":
			heldL = true
		case "ArrowRight":
			heldR = true
		case "Space":
			if !e.Get("repeat").Bool() {
				if state.Phase == game.Ready || state.Phase == game.Over {
					state.Start()
				} else {
					shootNow = true
				}
			}
		case "KeyP", "Escape":
			if !e.Get("repeat").Bool() {
				state.TogglePause()
			}
		}
		return nil
	})
}

func startFn(this js.Value, args []js.Value) any {
	if running {
		stopLoop()
	}
	if !getCanvas() {
		return "no-canvas"
	}
	seedStars()
	state = game.New()
	state.Start()
	lastT = 0
	win := js.Global()
	keyDnFn = onKey(true)
	keyUpFn = onKey(false)
	win.Call("addEventListener", "keydown", keyDnFn)
	win.Call("addEventListener", "keyup", keyUpFn)
	frameFn = js.FuncOf(frame)
	running = true
	rafID = win.Call("requestAnimationFrame", frameFn)
	return "ok"
}

func stopLoop() {
	if !running {
		return
	}
	running = false
	js.Global().Call("cancelAnimationFrame", rafID)
	win := js.Global()
	win.Call("removeEventListener", "keydown", keyDnFn)
	win.Call("removeEventListener", "keyup", keyUpFn)
	frameFn.Release()
	keyDnFn.Release()
	keyUpFn.Release()
	heldL, heldR, touchMv, shootNow = false, false, 0, false
}

func seedStars() {
	stars = stars[:0]
	seed := uint32(1234567)
	next := func() float64 {
		seed = seed*1664525 + 1013904223
		return float64(seed>>8) / float64(1<<24)
	}
	for i := 0; i < 70; i++ {
		stars = append(stars, [2]float64{next() * game.W, next() * game.H})
	}
}

var rowColors = []string{"#e05c8a", "#f0a500", "#00d4aa", "#4a90d9", "#9f7aea", "#ff6b35"}

func text(s string, x, y float64, size int, color, align string) {
	ctx.Set("font", "bold "+itoa(size)+"px ui-monospace, monospace")
	ctx.Set("fillStyle", color)
	ctx.Set("textAlign", align)
	ctx.Call("fillText", s, x, y)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func render() {
	ctx.Set("fillStyle", "#0b1020")
	ctx.Call("fillRect", 0, 0, game.W, game.H)
	ctx.Set("fillStyle", "rgba(255,255,255,0.5)")
	for _, st := range stars {
		ctx.Call("fillRect", st[0], st[1], 2, 2)
	}
	// aliens
	for i := range state.Aliens {
		a := &state.Aliens[i]
		if !a.Alive {
			continue
		}
		ctx.Set("fillStyle", rowColors[a.Row%len(rowColors)])
		ctx.Call("fillRect", a.X, a.Y, game.AlienW, game.AlienH)
		ctx.Set("fillStyle", "#0b1020")
		ctx.Call("fillRect", a.X+7, a.Y+6, 5, 5)
		ctx.Call("fillRect", a.X+game.AlienW-12, a.Y+6, 5, 5)
	}
	// balas
	for _, b := range state.Bullets {
		if b.Dead {
			continue
		}
		ctx.Set("fillStyle", "#00d4aa")
		ctx.Call("fillRect", b.X, b.Y, game.BulletW, game.BulletH)
	}
	for _, b := range state.EBullets {
		ctx.Set("fillStyle", "#ff6b35")
		ctx.Call("fillRect", b.X, b.Y, game.EBulletW, game.EBulletH)
	}
	// nave (triángulo)
	px := state.PlayerX
	py := game.PlayerY
	ctx.Set("fillStyle", "#f0f0f8")
	ctx.Call("beginPath")
	ctx.Call("moveTo", px, py-game.PlayerH/2)
	ctx.Call("lineTo", px-game.PlayerW/2, py+game.PlayerH/2)
	ctx.Call("lineTo", px+game.PlayerW/2, py+game.PlayerH/2)
	ctx.Call("closePath")
	ctx.Call("fill")
	// HUD
	text("SCORE "+itoa(state.Score), 16, 28, 16, "#f0f0f8", "left")
	text("NIVEL "+itoa(state.Level), game.W/2, 28, 16, "#f0f0f8", "center")
	text("VIDAS "+itoa(state.Lives), game.W-16, 28, 16, "#f0f0f8", "right")
	// overlays
	switch state.Phase {
	case game.Ready:
		text("INVASORES GO", game.W/2, game.H/2-30, 32, "#00d4aa", "center")
		text("Pulsa Jugar o ESPACIO", game.W/2, game.H/2+10, 16, "#f0f0f8", "center")
		text("Flechas mover · Espacio disparar · P pausa", game.W/2, game.H/2+36, 13, "#9898b8", "center")
	case game.Paused:
		ctx.Set("fillStyle", "rgba(0,0,0,0.6)")
		ctx.Call("fillRect", 0, 0, game.W, game.H)
		text("PAUSA", game.W/2, game.H/2, 32, "#f0f0f8", "center")
	case game.Over:
		ctx.Set("fillStyle", "rgba(0,0,0,0.65)")
		ctx.Call("fillRect", 0, 0, game.W, game.H)
		text("GAME OVER", game.W/2, game.H/2-10, 32, "#e05c8a", "center")
		text("Puntaje "+itoa(state.Score)+" · Pulsa Reiniciar", game.W/2, game.H/2+24, 15, "#f0f0f8", "center")
	}
}

func main() {
	js.Global().Set("invasores", map[string]any{
		"start": js.FuncOf(func(this js.Value, args []js.Value) any { return startFn(this, args) }),
		"stop": js.FuncOf(func(this js.Value, args []js.Value) any {
			stopLoop()
			return nil
		}),
		"pause": js.FuncOf(func(this js.Value, args []js.Value) any {
			state.TogglePause()
			return nil
		}),
		"setMove": js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) > 0 {
				touchMv = args[0].Float()
			}
			return nil
		}),
		"shoot": js.FuncOf(func(this js.Value, args []js.Value) any {
			if state.Phase == game.Ready || state.Phase == game.Over {
				state.Start()
			} else {
				shootNow = true
			}
			return nil
		}),
	})
	select {} // el runtime Go vive mientras la página viva
}
