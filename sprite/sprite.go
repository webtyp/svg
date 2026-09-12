package sprite

import (
	"webtyp.com/fmt"
	"webtyp.com/json"
	"webtyp.com/model"
	"webtyp.com/svg"
)

// Node is internal typed markup for a symbol's body.
type Node string

// Path renders <path fill="currentColor" d="<d>"/>.
func Path(d string) Node {
	var b fmt.Builder
	b.WriteString("<path fill=\"currentColor\" d=\"")
	b.WriteString(d)
	b.WriteString("\"/>")
	return Node(b.String())
}

// Raw is an escape hatch for markup that Path does not express.
func Raw(s string) Node {
	return Node(s)
}

// Definition holds the data needed to build a <symbol>.
type Definition struct {
	Icon    svg.Icon
	ViewBox string
	Body    string
}

// Define creates a reusable icon definition.
func Define(icon svg.Icon, viewBox string, body ...Node) Definition {
	var b fmt.Builder
	for _, n := range body {
		b.WriteString(string(n))
	}
	return Definition{
		Icon:    icon,
		ViewBox: viewBox,
		Body:    b.String(),
	}
}

// Sprite holds SVG icon definitions for server-side sprite sheet injection.
type Sprite struct {
	icons []Definition
}

// NewSprite constructs a *Sprite from typed icons.
func NewSprite(icons ...Definition) *Sprite {
	return &Sprite{icons: icons}
}

// AddRaw adds a raw symbol to the sprite. Used for assetmin's raw .svg file/favicon path.
func (s *Sprite) AddRaw(id, content, viewBox string) {
	s.icons = append(s.icons, Definition{
		Icon:    svg.Icon(id),
		Body:    content,
		ViewBox: viewBox,
	})
}

// String renders the sprite as inline SVG with all <symbol> elements. A sprite
// with zero icons renders EmptyWrapper, not an empty string: the document that
// injects it needs a stable insertion point. A nil sprite renders "" — there is
// no sprite to speak of.
func (s *Sprite) String() string {
	if s == nil {
		return ""
	}
	if len(s.icons) == 0 {
		return EmptyWrapper
	}
	var b fmt.Builder
	b.WriteString("<svg aria-hidden=\"true\" style=\"display:none\">")
	for _, e := range s.icons {
		b.WriteString("<symbol id=\"")
		b.WriteString(e.Icon.ID())
		b.WriteString("\" viewBox=\"")
		b.WriteString(e.ViewBox)
		b.WriteString("\">")
		b.WriteString(e.Body)
		b.WriteString("</symbol>")
	}
	b.WriteString("</svg>")
	return b.String()
}

// Encodable/Decodable implementation using webtyp.com/json

// IsNil reports whether the sprite is nil.
func (s *Sprite) IsNil() bool {
	return s == nil
}

// EncodeFields encodes the sprite fields.
func (s *Sprite) EncodeFields(w model.FieldWriter) {}

// DecodeFields decodes the sprite fields.
func (s *Sprite) DecodeFields(r model.FieldReader) {}

// MarshalJSON serializes the Sprite to JSON using webtyp/json.
func (s *Sprite) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	var out []byte
	err := json.Encode(s, &out)
	return out, err
}

// UnmarshalJSON deserializes the Sprite from JSON using webtyp/json.
func (s *Sprite) UnmarshalJSON(b []byte) error {
	return json.Decode(b, s)
}

// FielderSlice implementation for root array

// Icons returns a copy of the sprite's typed icon definitions. Callers that
// need to inspect or recombine icons (e.g. assetmin merging per-module
// sprites) read this instead of round-tripping through String().
func (s *Sprite) Icons() []Definition {
	if s == nil {
		return nil
	}
	out := make([]Definition, len(s.icons))
	copy(out, s.icons)
	return out
}

// Len returns the number of icons in the sprite.
func (s *Sprite) Len() int { return len(s.icons) }

// At returns the Fielder at index i.
func (s *Sprite) At(i int) model.Fielder {
	return &iconFielder{icon: &s.icons[i]}
}

// Append appends a new Definition to the sprite and returns its Fielder.
func (s *Sprite) Append() model.Fielder {
	s.icons = append(s.icons, Definition{})
	return &iconFielder{icon: &s.icons[len(s.icons)-1]}
}

type iconFielder struct {
	icon *Definition
}

func (f *iconFielder) Schema() []model.Field { return nil }
func (f *iconFielder) Pointers() []any       { return nil }
func (f *iconFielder) IsNil() bool           { return f.icon == nil }

func (f *iconFielder) EncodeFields(w model.FieldWriter) {
	w.String("id", f.icon.Icon.ID())
	w.String("content", f.icon.Body)
	w.String("viewBox", f.icon.ViewBox)
}

func (f *iconFielder) DecodeFields(r model.FieldReader) {
	f.icon.Icon = svg.Icon(id(r.String("id")))
	f.icon.Body, _ = r.String("content")
	f.icon.ViewBox, _ = r.String("viewBox")
}

func id(s string, _ bool) string { return s }
