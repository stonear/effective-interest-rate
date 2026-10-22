package ui

import (
	"math"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSampleFlow(t *testing.T) {
	m := New().(Model)
	next, _ := m.generate()
	m = next.(Model)
	if m.result == nil || math.Abs(m.result.Rate-.08315766747244197) > 1e-9 {
		t.Fatal(m.err)
	}
	if !strings.Contains(m.View().Content, "Effective interest expense") {
		t.Fatal("missing schedule label")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'e'})
	if next.(Model).result != nil {
		t.Fatal("did not return to form")
	}
}
func TestAmount(t *testing.T) {
	for _, s := range []string{"NaN", "Inf", "12,34", "1,23,456", ""} {
		if _, e := parseAmount(s); e == nil {
			t.Fatalf("accepted %q", s)
		}
	}
	v, e := parseAmount("1,234.50")
	if e != nil || v != 1234.5 {
		t.Fatalf("%v %v", v, e)
	}
}

func TestReturnToFormAfterGenerateButton(t *testing.T) {
	model := New().(Model)
	model.focus = len(model.inputs) + 1
	generated, _ := model.generate()
	result := generated.(Model)
	if result.result == nil {
		t.Fatal(result.err)
	}
	edited, _ := result.Update(tea.KeyPressMsg{Code: 'e'})
	if edited.(Model).result != nil {
		t.Fatal("expected form")
	}
}

func TestGenerateWithEnter(t *testing.T) {
	model := New().(Model)
	for range len(model.inputs) + 1 {
		updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		model = updated.(Model)
	}
	if model.focus != len(model.inputs)+1 {
		t.Fatalf("focus %d", model.focus)
	}
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	result := updated.(Model)
	if result.result == nil {
		t.Fatalf("generation failed: %s", result.err)
	}
	_ = result.View()
}

func TestResultScreenExportsCSV(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	model := New().(Model)
	generated, _ := model.generate()
	model = generated.(Model)
	updated, _ := model.Update(tea.KeyPressMsg{Code: 'c'})
	model = updated.(Model)
	if _, err := os.Stat("eir-schedule-2025-01-10-to-2025-03-10.csv"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(model.View().Content, "CSV saved") {
		t.Fatal("missing export confirmation")
	}
}
