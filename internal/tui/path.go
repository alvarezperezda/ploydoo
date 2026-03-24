package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func pathView(input string, pathErr string) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Installation path:") + "\n\n"
	s += fmt.Sprintf("  %s%s\n", infoStyle.Render("> "), input)

	if pathErr != "" {
		s += "\n" + errorStyle.Render("  "+pathErr)
	}

	s += "\n" + helpStyle.Render("Type the path where Odoo will be installed • enter confirm • q quit")
	return boxStyle.Render(s)
}
