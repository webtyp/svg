package sprite

import "testing"

const iconFile = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"/></svg>`

func TestAddFile(t *testing.T) {
	s := NewSprite()
	if err := s.AddFile("home", iconFile); err != nil {
		t.Fatalf("AddFile: %v", err)
	}

	got := s.String()
	want := `<svg xmlns="http://www.w3.org/2000/svg"><symbol id="home" viewBox="0 0 24 24"><path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"/></symbol></svg>`
	if got != want {
		t.Errorf("String()\ngot:  %s\nwant: %s", got, want)
	}
}

func TestAddFileRejectsMissingViewBox(t *testing.T) {
	s := NewSprite()
	err := s.AddFile("home", `<svg xmlns="http://www.w3.org/2000/svg"><path d="M1 2h3"/></svg>`)
	if err == nil {
		t.Fatal("expected an error: a symbol without viewBox is clipped at render time")
	}
	if s.Len() != 0 {
		t.Errorf("rejected file must not be added, got %d icons", s.Len())
	}
}

func TestViewBoxOf(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"double quotes", `<svg viewBox="0 0 24 24">`, "0 0 24 24", true},
		{"single quotes", `<svg viewBox='0 0 16 16'>`, "0 0 16 16", true},
		{"spaces around =", `<svg viewBox = "0 0 32 32">`, "0 0 32 32", true},
		{"negative origin", `<svg viewBox="-2 -2 20 20">`, "-2 -2 20 20", true},
		{"absent", `<svg width="24">`, "", false},
		{"empty value", `<svg viewBox="">`, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := viewBoxOf(tt.in)
			if ok != tt.ok || got != tt.want {
				t.Errorf("viewBoxOf(%s) = %q,%v; want %q,%v", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}
