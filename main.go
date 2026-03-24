package main

import (
	"fmt"
	"os"

	"odoo-cli/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("ploydoo %s\n", version)
		return
	}

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running CLI: %v\n", err)
		os.Exit(1)
	}

	m := finalModel.(tui.Model)
	if m.Err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", m.Err)
		os.Exit(1)
	}
}
