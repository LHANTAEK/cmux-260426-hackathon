package contract

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Compile(customer string, cacheDir string) (Contract, error) {
	if customer == "" {
		return Contract{}, fmt.Errorf("customer is required")
	}
	if _, err := os.Stat(filepath.Join(cacheDir, customer)); err != nil {
		return Contract{}, fmt.Errorf("cache for %q not found: run agentsail collect %s first", customer, customer)
	}

	contract := Contract{
		Customer:    customer,
		GeneratedAt: time.Now().UTC(),
		Criteria:    criteriaFor(customer),
	}
	if len(contract.Criteria) == 0 {
		return Contract{}, fmt.Errorf("no release contract rules for customer %q", customer)
	}
	return contract, nil
}

func criteriaFor(customer string) []Criterion {
	switch customer {
	case "finbank":
		return []Criterion{
			{ID: "citations_required", Label: "regulated answers include citations", Severity: SeverityHold, Expected: "non-empty citation list", Evidence: []string{"messenger.md:5", "github.json:6"}},
			{ID: "load_slo", Label: "50 concurrent users meet SLO", Severity: SeverityHold, Expected: "p95 <= 2500ms at 50 users", Evidence: []string{"messenger.md:6", "notion.md:4"}},
			{ID: "fallback_on_429", Label: "provider 429 returns short fallback", Severity: SeverityHold, Expected: "short_answer", Evidence: []string{"messenger.md:7"}},
			{ID: "pii_exposure", Label: "PII probe does not expose private data", Severity: SeverityHold, Expected: "pii_exposed=false", Evidence: []string{"email.md:3"}},
		}
	case "retailco":
		return []Criterion{
			{ID: "language_ko", Label: "refund policy answer stays Korean", Severity: SeverityHold, Expected: "ko", Evidence: []string{"messenger.md:3"}},
			{ID: "load_slo", Label: "20 concurrent users meet SLO", Severity: SeverityHold, Expected: "p95 <= 2500ms at 20 users", Evidence: []string{"email.md:3", "github.json:7", "notion.md:4"}},
		}
	case "acme-bank":
		return []Criterion{
			{ID: "csv_export_required", Label: "missing CSV export", Severity: SeverityBlock, Expected: "CSV export visible", Evidence: []string{"messenger.md:5", "github.json:6"}},
			{ID: "white_label_only", Label: "beta badge exposed", Severity: SeverityBlock, Expected: "no beta badge", Evidence: []string{"messenger.md:6", "github.json:6"}},
			{ID: "enterprise_tone", Label: "tone drift", Severity: SeverityBlock, Expected: "enterprise", Evidence: []string{"messenger.md:7", "github.json:6"}},
		}
	default:
		return nil
	}
}
