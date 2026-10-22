package ui

import (
	"fmt"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stonear/effective-interest-rate/internal/eir"
)

func money(v float64) string {
	s := fmt.Sprintf("%.2f", math.Abs(v))
	p := strings.Split(s, ".")
	for i := len(p[0]) - 3; i > 0; i -= 3 {
		p[0] = p[0][:i] + "," + p[0][i:]
	}
	if v < 0 {
		return "-" + p[0] + "." + p[1]
	}
	return p[0] + "." + p[1]
}

var headers = []string{"Date", "Principal", "Principal repayment", "Year fraction", "Contractual rate", "Contractual interest", "Effective interest expense", "Amortisation", "Amortised cost"}

func rowCells(r eir.Row) []string {
	return []string{r.Date.Format("2006-01-02"), money(r.Principal), money(r.PrincipalRepayment), fmt.Sprintf("%.9f", r.YearFraction), fmt.Sprintf("%.4f%%", r.ContractualRate*100), money(r.ContractualInterest), money(r.EffectiveInterestExpense), money(r.Amortisation), money(r.AmortisedCost)}
}
func (m Model) View() tea.View {
	var content string
	if m.result == nil {
		content = m.formView()
	} else {
		content = m.resultView()
	}
	return tea.NewView(clipLines(content, m.width))
}

func (m Model) formView() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	var b strings.Builder
	b.WriteString(title.Render("Effective interest rate calculator"))
	b.WriteString("\n\n")
	for index, input := range m.inputs {
		fmt.Fprintf(&b, "%s %-34s %s\n", focusMark(m.focus == index), labels[index], input.View())
	}
	fmt.Fprintf(&b, "%s Day-count convention: %s  (←/→)\n", focusMark(m.focus == len(m.inputs)), eir.Basis(m.basis))
	fmt.Fprintf(&b, "%s [ Generate ]\n", focusMark(m.focus == len(m.inputs)+1))
	if m.err != "" {
		fmt.Fprintf(&b, "\nError: %s\n", m.err)
	}
	b.WriteString("\nTab/Shift+Tab: move  •  Ctrl+G: generate  •  Ctrl+C: quit")
	return b.String()
}

func focusMark(selected bool) string {
	if selected {
		return ">"
	}
	return " "
}

func (m Model) resultView() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	var b strings.Builder
	m.writeSummary(&b, title)
	if m.status != "" {
		fmt.Fprintf(&b, "%s\n\n", m.status)
	}
	m.writeTable(&b)
	return b.String()
}

func (m Model) writeSummary(b *strings.Builder, title lipgloss.Style) {
	result := m.result
	fmt.Fprintf(b, "%s\nSettlement %s   Maturity %s   Basis: %s\n",
		title.Render("Effective interest rate and amortised cost"),
		result.Input.Settlement.Format("2006-01-02"),
		result.Input.Maturity.Format("2006-01-02"), result.Input.Basis)
	fmt.Fprintf(b, "EIR (annualised daily schedule rate): %.9f%%\n", result.Rate*100)
	fmt.Fprintf(b, "Initial amortised cost: %s   Closing: %s\n",
		money(result.InitialAmortisedCost), money(result.ClosingAmortisedCost))
	fmt.Fprintf(b, "Contractual interest: %s   Effective interest expense: %s   Amortisation: %s\n",
		money(result.TotalContractualInterest), money(result.TotalEffectiveInterestExpense), money(result.TotalAmortisation))
	if result.Input.CustomerName != "" {
		fmt.Fprintf(b, "Customer: %s\n", strings.ReplaceAll(result.Input.CustomerName, "\n", " "))
	}
	if result.Input.DealReference != "" {
		fmt.Fprintf(b, "Deal: %s\n", strings.ReplaceAll(result.Input.DealReference, "\n", " "))
	}
	b.WriteString("\n")
}

func (m Model) visibleColumns(widths []int) []int {
	first := max(m.col, 1)
	selected := []int{0}
	used := widths[0] + 3
	for column := first; column < len(headers); column++ {
		if m.width > 0 && used+widths[column]+3 > m.width {
			break
		}
		selected = append(selected, column)
		used += widths[column] + 3
	}
	return selected
}

func (m Model) writeTable(b *strings.Builder) {
	visibleRows := max(m.height-10, 2)
	start := max(m.row-visibleRows+1, 0)
	end := min(start+visibleRows, len(m.result.Rows))
	widths := []int{12, 21, 21, 17, 17, 22, 27, 21, 21}
	columns := m.visibleColumns(widths)

	for _, column := range columns {
		fmt.Fprintf(b, "%-*s │ ", widths[column], headers[column])
	}
	b.WriteString("\n")
	for index := start; index < end; index++ {
		cells := rowCells(m.result.Rows[index])
		b.WriteString(focusMark(index == m.row))
		for _, column := range columns {
			fmt.Fprintf(b, "%-*s │ ", widths[column]-1, cells[column])
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(b, "\nRow %d/%d  Column %d/9  ↑↓/PgUp/PgDn/Home/End rows  ←→ columns  c: save CSV  e: edit  Ctrl+C: quit",
		m.row+1, len(m.result.Rows), m.col+1)
}

func clipLines(content string, width int) string {
	if width <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		if len(line) > width {
			lines[index] = line[:width]
		}
	}
	return strings.Join(lines, "\n")
}
