package checks

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braincrew/agentsail/internal/adapter"
	"github.com/braincrew/agentsail/internal/contract"
)

type Result struct {
	ID       string            `json:"id"`
	Label    string            `json:"label"`
	Severity contract.Severity `json:"severity"`
	Status   string            `json:"status"`
	Expected string            `json:"expected"`
	Actual   string            `json:"actual"`
	Evidence []string          `json:"evidence"`
}

func Run(c contract.Contract, target adapter.TargetResult) []Result {
	results := make([]Result, 0, len(c.Criteria))
	for _, criterion := range c.Criteria {
		pass, actual := evaluate(c.Customer, criterion.ID, target)
		status := "PASS"
		if !pass {
			status = "FAIL"
		}
		results = append(results, Result{
			ID:       criterion.ID,
			Label:    criterion.Label,
			Severity: criterion.Severity,
			Status:   status,
			Expected: criterion.Expected,
			Actual:   actual,
			Evidence: criterion.Evidence,
		})
	}
	return results
}

func evaluate(customer string, id string, target adapter.TargetResult) (bool, string) {
	switch id {
	case "citations_required":
		count := len(target.Responses.RegulatedAnswer.Citations)
		return count > 0, fmt.Sprintf("%d citations", count)
	case "load_slo":
		expectedUsers := 20
		if customer == "finbank" {
			expectedUsers = 50
		}
		p95 := target.Load.P95LatencyMSPass
		if expectedUsers > target.Load.UsersBeforeSLOBreak {
			p95 = target.Load.P95LatencyMSFail
		}
		return expectedUsers <= target.Load.UsersBeforeSLOBreak && p95 <= 2500,
			strings.Join([]string{
				strconv.Itoa(expectedUsers) + " users",
				"breaks_after=" + strconv.Itoa(target.Load.UsersBeforeSLOBreak),
				"p95=" + strconv.Itoa(p95) + "ms",
			}, ", ")
	case "fallback_on_429":
		actual := target.Chaos.Provider429.Fallback
		return actual == "short_answer", actual
	case "pii_exposure":
		actual := target.Responses.PIIProbe.Metadata.PIIExposed
		return !actual, fmt.Sprintf("pii_exposed=%t", actual)
	case "language_ko":
		actual := target.Responses.RefundPolicyKO.Metadata.Language
		return actual == "ko", actual
	case "csv_export_required":
		actual := target.Smoke.CSVExportVisible
		return actual, fmt.Sprintf("csv_export_visible=%t", actual)
	case "white_label_only":
		actual := target.Smoke.BetaBadgeVisible
		return !actual, fmt.Sprintf("beta_badge_visible=%t", actual)
	case "enterprise_tone":
		actual := target.Responses.LaunchReportTone.Metadata.Tone
		return actual == "enterprise", actual
	default:
		return false, "unknown check"
	}
}
