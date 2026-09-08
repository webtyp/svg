//go:build wasm

package svg_test

import (
	"testing"

	"webtyp.com/dom"
	"webtyp.com/dom/domtest"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/svg"
)

type SearchOption struct {
	ID          string
	Label       string
	Description string
}

type SelectSearch struct {
	dom.Element

	Placeholder string
	Options     []SearchOption
	OnSelect    func(id, description string)
	OnSearch    func(term string) []SearchOption

	selectedLabel *dom.SignalString
	filterTerm    *dom.SignalString
	isOpen        *dom.SignalBool
	optionNodes   *dom.SignalNodes
}

const iconArrowDown = svg.Icon("ss-arrow-down")

func (c *SelectSearch) Init(ctx dom.Ctx) {
	c.selectedLabel = dom.NewString("")
	c.filterTerm = dom.NewString("")
	c.isOpen = dom.NewBool(false)
	c.optionNodes = dom.NewNodes(c.buildOptionNodes()...)
}

func (c *SelectSearch) buildOptionNodes() []*dom.Element {
	term := fmt.Convert(c.filterTerm.Get()).ToLower().String()
	var nodes []*dom.Element
	for _, opt := range c.Options {
		if !c.matches(opt, term) {
			continue
		}
		opt := opt
		item := Div().Class("ss-option").
			ID("ss-opt-"+opt.ID).
			Attr("data-id", opt.ID).
			On("click", func(e dom.Event) { c.selectOption(opt) }).
			Child(Span().Class("ss-label").Text(opt.Label))
		if opt.Description != "" {
			item.Child(Span().Class("ss-desc").Text(opt.Description))
		}
		nodes = append(nodes, item)
	}
	return nodes
}

func (c *SelectSearch) Render() *dom.Element {
	headerText := dom.DeriveString(func() string {
		if sel := c.selectedLabel.Get(); sel != "" {
			return sel
		}
		if c.Placeholder != "" {
			return c.Placeholder
		}
		return "Select..."
	})

	toggle := Input("checkbox").Class("ss-toggle").ID("ss-toggle-id").
		BindAttrBool("checked", c.isOpen)

	header := Label().For(toggle).Class("ss-header").Child(
		Span().BindText(headerText),
		iconArrowDown.Render("ss-icon"),
	)

	search := Input("search").
		ID("ss-search-id").
		Class("ss-search").
		Attr("placeholder", "Search...").
		BindAttr("value", c.filterTerm).
		On("input", c.onSearchInput)

	list := Div().Class("ss-options").ID("ss-options").BindChildren(c.optionNodes)

	return Div().Class("ss-box").Child(
		toggle,
		header,
		Div().Class("ss-dropdown").Child(search, list),
	)
}

func (c *SelectSearch) onSearchInput(e dom.Event) {
	c.filterTerm.Set(e.TargetValue())
	if len(c.filteredOptions()) == 0 && c.OnSearch != nil {
		c.Options = c.OnSearch(c.filterTerm.Get())
	}
	c.optionNodes.Set(c.buildOptionNodes())
}

func (c *SelectSearch) selectOption(opt SearchOption) {
	c.selectedLabel.Set(opt.Label)
	c.isOpen.Set(false)
	if c.OnSelect != nil {
		c.OnSelect(opt.ID, opt.Description)
	}
}

func (c *SelectSearch) matches(opt SearchOption, term string) bool {
	if term == "" {
		return true
	}
	return fmt.Contains(fmt.Convert(opt.Label).ToLower().String(), term) ||
		fmt.Contains(fmt.Convert(opt.Description).ToLower().String(), term)
}

func (c *SelectSearch) filteredOptions() []SearchOption {
	term := fmt.Convert(c.filterTerm.Get()).ToLower().String()
	out := make([]SearchOption, 0, len(c.Options))
	for _, o := range c.Options {
		if c.matches(o, term) {
			out = append(out, o)
		}
	}
	return out
}

func TestSelectSearch(t *testing.T) {
	domtest.Mount(t, "root")

	selectedID := ""
	c := &SelectSearch{
		Placeholder: "Choose one",
		Options: []SearchOption{
			{ID: "opt1", Label: "Option 1", Description: "First option"},
			{ID: "opt2", Label: "Option 2", Description: "Second option"},
		},
		OnSelect: func(id, desc string) {
			selectedID = id
		},
	}

	if err := dom.Render("root", c); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	_, ok := domtest.Query("#ss-toggle-id")
	if !ok {
		t.Fatal("toggle not found")
	}

	domtest.Fill("#ss-search-id", "Option 2")
	if c.filterTerm.Get() != "Option 2" {
		t.Errorf("expected filterTerm='Option 2', got %q", c.filterTerm.Get())
	}

	_, ok = domtest.Query("#ss-opt-opt1")
	if ok {
		t.Error("opt1 should be filtered out")
	}

	domtest.Fire("#ss-opt-opt2", "click")
	if selectedID != "opt2" {
		t.Errorf("expected selectedID='opt2', got %q", selectedID)
	}
	if c.selectedLabel.Get() != "Option 2" {
		t.Errorf("expected selectedLabel='Option 2', got %q", c.selectedLabel.Get())
	}
}
