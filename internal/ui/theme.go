package ui

import "github.com/dmsRosa6/glyph/core"

// A dark palette in the spirit of Zed's One Dark theme.
var (
	colBg        = core.NewColor(30, 33, 39)    // app background
	colPanel     = core.NewColor(36, 40, 47)    // panel interior
	colBorder    = core.NewColor(76, 83, 98)    // frames
	colText      = core.NewColor(220, 223, 228) // primary text
	colDim       = core.NewColor(138, 146, 160) // secondary text
	colAccent    = core.NewColor(97, 175, 239)  // blue accent
	colAccentDim = core.NewColor(62, 98, 135)   // inactive accent
	colSelBg     = core.NewColor(50, 58, 74)    // selected row
	colCap       = core.NewColor(60, 66, 80)    // key cap
	colCapSel    = core.NewColor(84, 98, 128)   // key cap on selected row
	colCapText   = core.NewColor(236, 239, 244) // key cap label
	colInputBg   = core.NewColor(46, 51, 62)    // search field
	colWarn      = core.NewColor(229, 192, 123) // hints / warnings
)

// Colors exposed for the app wiring in app.go.
var (
	Background = colBg
	Foreground = colText
)
