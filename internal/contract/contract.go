package contract

import "time"

type Severity string

const (
	SeverityHold  Severity = "hold"
	SeverityBlock Severity = "block"
)

type Criterion struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Severity Severity `json:"severity"`
	Expected string   `json:"expected"`
	Evidence []string `json:"evidence"`
}

type Contract struct {
	Customer    string      `json:"customer"`
	GeneratedAt time.Time   `json:"generated_at"`
	Criteria    []Criterion `json:"criteria"`
}
