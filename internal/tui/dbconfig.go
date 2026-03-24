package tui

import "fmt"

const (
	dbFieldUser = iota
	dbFieldPassword
	dbFieldName
	dbFieldPort
	dbFieldCount
)

var dbFieldLabels = [dbFieldCount]string{
	"DB User",
	"DB Password",
	"DB Name",
	"DB Port",
}

func dbConfigView(fields [dbFieldCount]string, activeField int, dbErr string) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Database configuration:") + "\n\n"

	for i := 0; i < dbFieldCount; i++ {
		label := dbFieldLabels[i]
		value := fields[i]

		cur := "  "
		style := unselectedStyle
		if i == activeField {
			cur = cursorStyle.Render("> ")
			style = selectedStyle
		}

		display := value
		if display == "" {
			display = dimStyle.Render("(empty)")
		} else if i == dbFieldPassword {
			masked := ""
			for range display {
				masked += "*"
			}
			display = style.Render(masked)
		} else {
			display = style.Render(display)
		}

		s += fmt.Sprintf("%s%s: %s\n", cur, infoStyle.Render(label), display)
	}

	if dbErr != "" {
		s += "\n" + errorStyle.Render("  "+dbErr)
	}

	s += "\n" + helpStyle.Render("↑/↓ switch field • type to edit • enter confirm • q quit")
	return boxStyle.Render(s)
}
