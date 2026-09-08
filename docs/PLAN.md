---
PLAN: "fix(svg/tests): restore the WASM suite on dom/domtest"
TAG: v0.3.9
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 826954424214104375
PR: https://github.com/webtyp/svg/pull/8
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **`webtyp.com/dom v0.13.12` is published**, and it is what introduces the
> `webtyp.com/dom/domtest` package this plan migrates to. Note `svg/tests` is a
> SEPARATE module with its own `go.mod` — bump it there, not in the parent.

# PLAN — `svg/tests` WASM suite compiles again

## The defect

`svg/tests/uc_selectsearch_test.go` (`//go:build wasm`) calls three helpers it
does not define:

```
vet: ./uc_selectsearch_test.go:135:2: undefined: SetupDOM
```

`SetupDOM`, `GetRef` and `TriggerEvent` live in `html/tests/uc_common_test.go` —
`package html_test`, in a **different module**. Test packages are not importable
across modules, so this file has not compiled since the module split.

It stayed invisible because `svg/tests/go.mod` pinned `webtyp.com/dom v0.13.9`, a
tag published under the old `github.com/tinywasm/dom` path, so every `go` command
in the module failed before reaching the compiler:

```
go: webtyp.com/dom@v0.13.9: parsing go.mod:
	module declares its path as: github.com/tinywasm/dom
	        but was required as: webtyp.com/dom
```

`svg v0.3.7` unpinned it. The `!wasm` suite (9 sanitize tests) passes again; the
`wasm` suite is what this plan restores.

## Rules for this plan

- **Do not copy the helpers in from `html/tests`.** That is what created the
  problem. `dom/domtest` is the one place this wiring lives now — per
  `CONSTRUCTION_HARNESS.md`, *"A consumer never re-creates a missing symbol
  locally."*
- **Do not change `svg` itself.** Only `svg/tests/`.
- **Do not delete the test to make the build pass.** It is a consumer-shaped
  proof that `svg` composes with `dom` + `html`; if it cannot be restored, stop
  and report why.
- **No `map`** in anything reaching WASM, and **no standard library** in code
  shared with the frontend — use `webtyp/fmt`.

## Stage 1 — depend on `domtest`

In `svg/tests/go.mod`, bump `webtyp.com/dom` to `v0.13.12` (or newer) and
`go mod tidy`. The `replace webtyp.com/svg => ../` line stays.

## Stage 2 — migrate the three call sites

In `svg/tests/uc_selectsearch_test.go`, import
`webtyp.com/dom/domtest` and replace:

| Current | Replacement |
|---|---|
| `SetupDOM(t)` (line 135) | `domtest.Mount(t, "root")` — the test renders into `"root"` via `dom.Render("root", c)`, so the id must match |
| `GetRef("ss-toggle-id")` | `domtest.Query("#ss-toggle-id")` |
| `TriggerEvent("ss-search-id", "input", "Option 2")` | `domtest.Fill("#ss-search-id", "Option 2")` — `Fill` sets the value and fires `"input"`, which is exactly what `TriggerEvent` did with a non-empty value |
| `TriggerEvent(id, evt, "")` (if any) | `domtest.Fire("#"+id, evt)` |

Grep the whole file for every occurrence — the table lists the ones the audit
found at lines 135 and below, but migrate **all** of them.

If `domtest.Query` returns `(nil, false)` for an element the old `GetRef` found,
that is the documented limitation of `Query` (it resolves through the node's
`id`, and an element `dom` never id'd has none). **Do not work around it with a
local helper** — give the element a `Key(...)` in the mock component so `dom`
assigns it an id (that is what `dom v0.13.11`'s keyed auto-id is for), and if
that is not possible, stop and report it.

## Stage 3 — verify

- `GOOS=js GOARCH=wasm go vet ./...` in `svg/tests` → clean. This is the exact
  command that currently fails; it is the acceptance signal.
- `gotest` in `svg/tests` → green, `wasm ✅` included.
- `gotest` in `svg/` (the parent module) stays green and untouched.

## Acceptance criteria

- `grep -rn 'SetupDOM\|GetRef(\|TriggerEvent(' svg/tests/` → **empty**.
- `grep -n 'webtyp.com/dom v0' svg/tests/go.mod` → `v0.13.12` or newer.
- No new helper function is declared in `svg/tests/` — `git diff` adds no
  `func` that is not a `Test*`.
- `GOOS=js GOARCH=wasm go vet ./...` clean in `svg/tests`.
- `gofmt -l .` → empty.

## Out of scope

- `webtyp/dom` — upstream, `domtest` arrives already published.
- `html/tests` — it still owns its copies of these helpers; migrating it is its
  own plan.
- `svg` proper (`icon.go`, `sprite/`, `sanitize/`, `watch/`).

## Stages

| # | Files | Change |
|---|---|---|
| 1 | `svg/tests/go.mod`, `go.sum` | require `dom v0.13.12` |
| 2 | `svg/tests/uc_selectsearch_test.go` | `SetupDOM`/`GetRef`/`TriggerEvent` → `domtest.Mount`/`Query`/`Fill`/`Fire` |
| 3 | — | `GOOS=js GOARCH=wasm go vet ./...` clean; `gotest` green |
