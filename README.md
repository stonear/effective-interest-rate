# Effective interest rate calculator

A terminal app for calculating the effective interest rate (EIR) and a daily amortised cost schedule for a deposit financial liability. The interface uses [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Run

```sh
go run ./cmd/eir
```

The form opens with example values. Replace them with your transaction details, enter dates as `YYYY-MM-DD`, and enter interest rates as percentages (for example, `5.5` for 5.5%).

## Controls

| Screen  | Keys                                       | Action                         |
| ------- | ------------------------------------------ | ------------------------------ |
| Form    | Tab / Shift+Tab                            | Move between fields            |
| Form    | Left / Right on day-count field            | Choose a convention            |
| Form    | Enter on Generate, or Ctrl+G               | Calculate the EIR and schedule |
| Results | Up / Down, Page Up / Page Down, Home / End | Move through dates             |
| Results | Left / Right                               | Move through table columns     |
| Results | `c`                                        | Save the schedule as CSV       |
| Results | `e` or Esc                                 | Return to the form             |
| Either  | Ctrl+C                                     | Quit                           |

## Calculation

The available day-count conventions are US (NASD) 30/360, Actual/Actual, Actual/360, Actual/365, and European 30/360. Actual/Actual is selected by default.

Initial amortised cost equals issue proceeds less upfront cashback and directly attributable transaction costs. Each day's effective interest expense is the opening amortised cost multiplied by the solved EIR and that day's year fraction. Contractual interest is assumed to be paid daily. The schedule records **amortisation** as contractual interest less effective interest expense, then reduces the carrying amount by contractual interest and any principal repayment while adding effective interest expense. The principal is repaid at maturity.

The displayed EIR is an annualised rate applied as `EIR × daily year fraction`; it is not a compounded annual yield. The example values produce an EIR of approximately **8.315767%**. Calculations retain full precision and display amounts to two decimal places.

## CSV export

Press `c` on the results screen to save the complete daily schedule in the current working directory. The filename uses the settlement and maturity dates, for example `eir-schedule-2025-01-10-to-2025-03-10.csv`. The file has the schedule's nine columns, with ungrouped numbers at calculation precision. If the file already exists, the app reports an error and leaves it untouched.

## Scope and accounting assumptions

The app treats cashback as an upfront incentive that reduces the initial carrying amount. Whether a specific incentive qualifies for this treatment requires an accounting assessment under [IFRS 9](https://www.ifrs.org/issued-standards/list-of-standards/ifrs-9-financial-instruments/). The app supports terms up to 36,600 days. It does not model changes to principal, rate resets, impairment, or monthly interest settlement.
