package eir

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"time"
)

const maxTermDays = 36_600
const solverIterations = 256

type Input struct {
	Settlement        time.Time
	Maturity          time.Time
	CustomerName      string
	DealReference     string
	Principal         float64
	IssuePricePercent float64
	Cashback          float64
	TransactionCosts  float64
	ContractualRate   float64
	Basis             Basis
}

type Row struct {
	Date                     time.Time
	Principal                float64
	PrincipalRepayment       float64
	YearFraction             float64
	ContractualRate          float64
	ContractualInterest      float64
	EffectiveInterestExpense float64
	Amortisation             float64
	AmortisedCost            float64
}

type Result struct {
	Input                         Input
	Rate                          float64
	InitialAmortisedCost          float64
	TotalContractualInterest      float64
	TotalEffectiveInterestExpense float64
	TotalAmortisation             float64
	ClosingAmortisedCost          float64
	Rows                          []Row
}

type cashFlow struct {
	date         time.Time
	yearFraction float64
	payment      float64
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func initialAmortisedCost(input Input) float64 {
	issueProceeds := input.Principal * input.IssuePricePercent / 100
	return issueProceeds - input.Cashback - input.TransactionCosts
}

func Validate(input Input) error {
	if input.Settlement.IsZero() || input.Maturity.IsZero() || !date(input.Maturity).After(date(input.Settlement)) {
		return errors.New("maturity must follow settlement")
	}
	if days(input.Settlement, input.Maturity) > maxTermDays {
		return errors.New("term exceeds 36,600 days")
	}
	if input.Basis < US30360 || input.Basis > European30360 {
		return errors.New("invalid day-count convention")
	}

	amounts := []float64{
		input.Principal,
		input.IssuePricePercent,
		input.Cashback,
		input.TransactionCosts,
		input.ContractualRate,
	}
	for _, amount := range amounts {
		if !finite(amount) {
			return errors.New("amounts and rates must be finite")
		}
	}
	if input.Principal <= 0 || input.IssuePricePercent <= 0 || input.Cashback < 0 || input.TransactionCosts < 0 || input.ContractualRate < 0 {
		return errors.New("principal and issue price must be positive; cashback, costs and rate cannot be negative")
	}
	if initial := initialAmortisedCost(input); !finite(initial) || initial <= 0 {
		return errors.New("initial amortised cost must be positive and finite")
	}
	return nil
}

func Calculate(input Input) (Result, error) {
	if err := Validate(input); err != nil {
		return Result{}, err
	}
	input.Settlement = date(input.Settlement)
	input.Maturity = date(input.Maturity)

	flows, err := dailyCashFlows(input)
	if err != nil {
		return Result{}, err
	}
	initial := initialAmortisedCost(input)
	rate, err := solveRate(flows, initial)
	if err != nil {
		return Result{}, err
	}
	return buildSchedule(input, flows, initial, rate)
}

func dailyCashFlows(input Input) ([]cashFlow, error) {
	periodCount := days(input.Settlement, input.Maturity)
	flows := make([]cashFlow, 0, periodCount)
	previousDate := input.Settlement

	for period := range periodCount {
		paymentDate := previousDate.AddDate(0, 0, 1)
		fraction, err := YearFraction(previousDate, paymentDate, input.Basis)
		if err != nil {
			return nil, err
		}
		payment := input.Principal * input.ContractualRate * fraction
		if period == periodCount-1 {
			payment += input.Principal
		}
		flows = append(flows, cashFlow{
			date:         paymentDate,
			yearFraction: fraction,
			payment:      payment,
		})
		previousDate = paymentDate
	}
	return flows, nil
}

// discountCashFlows works backwards from maturity. This avoids multiplying
// the carrying amount by a potentially large accumulation factor at each day.
func discountCashFlows(flows []cashFlow, rate float64) float64 {
	presentValue := 0.0
	for _, flow := range slices.Backward(flows) {
		factor := 1 + rate*flow.yearFraction
		if factor <= 0 {
			return math.Inf(1)
		}
		presentValue = (presentValue + flow.payment) / factor
	}
	return presentValue
}

func solveRate(flows []cashFlow, initial float64) (float64, error) {
	maximumFraction := 0.0
	for _, flow := range flows {
		maximumFraction = math.Max(maximumFraction, flow.yearFraction)
	}
	if maximumFraction == 0 {
		return 0, errors.New("all periods have zero year fraction")
	}

	// The lower bound keeps every daily accumulation factor positive.
	lower := -1/maximumFraction + 1e-12
	upper := 1.0
	for expansion := 0; expansion < 100 && discountCashFlows(flows, upper) > initial; expansion++ {
		upper *= 2
	}
	if discountCashFlows(flows, lower) < initial || discountCashFlows(flows, upper) > initial || !finite(discountCashFlows(flows, upper)) {
		return 0, errors.New("could not bracket effective interest rate")
	}

	for range solverIterations {
		middle := lower + (upper-lower)/2
		if discountCashFlows(flows, middle) > initial {
			lower = middle
		} else {
			upper = middle
		}
	}
	return lower + (upper-lower)/2, nil
}

func buildSchedule(input Input, flows []cashFlow, initial, rate float64) (Result, error) {
	result := Result{
		Input:                input,
		Rate:                 rate,
		InitialAmortisedCost: initial,
		Rows:                 make([]Row, 0, len(flows)+1),
	}
	result.Rows = append(result.Rows, Row{
		Date:          input.Settlement,
		Principal:     input.Principal,
		AmortisedCost: initial,
	})

	carryingAmount := initial
	for period, flow := range flows {
		contractualInterest := input.Principal * input.ContractualRate * flow.yearFraction
		effectiveInterest := carryingAmount * rate * flow.yearFraction
		amortisation := contractualInterest - effectiveInterest

		principalRepayment := 0.0
		if period == len(flows)-1 {
			principalRepayment = input.Principal
		}
		carryingAmount += effectiveInterest - contractualInterest - principalRepayment

		result.TotalContractualInterest += contractualInterest
		result.TotalEffectiveInterestExpense += effectiveInterest
		result.TotalAmortisation += amortisation
		result.Rows = append(result.Rows, Row{
			Date:                     flow.date,
			Principal:                input.Principal,
			PrincipalRepayment:       principalRepayment,
			YearFraction:             flow.yearFraction,
			ContractualRate:          input.ContractualRate,
			ContractualInterest:      contractualInterest,
			EffectiveInterestExpense: effectiveInterest,
			Amortisation:             amortisation,
			AmortisedCost:            carryingAmount,
		})
	}
	result.ClosingAmortisedCost = carryingAmount
	allowedResidual := math.Max(0.001, input.Principal*1e-12)
	if math.Abs(carryingAmount) > allowedResidual {
		return Result{}, fmt.Errorf("rate did not converge: closing balance %.6f", carryingAmount)
	}
	return result, nil
}
