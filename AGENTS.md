# Agent Guide — `webtyp/svg`

Constraints for agents modifying this library. Read before any change.

---

## Construction Harness — typed & explicit (the WebTyp approach)

This library is part of WebTyp's **construction harness**: the typed,
explicit API is what keeps an agent that doesn't know the library from building
wrong code. (Ecosystem rationale: the [`api-design` skill](https://github.com/webtyp/devskills/blob/main/skills/api-design/SKILL.md).)

**This library never uses `//go:build`.** It has two unconditional consumer
classes — the WASM browser client, and backend-only programs (`webtyp/ssr`'s
extractor, `assetmin`) that need sprite construction at all times. A tag
inside the library cannot serve both correctly (the extractor would need to
build the library for `!wasm` specifically, which has nothing to do with its
own target). **The split is by Go package, not by build tag** — the consumer
decides what to import, and tags its OWN files.

## Package layout

| Package | Contains | Who imports it |
|---|---|---|
| `webtyp.com/svg` | `type Icon string` + `Render()`. Imports only `webtyp/dom`. | Everyone — reachable from WASM |
| `webtyp.com/svg/sprite` | `Define`, `Sprite`, `Path`, `Raw`, `NewSprite`, `AddRaw`, `AddFile`, JSON (de)serialization. Imports `svg` + `webtyp.com/json` + `webtyp.com/model`. | Backend-only: a consumer's tagged `svg.go`, `webtyp/ssr`, `assetmin`. **Sólo iconos de confianza** |
| `webtyp.com/svg/sanitize` | `Clean` (`//go:build !wasm`). Imports `encoding/xml` + `webtyp.com/fmt`. | Backend-only: limpieza de SVG subidos por terceros (ej. `veltylabs/misitio`). **Nunca usar `sprite` (`viewBoxOf`/`innerOf`) para leer lo que subió un desconocido — ese lector rápido es para iconos de confianza y colaría `<script>`** |

Never move `sprite`'s declarations into the root `svg` package, and never make
the root package depend on `sprite` (dependency points one way: `sprite` →
`svg`). `sanitize` no depende de `sprite` ni del paquete raíz — son dos
caminos separados: `sprite` para iconos de confianza, `sanitize` para archivos
de terceros.

## The two halves of an icon (consumer side)

- **Reference (shared, any target):** `const iconX = svg.Icon("comp-x")` in
  untagged code, rendered with `iconX.Render(class)`. Only that string reaches
  the WASM binary.
- **Definition (consumer's own `!wasm` file):**
  `sprite.Define(iconX, "0 0 16 16", sprite.Path("..."))` inside the
  consumer's `//go:build !wasm` `svg.go`, exposed via
  `IconSvg() *sprite.Sprite`. `webtyp/ssr` extracts it; `assetmin` merges
  all sprites and injects them inline at the top of `<body>` — there is no
  `/assets/icons.svg` URL; `href="#id"` always resolves without a network request.

`Icon.Render` is the ONLY public path to `<svg><use>`. Do not reintroduce
`Svg()`/`Use()` builders here or anywhere else.

## What this library CANNOT guarantee (know the limit)

Because `svg/sprite` is a plain importable Go package (no build tag gate),
NOTHING stops a consumer from importing it into a file reachable by a WASM
build — that mistake compiles successfully and silently grows the binary. This
library cannot prevent that by itself; it is a consumer-side responsibility.
Every consumer AGENTS.md MUST record this mandatory pre-publish check:

```bash
GOOS=js GOARCH=wasm go list -deps ./... | grep webtyp/svg/sprite   # must be empty
```

When reviewing a consumer's `svg.go`, verify it carries `//go:build !wasm`
before it imports `svg/sprite` — that is the consumer's job, not this
library's, but it is the number one mistake this design is vulnerable to.

## Hard rules

- Root `svg` package (`icon.go`) imports ONLY `webtyp.com/dom`. No
  stdlib, no `sprite`.
- `svg/sprite` serializes via `webtyp.com/json` +
  `webtyp.com/model` (`Encodable`/`Decodable`, plus `FielderSlice`
  for the bare-array wire shape `Sprite` needs). **Never** import stdlib
  `encoding/json` here — `webtyp/json` is faster (no reflection, reused
  buffers) and it's the codec the ecosystem's own tests actually exercise, so
  using anything else means this package's serialization bugs go undetected
  by every other test that proves the codec correct. `Sprite` keeps
  `MarshalJSON`/`UnmarshalJSON` method NAMES (stdlib-compatible signatures)
  only because `webtyp/ssr`'s outer envelope decode
  (legitimately `encoding/json`, backend tooling) needs to find them by
  interface — their BODIES delegate entirely to `webtyp/json.Encode`/`Decode`.
- No `map`, no generics, no `any`. Value semantics; unexported struct fields.
- `Path(d)` hardcodes `fill="currentColor"` — icons are color-agnostic; theming
  happens in CSS at the use-site. Never emit fixed colors.
- `viewBox` is mandatory everywhere (`Define`, `AddRaw`, `AddFile`) — no implicit
  default. A symbol rendered in a box it was not drawn for is clipped or
  misaligned, and no default can recover the source coordinate system.
- `AddRaw` (body + explicit viewBox) and `AddFile` (a whole `.svg` file, viewBox
  parsed from its root element) exist for assetmin's raw file/favicon path only;
  they are not the consumer API. Parsing SVG markup belongs HERE, never in
  assetmin — assetmin bundles assets, it does not know the SVG format.
- No `regexp` and no stdlib in the file parser (`file.go`): plain byte scanning.

## Testing / publishing

- `gotest`, never `go test`. Stdlib assertions only (no testify).
- Publish with `gopush 'message'` (local developer only), never raw `git push`.
