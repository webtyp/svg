package svg

import (
	"webtyp.com/dom"
)

// Icon is a typed icon reference (the symbol ID).
// Reference (shared, any target): Only the ID string reaches the WASM binary.
type Icon string

// ID returns the symbol's ID.
func (i Icon) ID() string {
	return string(i)
}

// Render returns <svg aria-hidden='true' focusable='false' class=...><use href="...#id"/></svg>.
// It is the ONLY public path to <svg><use>.
func (i Icon) Render(classes ...string) *dom.Element {
	el := dom.NewElement("svg").
		Attr("aria-hidden", "true").
		Attr("focusable", "false").
		Child(dom.NewElement("use").Attr("href", currentSpriteURL()+"#"+string(i)))

	for _, c := range classes {
		el.Class(c)
	}
	return el
}
