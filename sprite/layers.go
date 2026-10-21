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
func (s *Sprite) AddLayers(prefix, svgFile string) error {
	viewBox, ok := viewBoxOf(svgFile)
	if !ok {
		return fmt.Err("svg file has no viewBox attribute:", prefix)
	}

	body, ok := innerOf(svgFile)
	if !ok {
		return fmt.Err("svg file has no root <svg> element:", prefix)
	}

	foundAny := false

	// Basic recursive byte scanner to find elements with id="..."
	var parseElement func(content string, parentTransform string)
	parseElement = func(content string, parentTransform string) {
		i := 0
		for i < len(content) {
			openTag := index(content, i, "<")
			if openTag < 0 {
				break
			}

			// If it's a closing tag or comment, skip
			if openTag+1 < len(content) && (content[openTag+1] == '/' || content[openTag+1] == '!' || content[openTag+1] == '?') {
				i = index(content, openTag, ">")
				if i < 0 {
					break
				}
				i++
				continue
			}

			// Find end of the tag
			closeBracket := index(content, openTag, ">")
			if closeBracket < 0 {
				break
			}

			// Check if self-closing
			isSelfClosing := false
			if content[closeBracket-1] == '/' {
				isSelfClosing = true
			}

			tagContent := content[openTag:closeBracket+1]

			// Inkscape puts a newline, not a space, after the tag name.
			nameEnd := 1
			for nameEnd < len(tagContent) {
				c := tagContent[nameEnd]
				if c == ' ' || c == '\n' || c == '\t' || c == '\r' || c == '>' || c == '/' {
					break
				}
				nameEnd++
			}

			if nameEnd > 1 {
				tagName := tagContent[1:nameEnd]

				// check for ID
				idVal := getAttr(tagContent, "id")

				// extract inner content if not self closing
				var innerContent string
				var endTagIdx int

				if !isSelfClosing {
					// find closing tag
					depth := 1
					searchIdx := closeBracket + 1
					for searchIdx < len(content) {
						nextOpen := index(content, searchIdx, "<"+tagName)
						nextClose := index(content, searchIdx, "</"+tagName+">")

						if nextClose < 0 {
							break // Invalid XML
						}

						// Determine if the nextOpen is an actual open tag, self-closing tag, etc.
						if nextOpen >= 0 && nextOpen < nextClose {
							// verify it's really an open tag (space or > or /> after)
							if nextOpen+len(tagName)+1 < len(content) {
								nextChar := content[nextOpen+len(tagName)+1]
								if nextChar == ' ' || nextChar == '\n' || nextChar == '\t' || nextChar == '\r' || nextChar == '>' || nextChar == '/' {
									depth++
									searchIdx = nextOpen + 1
									continue
								}
							}
							searchIdx = nextOpen + 1
							continue
						}

						depth--
						if depth == 0 {
							endTagIdx = nextClose + len("</" + tagName + ">")
							innerContent = content[closeBracket+1 : nextClose]
							break
						}
						searchIdx = nextClose + 1
					}
				} else {
					endTagIdx = closeBracket + 1
				}

				transform := getAttr(tagContent, "transform")

				// Determine transform to pass down
				nextTransform := parentTransform
				if transform != "" {
					if nextTransform != "" {
						nextTransform = nextTransform + " " + transform
					} else {
						nextTransform = transform
					}
				}

				if idVal != "" {
					foundAny = true

					// Reconstruct the element with modifications
					cleanContent := removeAttr(tagContent, "style", "display:none")
					cleanContent = removeAttr(cleanContent, "style", "display: none")

					if len(idVal) >= 3 && idVal[len(idVal)-3:] == "-bg" {
						cleanContent = replaceAttr(cleanContent, "fill", "#ffffff", "var(--color-surface, #ffffff)")
					}

					var symbolBody string
					if parentTransform != "" {
						symbolBody = "<g transform=\"" + parentTransform + "\">"
					}

					symbolBody += cleanContent
					if !isSelfClosing {
						symbolBody += showAll(innerContent) + "</" + tagName + ">"
					}

					if parentTransform != "" {
						symbolBody += "</g>"
					}

					s.AddRaw(prefix+"-"+idVal, symbolBody, viewBox)
				}

				// Recurse into inner content if we want to find children with IDs
				if !isSelfClosing && innerContent != "" {
					parseElement(innerContent, nextTransform)
				}

				if endTagIdx > 0 {
					i = openTag + endTagIdx - openTag // Advance to after this element
				} else {
					i = closeBracket + 1
				}

			} else {
				i = closeBracket + 1
			}
		}
	}

	parseElement(body, "")

	if !foundAny {
		return fmt.Err("svg file has no named layers:", prefix)
	}

	return nil
}

// showAll removes every Inkscape editing-time "display:none" inside a variant: once the
// variant is its own symbol, a hidden descendant would never be drawn.
func showAll(content string) string {
	out := content
	for _, hidden := range [...]string{"display:none;", "display: none;", "display:none", "display: none"} {
		for {
			idx := index(out, 0, hidden)
			if idx < 0 {
				break
			}
			out = out[:idx] + out[idx+len(hidden):]
		}
	}
	return out
}

// helper functions for byte scanning attributes
func getAttr(tag string, attr string) string {
	search := attr + "=\""
	start := index(tag, 0, search)
	if start < 0 {
		search = attr + "='"
		start = index(tag, 0, search)
		if start < 0 {
			return ""
		}
	}
	start += len(search)
	quote := tag[start-1]

	end := start
	for end < len(tag) && tag[end] != quote {
		end++
	}

	if end < len(tag) {
		return tag[start:end]
	}
	return ""
}

func removeAttr(tag string, attr string, valueToMatch string) string {
	// this is quite simplistic but suffices for the exact constraints
	search := attr + "=\"" + valueToMatch + "\""
	idx := index(tag, 0, search)
	if idx >= 0 {
		// return tag without this attribute
		// also remove preceding space if exists
		start := idx
		if start > 0 && tag[start-1] == ' ' {
			start--
		}
		return tag[:start] + tag[idx+len(search):]
	}

	search = attr + "='" + valueToMatch + "'"
	idx = index(tag, 0, search)
	if idx >= 0 {
		start := idx
		if start > 0 && tag[start-1] == ' ' {
			start--
		}
		return tag[:start] + tag[idx+len(search):]
	}

	return tag
}

func replaceAttr(tag string, attr string, oldVal string, newVal string) string {
	search := attr + "=\"" + oldVal + "\""
	idx := index(tag, 0, search)
	if idx >= 0 {
		return tag[:idx] + attr + "=\"" + newVal + "\"" + tag[idx+len(search):]
	}

	search = attr + "='" + oldVal + "'"
	idx = index(tag, 0, search)
	if idx >= 0 {
		return tag[:idx] + attr + "='" + newVal + "'" + tag[idx+len(search):]
	}
	return tag
}
