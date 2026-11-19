package svg

import "webtyp.com/dom"

// SpriteMetaName is the <meta name="..."> emitted by sitec with the external sprite URL.
const SpriteMetaName = "webtyp-sprite"

var (
	configuredSpriteURL string
	spriteMetaLoaded    bool
)

// SetSpriteURL sets the external sprite file URL (for example "/sprite.a1b2c3d4.svg")
// prefixed before "#<id>" by Icon.Render. Pass "" to emit bare "#<id>" references.
func SetSpriteURL(url string) {
	configuredSpriteURL = url
	spriteMetaLoaded = false
}

func currentSpriteURL() string {
	if configuredSpriteURL != "" || spriteMetaLoaded {
		return configuredSpriteURL
	}
	if metaURL := dom.GetMeta(SpriteMetaName); metaURL != "" {
		spriteMetaLoaded = true
		configuredSpriteURL = metaURL
		return configuredSpriteURL
	}
	return ""
}
