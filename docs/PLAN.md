---
PLAN: "feat(sprite): AddLayers extracts Inkscape layers and variants into individual symbols"
TAG: v0.0.12
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 145096485109354402
PR: https://github.com/webtyp/svg/pull/9
REVIEWED_COMMIT: 31969bccd02034a20983723f2bb58201e3d9292f
---

> Read AGENTS.md at the repo root before any change.

# Plan — `webtyp/svg` v0.0.12: `(*Sprite).AddLayers` for modular Inkscape SVGs

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Design gate

### 1. Prior art
- `webtyp/svg/sprite/file.go:12`: `(*Sprite).AddFile(id, svgFile)` imports whole SVGs as 1 single monolithic `<symbol>`.
- In modular SVG illustrations and avatars (Inkscape), drawings are composed of independent layers and named variants (`inkscape:groupmode="layer"`, `<g id="...">`). Monolithic `<symbol>`s prevent styling or swapping child nodes due to `<use>` Shadow DOM encapsulation.
- `webtyp/svg/sprite` already owns converting raw SVG files into symbols without regex or stdlib.

### 2. Novice-name test
- `(*Sprite).AddLayers(prefix, svgContent string) error`: self-describing, extracts all named layer variants with ID prefixed by `prefix + "-" + childID`.
- Returns `error` on malformed SVG or missing `viewBox`.

### 3. Complexity ledger
```
Concepts the developer must learn   +1 (AddLayers) / 0
Files they must touch to do X       0 (new method on Sprite)
Lines at the call site              1 line: s.AddLayers("cote", svgFile)
Ways to do the same thing           +1 (modular layers) / 0
```

### 4. Where does it belong
- In `webtyp.com/svg/sprite` (`sprite/layers.go`), next to `sprite/file.go`.

### 5. What does this change delete?
- Deletes the need for downstream consumers to manually slice, regex, or parse Inkscape XML in application code.

---

## Stage 1 — `sprite/layers.go`

Add `sprite/layers.go`:
```go
package sprite

import "webtyp.com/fmt"

// AddLayers parses an SVG document and adds each top-level layer or variant
// element with an "id" attribute as an individual <symbol> prefixed by prefix + "-" + id.
//
// Specifically for Inkscape SVGs:
// 1. Reads the document viewBox from root <svg viewBox="...">.
// 2. Traverses child elements (<g>, <path>, etc.). When an element or group has an id:
//    - Strips inline `display:none` style.
//    - Preserves `transform="..."` attributes.
//    - Replaces fill="#ffffff" on paths ending in "-bg" with fill="var(--color-surface, #ffffff)".
//    - Registers each named piece as Definition with Icon: svg.Icon(prefix + "-" + id).
// Returns error if no viewBox or no named layers are found.
func (s *Sprite) AddLayers(prefix, svgFile string) error
```

Implementation guidelines:
- Byte scanning only (follow `file.go:index`, `skipSpace`, no stdlib `regexp` or `encoding/xml` in `sprite/`).
- Handles nested `<g id="brows"><path id="brows-normal" .../><path id="brows-sad" .../></g>`: for each leaf or sub-group with an `id`, creates a symbol with the child's inner content wrapped in the parent's `transform` if any.

## Stage 2 — `sprite/layers_test.go`

Add comprehensive unit tests in `sprite/layers_test.go`:
1. `TestAddLayers_Inkscape`: parses a mock Inkscape snippet with `<g id="torso"><path id="torso-normal" .../></g>` and `<g id="eyes"><path id="eyes-normal" style="display:none" .../></g>`, verifies both symbols are registered with `prefix-torso-normal` and `prefix-eyes-normal` and without `display:none`.
2. `TestAddLayers_HandBgFill`: verifies paths with `id="pose-talk-a-bg"` and `fill="#ffffff"` have fill converted to `fill="var(--color-surface, #ffffff)"`.
3. `TestAddLayers_MissingViewBox`: returns error.

## Stage 3 — Documentation

Update `README.md` documenting `AddLayers` alongside `AddFile`.

## Acceptance
- `gotest` passes completely.
- `GOOS=js GOARCH=wasm go list -deps ./... | grep webtyp/svg/sprite` is empty.
