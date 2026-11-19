//go:build wasm

package svg_test

import (
	"strings"
	"syscall/js"
	"testing"

	"webtyp.com/svg"
)

func TestIconRender_ReadsWebtypSpriteMetaInWASM(t *testing.T) {
	svg.SetSpriteURL("")
	t.Cleanup(func() { svg.SetSpriteURL("") })

	doc := js.Global().Get("document")
	meta := doc.Call("createElement", "meta")
	meta.Call("setAttribute", "name", svg.SpriteMetaName)
	meta.Call("setAttribute", "content", "/sprite.deadbeef.svg")
	doc.Get("head").Call("appendChild", meta)
	t.Cleanup(func() { meta.Call("remove") })

	got := svg.Icon("home").Render().String()
	if !strings.Contains(got, "href='/sprite.deadbeef.svg#home'") {
		t.Fatalf("expected external sprite href from <meta>, got: %s", got)
	}
}
