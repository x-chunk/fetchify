package ui

// row is one line of the card before layout: an icon, a label and a
// value made of parts. Rows are data so that the screen can lay them out
// again with less detail when space is short.
type row struct {
	// icon is the glyph before the label, or 0 for none.
	icon rune
	// label names the row.
	label string
	// value is the text of the row in order. It is empty for a pure
	// meter row.
	value []segment
	// meter, when set, is drawn as a bar with numbers before the value.
	meter *meter
	// badge, when set, follows the value.
	badge *badge
}

// role is the style a part of a row is drawn in.
type role uint8

// The roles of row parts. They map to styles of the theme.
const (
	roleValue role = iota
	roleMuted
	roleGood
	roleWarning
	roleDanger
)

// segment is a piece of a value.
type segment struct {
	text string
	role role
	// detail marks secondary information that the screen drops first
	// when the card does not fit.
	detail bool
}

// meter is a share of a resource, drawn as a bar.
type meter struct {
	// percent is the used share, rounded and clamped to 0-100.
	percent int
	// used and total are in bytes. They are shown only when sizes is set.
	used, total uint64
	sizes       bool
	// inverted makes a low value the danger, as for a battery.
	inverted bool
}

// badgeLook is the style of a badge.
type badgeLook uint8

// Badge looks.
const (
	badgeNeutral badgeLook = iota
	badgeGood
	badgeWarning
)

// badge is a short state label such as the session type.
type badge struct {
	text string
	look badgeLook
}

// textValue returns a value of one segment in the value style.
func textValue(text string) []segment {
	return []segment{{text: text}}
}

// unknownRow is the row of a fact that could not be determined.
func unknownRow(label string) row {
	return row{label: label, value: []segment{{text: "unknown", role: roleMuted}}}
}
