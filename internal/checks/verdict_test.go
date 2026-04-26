package checks

import (
	"testing"

	"github.com/braincrew/agentsail/internal/contract"
)

func TestDecide(t *testing.T) {
	tests := []struct {
		name    string
		results []Result
		want    Verdict
		risk    int
	}{
		{name: "ship", results: []Result{{Status: "PASS", Severity: contract.SeverityHold}}, want: VerdictShip, risk: 0},
		{name: "hold", results: []Result{{Status: "FAIL", Severity: contract.SeverityHold}}, want: VerdictHold, risk: 20},
		{name: "block", results: []Result{{Status: "FAIL", Severity: contract.SeverityBlock}, {Status: "FAIL", Severity: contract.SeverityHold}}, want: VerdictBlock, risk: 70},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, risk := Decide(tt.results)
			if got != tt.want || risk != tt.risk {
				t.Fatalf("Decide() = %s/%d, want %s/%d", got, risk, tt.want, tt.risk)
			}
		})
	}
}
