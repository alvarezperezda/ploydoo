package tui

import "fmt"

func pythonView(pythonVersion string, pythonAvailable, pyenvAvailable bool) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Python environment setup:") + "\n\n"

	s += fmt.Sprintf("  %s Python %s is required for the selected Odoo version.\n\n",
		infoStyle.Render("ℹ"),
		infoStyle.Render(pythonVersion))

	if pythonAvailable {
		s += fmt.Sprintf("  %s Python %s found\n",
			successStyle.Render("✓"),
			successStyle.Render(pythonVersion))
		s += "\n" + helpStyle.Render("Press enter to continue • q quit")
	} else {
		s += fmt.Sprintf("  %s Python %s not found\n\n",
			errorStyle.Render("✗"),
			errorStyle.Render(pythonVersion))

		if pyenvAvailable {
			s += fmt.Sprintf("  Install Python %s with pyenv? (y/n)\n", pythonVersion)
			s += "\n" + helpStyle.Render("y install with pyenv • n skip and continue • q quit")
		} else {
			s += fmt.Sprintf("  %s pyenv is not installed. Cannot auto-install Python.\n",
				errorStyle.Render("✗"))
			s += fmt.Sprintf("  %s Odoo may not work without Python %s.\n",
				dimStyle.Render("⚠"),
				pythonVersion)
			s += "\n" + helpStyle.Render("Press enter to continue anyway • q quit")
		}
	}

	return boxStyle.Render(s)
}
