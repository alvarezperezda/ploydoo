package tui

import "fmt"

var pgVersions = []string{
	"17",
	"16",
	"15",
	"14",
}

func postgresView(cursor int) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Select a PostgreSQL version:") + "\n\n"

	for i, v := range pgVersions {
		cur := "  "
		style := unselectedStyle
		if i == cursor {
			cur = cursorStyle.Render("> ")
			style = selectedStyle
		}
		s += fmt.Sprintf("%s%s\n", cur, style.Render("PostgreSQL "+v))
	}

	s += "\n" + helpStyle.Render("↑/↓ navigate • enter select • q quit")
	return boxStyle.Render(s)
}
