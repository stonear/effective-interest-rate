package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/stonear/effective-interest-rate/internal/ui"
)

func main() {
	if _, err := tea.NewProgram(ui.New()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
