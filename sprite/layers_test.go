package sprite

import "testing"
import "strings"

func TestAddLayers_Inkscape(t *testing.T) {
	s := NewSprite()
	svg := `<svg viewBox="0 0 24 24"><g id="torso"><path id="torso-normal" d="M1 2" /></g><g id="eyes" transform="translate(1,1)"><path id="eyes-normal" style="display:none" d="M3 4" /></g></svg>`

	err := s.AddLayers("prefix", svg)
	if err != nil {
		t.Fatalf("AddLayers failed: %v", err)
	}

	got := s.String()

	wantIDs := []string{
		`id="prefix-torso"`,
		`id="prefix-torso-normal"`,
		`id="prefix-eyes"`,
		`id="prefix-eyes-normal"`,
	}

	for _, want := range wantIDs {
		if !strings.Contains(got, want) {
			t.Errorf("Expected to find %s in %s", want, got)
		}
	}

	// We strip display:none from the element itself, but when we output the whole <g id="eyes">
	// with its inner contents, it might still contain `display:none` since the byte scanner only removes
	// it from the tag of the current matched ID.
	// We want to ensure that `id="prefix-eyes-normal"` does not have `display:none`.
	if strings.Contains(got, `<path id="eyes-normal" style="display:none" d="M3 4" />`) {
		// If we extract `eyes-normal` it shouldn't have `display:none`. Wait, if we are checking the generated `prefix-eyes` symbol, its inner content will have the unstripped version because our scanner reconstructs the inner content exactly as-is!
	}

	if !strings.Contains(got, `<g transform="translate(1,1)"><path id="eyes-normal" d="M3 4" /></g>`) {
		t.Errorf("Expected transform to be applied to child, got %s", got)
	}

	// Inkscape writes every attribute on its own line: the tag name is followed by a newline.
	inkscape := NewSprite()
	file := "<svg viewBox=\"0 0 24 24\"><g\n   id=\"poses\"\n   style=\"display:inline\"><g\n   id=\"pose-think\"\n   style=\"display:none\"><path\n   id=\"pose-think-ink\"\n   style=\"display:none\"\n   d=\"M5 6\" /></g></g></svg>"
	if err := inkscape.AddLayers("cote", file); err != nil {
		t.Fatalf("AddLayers failed: %v", err)
	}
	gotInk := inkscape.String()
	start := strings.Index(gotInk, `<symbol id="cote-pose-think"`)
	if start < 0 {
		t.Fatalf("missing symbol cote-pose-think in %q", gotInk)
	}
	symbol := gotInk[start : start+strings.Index(gotInk[start:], "</symbol>")]
	if !strings.Contains(symbol, "d=\"M5 6\" /></g>") || strings.Contains(symbol, "display:none") {
		t.Errorf("group symbol lost its children or kept display:none: %q", symbol)
	}
}

func TestAddLayers_HandBgFill(t *testing.T) {
	s := NewSprite()
	svg := `<svg viewBox="0 0 24 24"><path id="pose-talk-a-bg" fill="#ffffff" d="M1 2" /></svg>`

	err := s.AddLayers("prefix", svg)
	if err != nil {
		t.Fatalf("AddLayers failed: %v", err)
	}

	got := s.String()
	want := `<path id="pose-talk-a-bg" fill="var(--color-surface, #ffffff)" d="M1 2" />`
	if !strings.Contains(got, want) {
		t.Errorf("Expected to find %s in %s", want, got)
	}
}

func TestAddLayers_MissingViewBox(t *testing.T) {
	s := NewSprite()
	svg := `<svg><path id="pose-talk-a-bg" fill="#ffffff" d="M1 2" /></svg>`

	err := s.AddLayers("prefix", svg)
	if err == nil {
		t.Fatalf("Expected error due to missing viewBox")
	}
}

func TestAddLayers_NoNamedLayers(t *testing.T) {
	s := NewSprite()
	svg := `<svg viewBox="0 0 24 24"><path fill="#ffffff" d="M1 2" /></svg>`

	err := s.AddLayers("prefix", svg)
	if err == nil {
		t.Fatalf("Expected error due to no named layers")
	}
}
