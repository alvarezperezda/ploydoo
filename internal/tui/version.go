package tui

import "fmt"

var odooVersions = []string{
	"18.0",
	"17.0",
	"16.0",
}

func versionView(cursor int) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Select an Odoo version:") + "\n\n"

	for i, v := range odooVersions {
		cur := "  "
		style := unselectedStyle
		if i == cursor {
			cur = cursorStyle.Render("> ")
			style = selectedStyle
		}
		s += fmt.Sprintf("%s%s\n", cur, style.Render("Odoo "+v))
	}

	s += "\n" + helpStyle.Render("↑/↓ navigate • enter select • q quit")
	return boxStyle.Render(s)
}
