package tui

import "fmt"

func pythonView(pythonVersion string, pythonAvailable, pyenvAvailable, poetryAvailable bool) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Python environment setup:") + "\n\n"

	s += fmt.Sprintf("  %s Python %s is required for the selected Odoo version.\n\n",
		infoStyle.Render("ℹ"),
		infoStyle.Render(pythonVersion))

	// Python status
	if pythonAvailable {
		s += fmt.Sprintf("  %s Python %s found\n",
			successStyle.Render("✓"),
			successStyle.Render(pythonVersion))
	} else {
		s += fmt.Sprintf("  %s Python %s not found\n",
			errorStyle.Render("✗"),
			errorStyle.Render(pythonVersion))
	}

	// Poetry status
	if poetryAvailable {
		s += fmt.Sprintf("  %s Poetry found\n",
			successStyle.Render("✓"))
	} else {
		s += fmt.Sprintf("  %s Poetry not found\n",
			errorStyle.Render("✗"))
	}

	s += "\n"

	// Determine what needs to be installed
	needsPython := !pythonAvailable && pyenvAvailable
	needsPoetry := !poetryAvailable

	if !needsPython && !needsPoetry {
		s += helpStyle.Render("Press enter to continue • q quit")
	} else {
		var installs []string
		if needsPython {
			installs = append(installs, fmt.Sprintf("Python %s (pyenv)", pythonVersion))
		}
		if !pythonAvailable && !pyenvAvailable {
			s += fmt.Sprintf("  %s pyenv not installed — cannot auto-install Python\n",
				dimStyle.Render("⚠"))
		}
		if needsPoetry {
			installs = append(installs, "Poetry (pipx)")
		}

		if len(installs) > 0 {
			s += fmt.Sprintf("  Install %s? (y/n)\n", joinWithAnd(installs))
		}

		s += "\n" + helpStyle.Render("y install missing tools • n skip and continue • q quit")
	}

	return boxStyle.Render(s)
}

func joinWithAnd(items []string) string {
	if len(items) == 1 {
		return items[0]
	}
	return items[0] + " and " + items[1]
}
