# Invasores Go — Minijuego arcade con motor Go + WebAssembly

Minijuego estilo Space Invaders en **React + TypeScript + Tailwind**: el motor
completo (loop, input, física, render en canvas) está escrito en **Go compilado
a WebAssembly** y corre en el navegador a 60 fps.

**Demo en vivo:** https://www.ingeabogado.com/dev/proyectos/invasores-go

## Qué problema resuelve (valor de mercado)

- Juegos y simuladores web sin plugins ni motores pesados: Go compila a un
  solo `.wasm` autocontenido que corre donde haya un navegador moderno.
- La lógica vive en un lenguaje tipado y testeado (`go test`), no en JS
  disperso: colisiones, oleadas y estados con pruebas reales.
- Sirve como plantilla: cambia el paquete `game` y tienes cualquier arcade
  2D (rompebloques, snake, plataformas) con la misma carcasa.

## Lo que hace

- Nave movible con disparo, oleada alienígena que marcha y desciende, balas
  enemigas, puntaje por fila, 3 vidas y niveles con dificultad creciente.
- Pantallas de inicio/pausa/game-over dibujadas en canvas + HUD de puntaje.
- Teclado (flechas/espacio/P, sin secuestrar el scroll fuera del juego) y
  botones táctiles ≥44px en móvil.
- Carga diferida del `.wasm` (~2.3 MB) al pulsar Jugar, con estados de carga,
  error y reintento.
- Responsive móvil→escritorio, canvas con `role="img"` y etiquetas.

## Lo que demuestra (empleadores y clientes)

| Señal que buscan | Dónde verla |
|---|---|
| Go + frontend en un solo entregable | Motor en Go con puente `syscall/js` mínimo (`go/main.go`) |
| Frontera WASM deliberada | Go corre loop+render; JS solo carga (`wasm_exec.js` + bytes) y botones |
| Piensa en producción, no en juguetes | Estados de carga/error/reintento; `wasm_exec.js` acoplado a su versión |
| Tests donde importan | `go test` con 9 casos (colisiones, bordes, oleadas, determinismo) |
| TypeScript estricto | `strict` + `noUnusedLocals`, puente tipado sin `any` |
| Accesibilidad y responsive reales | Botones etiquetados, `role="img"`, 44px táctiles |

## Cómo correrlo

```bash
bun install
bun run dev      # desarrollo en http://localhost:5173 (usa public/game/ ya compilado)
bun run build    # revisa tipos (tsc) y genera dist/
bun run preview  # sirve dist/ local
```

Con npm: `npm install`, `npm run dev`, `npm run build`.

## Cómo recompilar el WASM

Requiere Go 1.27+ (verificado: go1.27.4):

```bash
cd go
go test ./...                                   # lógica pura (host)
GOOS=js GOARCH=wasm go build -o ../public/game/invasores.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../public/game/
```

`public/game/` se commitea a propósito (demo autocontenida: corre sin toolchain).
OJO: `wasm_exec.js` debe ser de la misma versión de Go que compiló el `.wasm`.

## Estructura

```text
InvasoresGo/
├── index.html            # entrada Vite
├── go/                   # módulo Go (invasores)
│   ├── go.mod
│   ├── game/             # lógica pura + tests (sin syscall/js)
│   └── main.go           # puente syscall/js + loop rAF + render canvas
├── public/game/          # build commiteado: .wasm + wasm_exec.js
├── src/
│   ├── main.tsx          # bootstrap React
│   ├── App.tsx           # layout de la demo
│   ├── index.css         # tokens (Tailwind v4) + glass
│   └── invasores/
│       └── Invasores.tsx # vista: canvas + controles + loader WASM
├── Readme_InvasoresGo.md
└── package.json
```

## Autor y licencia

Desarrollado por José Andrés Dávila L. 2026. Licencia MIT (ver `LICENSE`): uso libre con atribución al autor.
