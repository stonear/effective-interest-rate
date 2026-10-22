package eir

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

var csvHeader = []string{
	"Date", "Principal", "Principal repayment", "Year fraction", "Contractual rate (%)",
	"Contractual interest", "Effective interest expense", "Amortisation", "Amortised cost",
}

func WriteCSV(destination io.Writer, result Result) error {
	writer := csv.NewWriter(destination)
	if err := writer.Write(csvHeader); err != nil {
		return err
	}
	for _, row := range result.Rows {
		record := []string{
			row.Date.Format("2006-01-02"),
			csvNumber(row.Principal),
			csvNumber(row.PrincipalRepayment),
			csvNumber(row.YearFraction),
			csvNumber(row.ContractualRate * 100),
			csvNumber(row.ContractualInterest),
			csvNumber(row.EffectiveInterestExpense),
			csvNumber(row.Amortisation),
			csvNumber(row.AmortisedCost),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func csvNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// ExportCSV writes a new schedule file without replacing an existing export.
func ExportCSV(path string, result Result) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	writeErr := WriteCSV(file, result)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
