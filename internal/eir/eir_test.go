package eir

import (
	"math"
	"testing"
	"time"
)

func d(s string) time.Time { v, _ := time.Parse("2006-01-02", s); return v }
func TestWorkbook(t *testing.T) {
	r, e := Calculate(Input{Settlement: d("2025-01-10"), Maturity: d("2025-03-10"), Principal: 1e10, IssuePricePercent: 100, Cashback: 45205479.45, ContractualRate: .055, Basis: ActualActual})
	if e != nil {
		t.Fatal(e)
	}
	if math.Abs(r.Rate-.08315766747244197) > 1e-9 || len(r.Rows) != 60 || math.Abs(r.ClosingAmortisedCost) > .01 {
		t.Fatalf("unexpected result: rate %.12f rows %d closing %.4f", r.Rate, len(r.Rows), r.ClosingAmortisedCost)
	}
}
func TestBasis(t *testing.T) {
	for _, b := range []Basis{US30360, ActualActual, Actual360, Actual365, European30360} {
		f, e := YearFraction(d("2025-01-10"), d("2025-01-11"), b)
		if e != nil || f <= 0 {
			t.Fatalf("%v: %v %v", b, f, e)
		}
	}
	f, _ := YearFraction(d("2024-02-28"), d("2024-02-29"), ActualActual)
	if math.Abs(f-1.0/366) > 1e-12 {
		t.Fatal(f)
	}
}
