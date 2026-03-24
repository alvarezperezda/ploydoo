package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// CloneStatus represents the result of a clone operation.
type CloneStatus struct {
	Name    string
	Success bool
	Err     error
}

func progressVisibleCount(termHeight int) int {
	// overhead: logo 8 + blanks 2 + subtitle 1 + blank 1 + spinner 2 + summary 6 + help 2 + border/padding 4 + scroll hints 2 = 28
	const overhead = 28
	visible := termHeight - overhead
	if visible < 5 {
		visible = 5
	}
	return visible
}

func progressView(statuses []CloneStatus, currentTask string, done bool, spinnerView string, offset, termHeight int) string {
	s := renderLogo() + "\n\n"

	if !done {
		s += subtitleStyle.Render("Cloning repositories...") + "\n\n"
	} else {
		s += subtitleStyle.Render("Setup complete!") + "\n\n"
	}

	visible := progressVisibleCount(termHeight)
	total := len(statuses)

	if offset < 0 {
		offset = 0
	}
	if total > visible && offset > total-visible {
		offset = total - visible
	}

	if offset > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↑ %d more above", offset)) + "\n"
	}

	end := offset + visible
	if end > total {
		end = total
	}

	for i := offset; i < end; i++ {
		st := statuses[i]
		icon := successStyle.Render("✓")
		if !st.Success {
			icon = errorStyle.Render("✗")
		}
		name := st.Name
		if st.Success {
			name = successStyle.Render(name)
		} else {
			name = errorStyle.Render(name)
		}
		line := fmt.Sprintf("  %s %s", icon, name)
		if st.Err != nil {
			line += errorStyle.Render(fmt.Sprintf(" — %v", st.Err))
		}
		s += line + "\n"
	}

	remaining := total - end
	if remaining > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↓ %d more below", remaining)) + "\n"
	}

	if !done && currentTask != "" {
		s += "\n  " + spinnerView + " " + infoStyle.Render(currentTask)
	}

	if done {
		s += "\n" + configSummary(statuses)
		s += "\n" + helpStyle.Render("Press q or enter to exit")
	}

	return boxStyle.Render(s)
}

func configSummary(statuses []CloneStatus) string {
	var succeeded []string
	var failed []string
	for _, st := range statuses {
		if st.Success {
			succeeded = append(succeeded, st.Name)
		} else {
			failed = append(failed, st.Name)
		}
	}

	summary := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(0, 1).
		Render(
			fmt.Sprintf(
				"%s\n  Cloned: %d  |  Failed: %d\n  Config: odoo.conf generated",
				successStyle.Render("Summary"),
				len(succeeded),
				len(failed),
			),
		)

	if len(failed) > 0 {
		summary += "\n\n" + errorStyle.Render("Failed repos: ") + strings.Join(failed, ", ")
	}

	return summary
}
