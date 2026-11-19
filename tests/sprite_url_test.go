package svg_test

import (
	"strings"
	"testing"

	"webtyp.com/svg"
	"webtyp.com/svg/sprite"
)

func TestIconRender_DefaultHashOnly(t *testing.T) {
	svg.SetSpriteURL("")
	t.Cleanup(func() { svg.SetSpriteURL("") })

	got := svg.Icon("check").Render().String()
	if !strings.Contains(got, "href='#check'") {
		t.Fatalf("expected bare #check href, got: %s", got)
	}
}

func TestIconRender_ExternalSpriteURL(t *testing.T) {
	svg.SetSpriteURL("/sprite.abc123.svg")
	t.Cleanup(func() { svg.SetSpriteURL("") })

	got := svg.Icon("check").Render().String()
	if !strings.Contains(got, "href='/sprite.abc123.svg#check'") {
		t.Fatalf("expected external sprite href, got: %s", got)
	}
}

func TestSprite_ExternalXMLNSAndNoInlineStyle(t *testing.T) {
	s := sprite.NewSprite(sprite.Define(svg.Icon("check"), "0 0 16 16", sprite.Path("M1 2")))
	got := s.String()
	if !strings.HasPrefix(got, `<svg xmlns="http://www.w3.org/2000/svg">`) {
		t.Fatalf("expected external SVG xmlns header, got: %s", got)
	}
	if strings.Contains(got, "style=") {
		t.Fatalf("sprite must not contain inline style attribute, got: %s", got)
	}
	if !strings.HasPrefix(sprite.EmptyWrapper, `<svg xmlns="http://www.w3.org/2000/svg">`) {
		t.Fatalf("EmptyWrapper must carry xmlns without inline style, got: %s", sprite.EmptyWrapper)
	}
}
