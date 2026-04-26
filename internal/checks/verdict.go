package checks

import "github.com/braincrew/agentsail/internal/contract"

type Verdict string

const (
	VerdictShip  Verdict = "SHIP"
	VerdictHold  Verdict = "HOLD"
	VerdictBlock Verdict = "BLOCK"
)

func Decide(results []Result) (Verdict, int) {
	verdict := VerdictShip
	risk := 0
	for _, result := range results {
		if result.Status != "FAIL" {
			continue
		}
		switch result.Severity {
		case contract.SeverityBlock:
			verdict = VerdictBlock
			risk += 50
		case contract.SeverityHold:
			if verdict != VerdictBlock {
				verdict = VerdictHold
			}
			risk += 20
		}
	}
	if risk > 100 {
		risk = 100
	}
	return verdict, risk
}
