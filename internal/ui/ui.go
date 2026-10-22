package ui

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stonear/effective-interest-rate/internal/eir"
)

const (
	settlementField = iota
	customerField
	dealField
	principalField
	issuePriceField
	cashbackField
	costsField
	maturityField
	contractualRateField
	fieldCount
)

var labels = []string{
	"Settlement date (YYYY-MM-DD)",
	"Customer name (optional)",
	"Deal reference (optional)",
	"Principal",
	"Issue price (%)",
	"Upfront cashback",
	"Transaction costs",
	"Maturity date (YYYY-MM-DD)",
	"Contractual rate (%)",
}

var sampleValues = []string{
	"2025-01-10", "", "", "10000000000", "100",
	"45205479.45", "0", "2025-03-10", "5.5",
}
var amountPattern = regexp.MustCompile(`^[+-]?(?:\d+|\d{1,3}(?:,\d{3})+)(?:\.\d+)?$`)

func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if !amountPattern.MatchString(s) {
		return 0, fmt.Errorf("enter a valid decimal amount")
	}
	v, e := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	if e != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("amount must be finite")
	}
	return v, nil
}

type Model struct {
	inputs []textinput.Model
	focus  int
	basis  int
	result *eir.Result
	err    string
	status string
	width  int
	height int
	row    int
	col    int
}

func New() tea.Model {
	m := Model{
		inputs: make([]textinput.Model, fieldCount),
		basis:  int(eir.ActualActual),
		width:  100,
		height: 30,
	}
	for i := range m.inputs {
		input := textinput.New()
		input.SetValue(sampleValues[i])
		input.SetWidth(24)
		input.CharLimit = 120
		m.inputs[i] = input
	}
	m.inputs[0].Focus()
	return m
}
func (m Model) Init() tea.Cmd { return nil }
func (m Model) parse() (eir.Input, error) {
	settlement, err := time.Parse("2006-01-02", m.inputs[settlementField].Value())
	if err != nil {
		return eir.Input{}, fmt.Errorf("settlement date: use YYYY-MM-DD")
	}
	maturity, err := time.Parse("2006-01-02", m.inputs[maturityField].Value())
	if err != nil {
		return eir.Input{}, fmt.Errorf("maturity date: use YYYY-MM-DD")
	}
	principal, err := m.amount(principalField)
	if err != nil {
		return eir.Input{}, err
	}
	issuePrice, err := m.amount(issuePriceField)
	if err != nil {
		return eir.Input{}, err
	}
	cashback, err := m.amount(cashbackField)
	if err != nil {
		return eir.Input{}, err
	}
	costs, err := m.amount(costsField)
	if err != nil {
		return eir.Input{}, err
	}
	ratePercent, err := m.amount(contractualRateField)
	if err != nil {
		return eir.Input{}, err
	}

	input := eir.Input{
		Settlement:        settlement,
		Maturity:          maturity,
		CustomerName:      m.inputs[customerField].Value(),
		DealReference:     m.inputs[dealField].Value(),
		Principal:         principal,
		IssuePricePercent: issuePrice,
		Cashback:          cashback,
		TransactionCosts:  costs,
		ContractualRate:   ratePercent / 100,
		Basis:             eir.Basis(m.basis),
	}
	return input, eir.Validate(input)
}

func (m Model) amount(field int) (float64, error) {
	value, err := parseAmount(m.inputs[field].Value())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", labels[field], err)
	}
	return value, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
	case tea.KeyPressMsg:
		k := v.String()
		if k == "ctrl+c" {
			return m, tea.Quit
		}
		if m.result != nil {
			switch k {
			case "c":
				filename := fmt.Sprintf("eir-schedule-%s-to-%s.csv",
					m.result.Input.Settlement.Format("2006-01-02"),
					m.result.Input.Maturity.Format("2006-01-02"))
				if err := eir.ExportCSV(filename, *m.result); err != nil {
					m.status = "CSV export failed: " + err.Error()
				} else {
					m.status = "CSV saved: " + filename
				}
				return m, nil
			case "e", "esc":
				m.result = nil
				m.err = ""
				m.status = ""
				if m.focus >= len(m.inputs) {
					m.focus = 0
				}
				return m, m.inputs[m.focus].Focus()
			case "up", "k":
				if m.row > 0 {
					m.row--
				}
			case "down", "j":
				if m.row < len(m.result.Rows)-1 {
					m.row++
				}
			case "pgup":
				m.row -= 10
				if m.row < 0 {
					m.row = 0
				}
			case "pgdown":
				m.row += 10
				if m.row >= len(m.result.Rows) {
					m.row = len(m.result.Rows) - 1
				}
			case "home":
				m.row = 0
			case "end":
				m.row = len(m.result.Rows) - 1
			case "left", "h":
				if m.col > 0 {
					m.col--
				}
			case "right", "l":
				if m.col < 8 {
					m.col++
				}
			}
			return m, nil
		}
		switch k {
		case "tab", "down":
			if m.focus < len(m.inputs) {
				m.inputs[m.focus].Blur()
			}
			m.focus = (m.focus + 1) % (len(m.inputs) + 2)
			if m.focus < len(m.inputs) {
				return m, m.inputs[m.focus].Focus()
			}
			return m, nil
		case "shift+tab", "up":
			if m.focus < len(m.inputs) {
				m.inputs[m.focus].Blur()
			}
			m.focus = (m.focus + len(m.inputs) + 1) % (len(m.inputs) + 2)
			if m.focus < len(m.inputs) {
				return m, m.inputs[m.focus].Focus()
			}
			return m, nil
		case "left":
			if m.focus == len(m.inputs)+0 {
				m.basis = (m.basis + 4) % 5
				return m, nil
			}
		case "right":
			if m.focus == len(m.inputs)+0 {
				m.basis = (m.basis + 1) % 5
				return m, nil
			}
		case "enter":
			if m.focus == len(m.inputs)+1 {
				return m.generate()
			}
		case "ctrl+g":
			return m.generate()
		}
		if m.focus < len(m.inputs) {
			var cmd tea.Cmd
			m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
			return m, cmd
		}
	}
	return m, nil
}
func (m Model) generate() (tea.Model, tea.Cmd) {
	in, e := m.parse()
	if e != nil {
		m.err = e.Error()
		return m, nil
	}
	r, e := eir.Calculate(in)
	if e != nil {
		m.err = e.Error()
		return m, nil
	}
	m.err = ""
	m.status = ""
	m.result = &r
	m.row = 0
	m.col = 0
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	return m, nil
}
