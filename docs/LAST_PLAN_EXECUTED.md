---
PLAN: "feat: support external sprite URL via SetSpriteURL and webtyp-sprite meta tag"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan: `svg` con sprite externo (`sprite.<hash>.svg`)

Master: `SECURITY_OBSERVABILITY_MASTER_PLAN.md` (D3, O3b-2).

## Tools

```bash
export GOPRIVATE=webtyp.com/*,github.com/veltylabs/*
export GOFLAGS=-mod=mod
go install webtyp.com/devflow/cmd/gotest@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Ecosystem search

- Buscado en `~/Dev/Project/webtyp` (`svg`, `icons`, `sitec`, `components`, `layout`).
- Todos los íconos del ecosistema pasan por `svg.Icon.Render` (`svg/icon.go:18`), que hoy emite `<use href="#id"/>`.
- **Decisión:** extender `webtyp.com/svg` para resolver la URL del sprite externo en una sola pieza, manteniendo intactos todos los call sites en `icons`, `components` y `layout`.

## Design gate

1. **Prior art:**
   - [MDN — SVG `<use>`](https://developer.mozilla.org/en-US/docs/Web/SVG/Reference/Element/use): `<use href="/sprite.<hash>.svg#id"/>` referencia símbolos de un archivo SVG externo del mismo origen sin incrustar `<svg>` en el HTML.
   - [web.dev — Strict CSP](https://web.dev/articles/strict-csp): sacar el SVG del HTML evita estilos en línea y permite cachear el sprite de forma inmutable por hash.
   - [Vite / webpack SVG sprite plugins](https://vitejs.dev/guide/assets.html): emiten un archivo con hash en el nombre y prefijan `href` con su ruta.
2. **Novice-name test:** `svg.SetSpriteURL(url string)` y `svg.SpriteMetaName = "webtyp-sprite"` — nombres literales que dicen exactamente qué configuran y qué `<meta>` lee el navegador.
3. **Complexity ledger:**
   - Conceptos a aprender: `+1` (`SetSpriteURL`, solo lo llama `sitec` en SSR) / `-0`
   - Archivos a tocar en módulos/componentes: `0`
   - Líneas en el call site (`icons`, `components`, `layout`): `0`
   - Formas de hacer lo mismo: `0` (`Icon.Render` sigue siendo el único camino).
4. **Where it belongs:** `webtyp.com/svg` posee `Icon.Render`. Poner la resolución en `svg` evita acoplar `components`, `layout` e `icons` a `sitec` o al DOM del `<head>`.
5. **What does this change delete?** Nada nuevo se duplica; reemplaza el prefijo fijo `"#"` por `spriteURL() + "#"` cuando hay sprite externo.

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `dom.NewElement` | `func NewElement(tag string) *Element` | `v0.13.29` | `webtyp.com/dom` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Reglas de entrega y anti-footguns

- `svg` compila a WASM con TinyGo: **no usar `map` ni stdlib `strings`/`fmt`** en archivos compartidos o `wasm`.
- Hacer commit después de cada etapa. No usar `git reset`, `git restore`, `git checkout -- <file>` ni `git stash`.

---

## Etapa 1: Diagrama y pruebas (TDD)

1. Crear `docs/diagrams/SPRITE_URL.md` (contexto, `flowchart TD`, walkthrough numerado):
   - SSR (`!wasm`): `sitec` llama `svg.SetSpriteURL("/sprite.<hash>.svg")` → `Icon.Render` emite `<use href="/sprite.<hash>.svg#id"/>`.
   - Navegador (`wasm`): primera llamada a `Icon.Render` lee `<meta name="webtyp-sprite" content="/sprite.<hash>.svg">` → cachea URL → emite `<use href="/sprite.<hash>.svg#id"/>`.
   - Sin URL (tests unitarios aislados): emite `<use href="#id"/>`.
2. Crear `tests/sprite_url_test.go` y registrar en `docs/TESTS.md` (`| [TestName](../tests/sprite_url_test.go) | regla |`):
   - `TestIconRender_DefaultHashOnly`: con `SetSpriteURL("")`, `Icon("check").Render().String()` contiene `href="#check"`.
   - `TestIconRender_ExternalSpriteURL`: con `SetSpriteURL("/sprite.abc123.svg")`, `Icon("check").Render().String()` contiene `href="/sprite.abc123.svg#check"`.
   - `TestSprite_ExternalXMLNSAndNoInlineStyle`: `sprite.NewSprite(def).String()` y `sprite.EmptyWrapper` empiezan con `<svg xmlns="http://www.w3.org/2000/svg">` y NO contienen `style=`.
   - Actualizar `sprite/file_test.go:14` al nuevo wrapper `<svg xmlns="http://www.w3.org/2000/svg">`.

## Etapa 2: Implementar `SetSpriteURL`, `<meta>` en WASM y wrapper SVG externo sin `style=""`

1. En `sprite/sprite.go:78` y `sprite/merge.go:7`:
   - Cambiar `<svg aria-hidden="true" style="display:none">` por `<svg xmlns="http://www.w3.org/2000/svg">` (obligatorio para que el navegador parsee un archivo `.svg` externo en `<use href="/sprite.<hash>.svg#id"/>` y sin `style=""` en línea bajo CSP `style-src 'self'`).
2. Crear `sprite_url.go`:
```go
package svg

// SpriteMetaName is the <meta name="..."> emitted by sitec with the external sprite URL.
const SpriteMetaName = "webtyp-sprite"

var configuredSpriteURL string

// SetSpriteURL sets the external sprite file URL (for example "/sprite.a1b2c3d4.svg")
// prefixed before "#<id>" by Icon.Render. Pass "" to emit bare "#<id>" references.
func SetSpriteURL(url string) {
	configuredSpriteURL = url
}
```
2. Crear `sprite_url_back.go` (`//go:build !wasm`):
```go
//go:build !wasm

package svg

func currentSpriteURL() string { return configuredSpriteURL }
```
3. Crear `sprite_url_front.go` (`//go:build wasm`):
```go
//go:build wasm

package svg

import "syscall/js"

var spriteMetaLoaded bool

func currentSpriteURL() string {
	if configuredSpriteURL != "" || spriteMetaLoaded {
		return configuredSpriteURL
	}
	doc := js.Global().Get("document")
	if doc.IsUndefined() || doc.IsNull() {
		return ""
	}
	meta := doc.Call("querySelector", `meta[name="`+SpriteMetaName+`"]`)
	if meta.IsNull() || meta.IsUndefined() {
		return ""
	}
	spriteMetaLoaded = true
	configuredSpriteURL = meta.Get("content").String()
	return configuredSpriteURL
}
```
4. En `icon.go:22`, cambiar:
   `Child(dom.NewElement("use").Attr("href", currentSpriteURL()+"#"+string(i)))`

## Etapa 3: Documentación y entrega

1. Actualizar `README.md` (≤ 150 líneas, 4 secciones obligatorias), `docs/ARCHITECTURE.md` (4 secciones) y `docs/TESTS.md`.
2. Ejecutar `gotest`.

## Tabla de etapas

| Etapa | Archivos | Qué |
|---|---|---|
| 1 | `docs/diagrams/SPRITE_URL.md`, `tests/sprite_url_test.go`, `docs/TESTS.md` | Diagrama y pruebas TDD |
| 2 | `sprite_url.go`, `sprite_url_back.go`, `sprite_url_front.go`, `icon.go` | Resolución de URL de sprite en SSR y WASM |
| 3 | `README.md`, `docs/ARCHITECTURE.md`, `docs/TESTS.md` | Documentación y `gotest` verde |
