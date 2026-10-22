package eir

import (
	"errors"
	"time"
)

type Basis int

const (
	US30360 Basis = iota
	ActualActual
	Actual360
	Actual365
	European30360
)

var basisNames = []string{
	"US (NASD) 30/360",
	"Actual/Actual",
	"Actual/360",
	"Actual/365",
	"European 30/360",
}

func (basis Basis) String() string {
	if basis < US30360 || basis > European30360 {
		return "Unknown"
	}
	return basisNames[basis]
}

func Bases() []string {
	return append([]string(nil), basisNames...)
}

func date(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func days(start, end time.Time) int {
	return int(date(end).Sub(date(start)).Hours() / 24)
}

func leap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func lastFebruary(value time.Time) bool {
	return value.Month() == time.February && value.AddDate(0, 0, 1).Month() != time.February
}

func YearFraction(start, end time.Time, basis Basis) (float64, error) {
	if basis < US30360 || basis > European30360 {
		return 0, errors.New("invalid day-count convention")
	}
	start = date(start)
	end = date(end)
	if end.Before(start) {
		return 0, errors.New("end date precedes start date")
	}
	if start.Equal(end) {
		return 0, nil
	}

	switch basis {
	case Actual360:
		return float64(days(start, end)) / 360, nil
	case Actual365:
		return float64(days(start, end)) / 365, nil
	case ActualActual:
		return actualActualFraction(start, end), nil
	case US30360:
		return usThirty360Fraction(start, end), nil
	default:
		return europeanThirty360Fraction(start, end), nil
	}
}

func actualActualFraction(start, end time.Time) float64 {
	fraction := 0.0
	for current := start; current.Before(end); {
		nextYear := time.Date(current.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := nextYear
		if end.Before(periodEnd) {
			periodEnd = end
		}
		daysInYear := 365.0
		if leap(current.Year()) {
			daysInYear = 366
		}
		fraction += float64(days(current, periodEnd)) / daysInYear
		current = periodEnd
	}
	return fraction
}

func europeanThirty360Fraction(start, end time.Time) float64 {
	startDay := min(start.Day(), 30)
	endDay := min(end.Day(), 30)
	return thirty360Fraction(start, end, startDay, endDay)
}

func usThirty360Fraction(start, end time.Time) float64 {
	startDay := start.Day()
	endDay := end.Day()

	if lastFebruary(start) || startDay == 31 {
		startDay = 30
	}
	if lastFebruary(start) && lastFebruary(end) {
		endDay = 30
	}
	if endDay == 31 {
		if startDay < 30 {
			endDay = 1
			end = end.AddDate(0, 1, 0)
		} else {
			endDay = 30
		}
	}
	return thirty360Fraction(start, end, startDay, endDay)
}

func thirty360Fraction(start, end time.Time, startDay, endDay int) float64 {
	fullYears := (end.Year() - start.Year()) * 360
	fullMonths := (int(end.Month()) - int(start.Month())) * 30
	return float64(fullYears+fullMonths+endDay-startDay) / 360
}
