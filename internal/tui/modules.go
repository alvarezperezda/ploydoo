package tui

import "fmt"

var ocaModules = []string{
	"account-analytic",
	"account-closing",
	"account-financial-reporting",
	"account-financial-tools",
	"account-invoice-reporting",
	"account-invoicing",
	"account-payment",
	"account-reconcile",
	"bank-payment",
	"bank-statement-import",
	"commission",
	"community-data-files",
	"connector",
	"credit-control",
	"crm",
	"event",
	"hr",
	"hr-attendance",
	"hr-holidays",
	"l10n-spain",
	"manufacture",
	"mis-builder",
	"multi-company",
	"partner-contact",
	"pos",
	"product-attribute",
	"product-variant",
	"project",
	"purchase-workflow",
	"queue",
	"reporting-engine",
	"sale-reporting",
	"sale-workflow",
	"server-tools",
	"server-ux",
	"social",
	"stock-logistics-barcode",
	"stock-logistics-reporting",
	"stock-logistics-warehouse",
	"stock-logistics-workflow",
	"timesheet",
	"web",
}

// modulesVisibleCount returns how many module rows can fit given a terminal height.
func modulesVisibleCount(termHeight int) int {
	// overhead: logo 8 (7 art + 1 signature) + blanks 2 + subtitle 1 + blank 1 + help 2 + border/padding 4 + scroll hints 2 = 20
	const overhead = 20
	visible := termHeight - overhead
	if visible < 5 {
		visible = 5
	}
	if visible > len(ocaModules) {
		visible = len(ocaModules)
	}
	return visible
}

func modulesView(cursor int, selected map[int]bool, offset, termHeight int) string {
	s := renderLogo() + "\n\n"
	s += subtitleStyle.Render("Select OCA modules to install:") + "\n\n"

	visible := modulesVisibleCount(termHeight)

	// Clamp offset
	if offset < 0 {
		offset = 0
	}
	if offset > len(ocaModules)-visible {
		offset = len(ocaModules) - visible
	}

	// Scroll-up indicator
	if offset > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↑ %d more above", offset)) + "\n"
	}

	end := offset + visible
	if end > len(ocaModules) {
		end = len(ocaModules)
	}

	for i := offset; i < end; i++ {
		mod := ocaModules[i]
		cur := "  "
		if i == cursor {
			cur = cursorStyle.Render("> ")
		}

		check := uncheckStyle.Render("[ ]")
		style := unselectedStyle
		if selected[i] {
			check = checkStyle.Render("[✓]")
			style = selectedStyle
		}
		if i == cursor {
			style = selectedStyle
		}

		s += fmt.Sprintf("%s%s %s\n", cur, check, style.Render(mod))
	}

	// Scroll-down indicator
	remaining := len(ocaModules) - end
	if remaining > 0 {
		s += dimStyle.Render(fmt.Sprintf("  ↓ %d more below", remaining)) + "\n"
	}

	s += "\n" + helpStyle.Render("↑/↓ navigate • space toggle • a select all • n deselect all • enter confirm • q quit")
	return boxStyle.Render(s)
}
