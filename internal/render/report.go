package render

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/braincrew/agentsail/internal/checks"
	"github.com/braincrew/agentsail/internal/contract"
)

type RunReport struct {
	RunID       string            `json:"run_id"`
	Customer    string            `json:"customer"`
	Target      string            `json:"target"`
	GeneratedAt time.Time         `json:"generated_at"`
	Contract    contract.Contract `json:"contract"`
	Checks      []checks.Result   `json:"checks"`
	Verdict     checks.Verdict    `json:"verdict"`
	RiskScore   int               `json:"risk_score"`
	ReportPath  string            `json:"report_path,omitempty"`
}

func PrintSummary(report RunReport) {
	fmt.Printf("Agent Sail verdict for %s: %s (risk=%d)\n", report.Customer, report.Verdict, report.RiskScore)
	for _, check := range report.Checks {
		fmt.Printf("  [%s] %s (%s) actual=%s\n", check.Status, check.Label, check.Severity, check.Actual)
	}
}

func Alert(message string) {
	fmt.Printf("\033]9;%s\033\\\n", message)
}

func WriteHTML(report RunReport) (string, error) {
	path := filepath.Join(".agentsail", "reports", strings.TrimSuffix(filepath.Base(report.RunID), ".json")+".html")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if err := reportTemplate.Execute(file, report); err != nil {
		return "", err
	}
	return path, nil
}

func Open(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("open is not supported on %s", runtime.GOOS)
	}
	return cmd.Start()
}

var reportTemplate = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Agent Sail {{.Customer}} {{.Verdict}}</title>
  <style>
    body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;margin:0;background:#f7f7f4;color:#171717}
    header{padding:40px 56px;background:#171717;color:white}
    main{padding:32px 56px;max-width:1120px}
    .verdict{font-size:56px;font-weight:800;letter-spacing:0}
    .SHIP{color:#16a34a}.HOLD{color:#f59e0b}.BLOCK{color:#ef4444}
    .grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:16px}
    .card{background:white;border:1px solid #ddd;border-radius:8px;padding:18px}
    .FAIL{border-left:8px solid #ef4444}.PASS{border-left:8px solid #16a34a}
    code{background:#eee;padding:2px 6px;border-radius:4px}
  </style>
</head>
<body>
  <header>
    <div>Same agent. Different customer. Different launch gate.</div>
    <div class="verdict {{.Verdict}}">VERDICT: {{.Verdict}}</div>
    <div>{{.Customer}} · risk {{.RiskScore}} · {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}</div>
  </header>
  <main>
    <h2>Why This Gate Decided {{.Verdict}}</h2>
    <div class="grid">
      {{range .Checks}}
      <section class="card {{.Status}}">
        <h3>{{.Status}} · {{.Label}}</h3>
        <p><strong>Expected:</strong> {{.Expected}}</p>
        <p><strong>Actual:</strong> {{.Actual}}</p>
        <p><strong>Severity:</strong> <code>{{.Severity}}</code></p>
        <p><strong>Evidence:</strong> {{range .Evidence}}<code>{{.}}</code> {{end}}</p>
      </section>
      {{end}}
    </div>
  </main>
</body>
</html>
`))
