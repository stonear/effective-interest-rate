package eir

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteCSVIncludesEveryScheduleRow(t *testing.T) {
	settlement := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	result, err := Calculate(Input{Settlement: settlement, Maturity: settlement.AddDate(0, 0, 2), Principal: 1000, IssuePricePercent: 100, Cashback: 10, ContractualRate: .055, Basis: Actual365})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := WriteCSV(&output, result); err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(&output).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 4 {
		t.Fatalf("got %d rows", len(records))
	}
	if records[0][0] != "Date" || records[0][8] != "Amortised cost" {
		t.Fatalf("unexpected header: %v", records[0])
	}
	if records[1][0] != "2025-01-10" || records[3][0] != "2025-01-12" {
		t.Fatalf("unexpected dates: %v", records)
	}
	if records[1][8] != "990" {
		t.Fatalf("unexpected initial amortised cost: %s", records[1][8])
	}
}

func TestExportCSVDoesNotOverwriteExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.csv")
	result := Result{Rows: []Row{{Date: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), AmortisedCost: 990}}}
	if err := ExportCSV(path, result); err != nil {
		t.Fatal(err)
	}
	if err := ExportCSV(path, result); err == nil {
		t.Fatal("expected existing-file error")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(contents, []byte("2025-01-10")) {
		t.Fatal("missing schedule date")
	}
}
