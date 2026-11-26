// Invasores Go: minijuego arcade con motor Go+WASM. React aporta página,
// botones y puente de carga; el loop, input y render viven en Go.
import { useCallback, useEffect, useRef, useState } from "react"

declare global {
  interface Window {
    Go?: new () => { importObject: WebAssembly.Imports; run: (i: WebAssembly.Instance) => Promise<void> }
    invasores?: {
      start: () => string | undefined
      stop: () => void
      pause: () => void
      setMove: (v: number) => void
      shoot: () => void
    }
  }
}

type Phase = "idle" | "loading" | "playing" | "error"

async function loadScriptOnce(src: string): Promise<void> {
  if (document.querySelector(`script[data-invasores]`) || window.Go) return
  await new Promise<void>((resolve, reject) => {
    const el = document.createElement("script")
    el.src = src
    el.async = true
    el.dataset.invasores = "1"
    el.onload = () => resolve()
    el.onerror = () => reject(new Error("No se pudo cargar el soporte WASM de Go."))
    document.head.appendChild(el)
  })
}

export function Invasores({ gameBase }: { gameBase: string }) {
  const goRef = useRef<InstanceType<NonNullable<Window["Go"]>> | null>(null)
  const readyRef = useRef<Promise<void> | null>(null)
  const [phase, setPhase] = useState<Phase>("idle")
  const [status, setStatus] = useState("Pulsa Jugar para descargar el juego (~2.3 MB) y empezar.")
  const [error, setError] = useState<string | null>(null)

  const ensureReady = useCallback(() => {
    if (!readyRef.current) {
      readyRef.current = (async () => {
        await loadScriptOnce(`${gameBase}/wasm_exec.js`)
        if (!window.Go) throw new Error("Soporte WASM de Go no disponible.")
        const bytes = await fetch(`${gameBase}/invasores.wasm`).then((r) => {
          if (!r.ok) throw new Error("No se pudo descargar el juego.")
          return r.arrayBuffer()
        })
        const go = new window.Go()
        const { instance } = await WebAssembly.instantiate(bytes, go.importObject)
        goRef.current = go
        void go.run(instance)
      })().catch((e: unknown) => {
        readyRef.current = null
        throw e
      })
    }
    return readyRef.current
  }, [gameBase])

  useEffect(() => {
    return () => {
      try {
        window.invasores?.stop()
      } catch {
        /* noop */
      }
    }
  }, [])

  const play = useCallback(async () => {
    setError(null)
    setPhase("loading")
    setStatus("Descargando juego (~2.3 MB, solo la primera vez)…")
    try {
      await ensureReady()
      // Ceder un tick para que el runtime Go publique el puente.
      for (let i = 0; i < 50 && !window.invasores; i++) {
        await new Promise((r) => setTimeout(r, 100))
      }
      if (!window.invasores) throw new Error("El juego no respondió a tiempo.")
      window.invasores.start()
      setPhase("playing")
      setStatus("Flechas o botones para moverte · Espacio o ● para disparar · P pausa.")
    } catch (e) {
      setError(e instanceof Error ? e.message : "No se pudo iniciar.")
      setPhase("error")
    }
  }, [ensureReady])

  const pause = useCallback(() => {
    try {
      window.invasores?.pause()
    } catch {
      /* noop */
    }
  }, [])

  const move = useCallback(
    (v: number) => () => {
      try {
        window.invasores?.setMove(v)
      } catch {
        /* noop */
      }
    },
    [],
  )

  const shoot = useCallback(() => {
    try {
      window.invasores?.shoot()
    } catch {
      /* noop */
    }
  }, [])

  const busy = phase === "loading"

  return (
    <div className="flex flex-col gap-5">
      <div>
        <a
          href="https://www.ingeabogado.com/dev/proyectos/invasores-go"
          className="text-sm text-muted-foreground underline underline-offset-4"
        >
          ← Ver en ingeabogado.com
        </a>
        <h1 className="text-3xl md:text-5xl font-bold mt-2 mb-2" style={{ color: "var(--module-dev)" }}>
          Invasores Go
        </h1>
        <p className="text-muted-foreground max-w-2xl">
          Minijuego arcade estilo Space Invaders con el motor escrito en Go y compilado a
          WebAssembly: oleadas, puntaje, vidas y niveles, a 60 fps en canvas.
        </p>
        <p className="text-xs text-muted-foreground mt-2">
          Go · WebAssembly · Canvas · Desarrollado por José Andrés Dávila L. 2026. Licencia libre.
        </p>
      </div>

      <div className="glass rounded-2xl p-4 md:p-5 flex flex-col gap-3">
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => void play()}
            disabled={busy}
            className="rounded-xl min-h-[44px] px-6 text-sm font-bold touch-manipulation disabled:opacity-40"
            style={{ background: "var(--module-dev)", color: "#fff" }}
          >
            {phase === "loading" ? "Cargando…" : phase === "playing" ? "↻ Reiniciar" : "▶ Jugar"}
          </button>
          <button
            type="button"
            onClick={pause}
            disabled={phase !== "playing"}
            className="rounded-xl min-h-[44px] px-4 text-sm font-medium glass touch-manipulation disabled:opacity-40"
          >
            Pausar
          </button>
        </div>
        <p className="text-xs text-muted-foreground" role="status">
          {status}
        </p>
      </div>

      {error && (
        <div role="alert" className="glass rounded-2xl p-4 flex flex-col sm:flex-row gap-3 sm:items-center">
          <p className="text-sm text-foreground flex-1">{error}</p>
          <button
            type="button"
            onClick={() => void play()}
            disabled={busy}
            className="rounded-xl min-h-[44px] px-5 text-sm font-medium glass touch-manipulation disabled:opacity-40"
          >
            Reintentar
          </button>
        </div>
      )}

      <div className="glass rounded-2xl p-4 md:p-6">
        <canvas
          id="invasores-canvas"
          width={480}
          height={640}
          className="w-full h-auto rounded-xl mx-auto"
          style={{ maxWidth: 480, touchAction: "none", background: "#0b1020" }}
          role="img"
          aria-label="Juego Invasores Go: defiende tu nave de la oleada alienígena."
        />
        <div className="flex md:hidden justify-between items-center gap-3 mt-4">
          <div className="flex gap-3">
            <button
              type="button"
              aria-label="Mover a la izquierda"
              disabled={phase !== "playing"}
              onPointerDown={move(-1)}
              onPointerUp={move(0)}
              onPointerCancel={move(0)}
              onPointerLeave={move(0)}
              className="rounded-xl w-16 h-16 text-2xl glass touch-manipulation disabled:opacity-40 select-none"
            >
              ◀
            </button>
            <button
              type="button"
              aria-label="Mover a la derecha"
              disabled={phase !== "playing"}
              onPointerDown={move(1)}
              onPointerUp={move(0)}
              onPointerCancel={move(0)}
              onPointerLeave={move(0)}
              className="rounded-xl w-16 h-16 text-2xl glass touch-manipulation disabled:opacity-40 select-none"
            >
              ▶
            </button>
          </div>
          <button
            type="button"
            aria-label="Disparar"
            disabled={phase !== "playing"}
            onPointerDown={shoot}
            className="rounded-full w-20 h-20 text-2xl font-bold touch-manipulation disabled:opacity-40 select-none"
            style={{ background: "var(--module-dev)", color: "#fff" }}
          >
            ●
          </button>
        </div>
        <p className="hidden md:block text-xs text-muted-foreground mt-3 text-center">
          Teclado: ← → moverse · Espacio disparar · P pausa.
        </p>
      </div>

      <p className="text-xs text-muted-foreground">
        Motor: Go + WebAssembly en tu navegador · Demo educativa.
        Desarrollado por José Andrés Dávila L. 2026. Licencia libre.
      </p>
    </div>
  )
}
