package tui

import "fmt"

func customAddonsView() string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Do you have a custom addons repository?") + "\n\n"
	s += "  " + infoStyle.Render("y") + " Yes\n"
	s += "  " + infoStyle.Render("n") + " No\n"
	s += "\n" + helpStyle.Render("y/n choose • ctrl+c quit")
	return boxStyle.Render(s)
}

func customURLView(input string, urlErr string) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Enter custom addons repository URL:") + "\n\n"
	s += fmt.Sprintf("  %s%s\n", infoStyle.Render("> "), input)

	if urlErr != "" {
		s += "\n" + errorStyle.Render("  "+urlErr)
	}

	s += "\n" + helpStyle.Render("Type the git repo URL • enter confirm • ctrl+c quit")
	return boxStyle.Render(s)
}

func customBranchView(branches []string, cursor int, loading bool, branchErr string, offset, termHeight int) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Select branch for custom addons:") + "\n\n"

	if loading {
		s += dimStyle.Render("  Fetching branches...") + "\n"
		return boxStyle.Render(s)
	}

	if branchErr != "" {
		s += errorStyle.Render("  "+branchErr) + "\n"
		s += "\n" + helpStyle.Render("enter skip custom addons • q quit")
		return boxStyle.Render(s)
	}

	if len(branches) == 0 {
		s += dimStyle.Render("  No branches found") + "\n"
		s += "\n" + helpStyle.Render("enter continue • q quit")
		return boxStyle.Render(s)
	}

	visible := customBranchVisibleCount(termHeight)

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

func customBranchVisibleCount(termHeight int) int {
	const overhead = 20
	visible := termHeight - overhead
	if visible < 5 {
		visible = 5
	}
	return visible
}
