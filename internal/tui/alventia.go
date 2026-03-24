package tui

import "fmt"

func alventiaView(branches []string, cursor int, loading bool, loadErr string, offset, termHeight int) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Select alventia_modules branch:") + "\n\n"

	if loading {
		s += dimStyle.Render("  Fetching branches...") + "\n"
		return boxStyle.Render(s)
	}

	if loadErr != "" {
		s += errorStyle.Render("  "+loadErr) + "\n"
		s += "\n" + helpStyle.Render("enter skip alventia • q quit")
		return boxStyle.Render(s)
	}

	if len(branches) == 0 {
		s += dimStyle.Render("  No branches found") + "\n"
		s += "\n" + helpStyle.Render("enter continue • q quit")
		return boxStyle.Render(s)
	}

	visible := alventiaVisibleCount(termHeight)

	if offset < 0 {
		offset = 0
	}
	if offset > len(branches)-visible {
		offset = len(branches) - visible
	}
	if offset < 0 {
		offset = 0
	}

	if offset > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↑ %d more above", offset)) + "\n"
	}

	end := offset + visible
	if end > len(branches) {
		end = len(branches)
	}

	for i := offset; i < end; i++ {
		cur := "  "
		style := unselectedStyle
		if i == cursor {
			cur = cursorStyle.Render("> ")
			style = selectedStyle
		}
		s += fmt.Sprintf("%s%s\n", cur, style.Render(branches[i]))
	}

	remaining := len(branches) - end
	if remaining > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↓ %d more below", remaining)) + "\n"
	}

	s += "\n" + helpStyle.Render("↑/↓ navigate • enter select • q quit")
	return boxStyle.Render(s)
}

func alventiaVisibleCount(termHeight int) int {
	const overhead = 20
	visible := termHeight - overhead
	if visible < 5 {
		visible = 5
	}
	return visible
}
