package cli

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/braincrew/agentsail/internal/adapter"
	"github.com/braincrew/agentsail/internal/checks"
	"github.com/braincrew/agentsail/internal/contract"
	"github.com/braincrew/agentsail/internal/evidence"
	"github.com/braincrew/agentsail/internal/install"
	"github.com/braincrew/agentsail/internal/loadtest"
	"github.com/braincrew/agentsail/internal/render"
	"github.com/braincrew/agentsail/pkg/version"
)

func Run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Println("agentsail", version.Version)
		return 0
	case "doctor":
		fmt.Println("agentsail doctor: ok")
		fmt.Println("state dir:", evidence.StateDir())
		fmt.Println("install dir:", install.InstallDir())
		return 0
	case "init":
		return initCmd(args[1:])
	case "collect":
		return collect(args[1:])
	case "compile":
		return compileCmd(args[1:])
	case "check":
		return checkCmd(args[1:])
	case "verdict":
		return verdictCmd(args[1:])
	case "report":
		return reportCmd(args[1:])
	case "ci":
		return ci(args[1:])
	case "loadtest":
		return loadtestCmd(args[1:])
	default:
		fmt.Printf("unknown command %q\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Println(`agentsail commands:
  agentsail collect <customer>
  agentsail init [path]
  agentsail compile --customer <customer>
  agentsail check --customer <customer> --target mock:support_agent_v12
  agentsail verdict --customer <customer>
  agentsail ci --customer <customer> --target <target> [--report] [--open] [--cmux-alert] [--soft-exit]
  agentsail loadtest init|install|doctor|explain|run|tui [--config agentsail.loadtest.yaml] [--dry-run] [--no-install]
  agentsail report .agentsail/runs/<run>.json [--open]
  agentsail version
  agentsail doctor`)
}

func collect(args []string) int {
	if len(args) != 1 {
		fmt.Println("usage: agentsail collect <customer>")
		return 2
	}
	if err := evidence.Collect(args[0]); err != nil {
		fmt.Println("collect failed:", err)
		return 1
	}
	fmt.Printf("collected fixture context for %s into .agentsail/cache/%s\n", args[0], args[0])
	return 0
}

func initCmd(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	if fs.Parse(args) != nil {
		return 2
	}
	path := "."
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	summary, err := install.Init(path)
	if err != nil {
		fmt.Println("init failed:", err)
		return 1
	}
	install.PrintSummary(summary)
	return 0
}

func compileCmd(args []string) int {
	fs := flag.NewFlagSet("compile", flag.ContinueOnError)
	customer := fs.String("customer", "", "customer id")
	if fs.Parse(args) != nil {
		return 2
	}
	compiled, err := compileContract(*customer)
	if err != nil {
		fmt.Println("compile failed:", err)
		return 1
	}
	path := contractPath(*customer)
	if err := evidence.WriteJSON(path, compiled); err != nil {
		fmt.Println("write failed:", err)
		return 1
	}
	fmt.Println(path)
	return 0
}

func checkCmd(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	customer := fs.String("customer", "", "customer id")
	target := fs.String("target", "mock:support_agent_v12", "target adapter")
	if fs.Parse(args) != nil {
		return 2
	}
	report, err := runGate(*customer, *target, false, false)
	if err != nil {
		fmt.Println("check failed:", err)
		return 1
	}
	render.PrintSummary(report)
	return 0
}

func verdictCmd(args []string) int {
	fs := flag.NewFlagSet("verdict", flag.ContinueOnError)
	customer := fs.String("customer", "", "customer id")
	if fs.Parse(args) != nil {
		return 2
	}
	path, err := evidence.LatestRunPath(*customer)
	if err != nil {
		fmt.Println("verdict failed:", err)
		return 1
	}
	var report render.RunReport
	if err := evidence.ReadJSON(path, &report); err != nil {
		fmt.Println("verdict failed:", err)
		return 1
	}
	fmt.Println(report.Verdict)
	if report.Verdict == checks.VerdictShip {
		return 0
	}
	return 1
}

func reportCmd(args []string) int {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	openReport := fs.Bool("open", false, "open generated report")
	if fs.Parse(args) != nil || fs.NArg() != 1 {
		fmt.Println("usage: agentsail report .agentsail/runs/<run>.json [--open]")
		return 2
	}
	var report render.RunReport
	if err := evidence.ReadJSON(fs.Arg(0), &report); err != nil {
		fmt.Println("report failed:", err)
		return 1
	}
	path, err := render.WriteHTML(report)
	if err != nil {
		fmt.Println("report failed:", err)
		return 1
	}
	fmt.Println(path)
	if *openReport {
		if err := render.Open(path); err != nil {
			fmt.Println("open failed:", err)
			return 1
		}
	}
	return 0
}

func ci(args []string) int {
	fs := flag.NewFlagSet("ci", flag.ContinueOnError)
	customer := fs.String("customer", "", "customer id")
	target := fs.String("target", "", "mock:<name> or http(s) URL")
	writeReport := fs.Bool("report", false, "write HTML report")
	openReport := fs.Bool("open", false, "open HTML report")
	cmuxAlert := fs.Bool("cmux-alert", false, "emit cmux OSC9 alert")
	softExit := fs.Bool("soft-exit", false, "return zero even for HOLD/BLOCK")
	if fs.Parse(args) != nil {
		return 2
	}
	report, err := runGate(*customer, *target, *writeReport || *openReport, *openReport)
	if err != nil {
		fmt.Println("ci failed:", err)
		return 1
	}
	render.PrintSummary(report)
	if *cmuxAlert {
		render.Alert(fmt.Sprintf("Agent Sail %s: %s", report.Customer, report.Verdict))
	}
	if report.Verdict != checks.VerdictShip && !*softExit {
		return 1
	}
	return 0
}

func loadtestCmd(args []string) int {
	if len(args) == 0 {
		fmt.Println("usage: agentsail loadtest init|install|doctor|explain|run|tui [--config agentsail.loadtest.yaml] [--dry-run] [--no-install]")
		return 2
	}
	switch args[0] {
	case "init":
		fs := flag.NewFlagSet("loadtest init", flag.ContinueOnError)
		configPath := fs.String("config", loadtest.DefaultConfigPath, "load-test YAML path")
		if fs.Parse(args[1:]) != nil {
			return 2
		}
		if err := loadtest.InitConfig(*configPath); err != nil {
			fmt.Println("loadtest init failed:", err)
			return 1
		}
		fmt.Println(*configPath)
		fmt.Println("locust/agentsail/locustfile.py")
		return 0
	case "explain":
		loadtest.Explain()
		return 0
	case "install":
		fs := flag.NewFlagSet("loadtest install", flag.ContinueOnError)
		configPath := fs.String("config", loadtest.DefaultConfigPath, "load-test YAML path")
		if fs.Parse(args[1:]) != nil {
			return 2
		}
		cfg, err := loadtest.ReadConfig(*configPath)
		if err != nil {
			fmt.Println("loadtest install failed:", err)
			return 1
		}
		if err := loadtest.Install(cfg); err != nil {
			fmt.Println("loadtest install failed:", err)
			return 1
		}
		return 0
	case "doctor":
		fs := flag.NewFlagSet("loadtest doctor", flag.ContinueOnError)
		configPath := fs.String("config", loadtest.DefaultConfigPath, "load-test YAML path")
		if fs.Parse(args[1:]) != nil {
			return 2
		}
		cfg, err := loadtest.ReadConfig(*configPath)
		if err != nil {
			fmt.Println("loadtest doctor failed:", err)
			return 1
		}
		if err := loadtest.Doctor(cfg); err != nil {
			fmt.Println("loadtest doctor failed:", err)
			return 1
		}
		return 0
	case "run":
		fs := flag.NewFlagSet("loadtest run", flag.ContinueOnError)
		configPath := fs.String("config", loadtest.DefaultConfigPath, "load-test YAML path")
		dryRun := fs.Bool("dry-run", false, "print Locust command without executing")
		noInstall := fs.Bool("no-install", false, "fail instead of auto-installing Locust when missing")
		if fs.Parse(args[1:]) != nil {
			return 2
		}
		cfg, err := loadtest.ReadConfig(*configPath)
		if err != nil {
			fmt.Println("loadtest run failed:", err)
			return 1
		}
		if err := loadtest.Run(cfg, *dryRun, *noInstall); err != nil {
			fmt.Println("loadtest run failed:", err)
			return 1
		}
		return 0
	case "tui", "watch":
		fs := flag.NewFlagSet("loadtest tui", flag.ContinueOnError)
		configPath := fs.String("config", loadtest.DefaultConfigPath, "load-test YAML path")
		dryRun := fs.Bool("dry-run", false, "render the load-test board without executing Locust")
		noInstall := fs.Bool("no-install", false, "fail instead of auto-installing Locust when missing")
		if fs.Parse(args[1:]) != nil {
			return 2
		}
		cfg, err := loadtest.ReadConfig(*configPath)
		if err != nil {
			fmt.Println("loadtest tui failed:", err)
			return 1
		}
		if err := loadtest.RunTUI(cfg, *dryRun, *noInstall); err != nil {
			fmt.Println("loadtest tui failed:", err)
			return 1
		}
		return 0
	default:
		fmt.Printf("unknown loadtest command %q\n", args[0])
		return 2
	}
}

func runGate(customer string, target string, writeHTML bool, openHTML bool) (render.RunReport, error) {
	if customer == "" {
		return render.RunReport{}, fmt.Errorf("--customer is required")
	}
	if target == "" {
		return render.RunReport{}, fmt.Errorf("--target is required")
	}
	if err := evidence.Collect(customer); err != nil {
		return render.RunReport{}, err
	}
	compiled, err := compileContract(customer)
	if err != nil {
		return render.RunReport{}, err
	}
	if err := evidence.WriteJSON(contractPath(customer), compiled); err != nil {
		return render.RunReport{}, err
	}
	targetResult, err := adapter.Run(target)
	if err != nil {
		return render.RunReport{}, err
	}
	results := checks.Run(compiled, targetResult)
	verdict, risk := checks.Decide(results)
	runPath := evidence.NextRunPath(customer)
	report := render.RunReport{
		RunID:       strings.TrimSuffix(filepath.Base(runPath), ".json"),
		Customer:    customer,
		Target:      target,
		GeneratedAt: time.Now().UTC(),
		Contract:    compiled,
		Checks:      results,
		Verdict:     verdict,
		RiskScore:   risk,
	}
	if writeHTML {
		htmlPath, err := render.WriteHTML(report)
		if err != nil {
			return render.RunReport{}, err
		}
		report.ReportPath = htmlPath
		if openHTML {
			if err := render.Open(htmlPath); err != nil {
				return render.RunReport{}, err
			}
		}
	}
	if err := evidence.WriteJSON(runPath, report); err != nil {
		return render.RunReport{}, err
	}
	return report, nil
}

func compileContract(customer string) (contract.Contract, error) {
	return contract.Compile(customer, filepath.Join(evidence.StateDir(), "cache"))
}

func contractPath(customer string) string {
	return filepath.Join(evidence.StateDir(), "contracts", customer+"-contract.json")
}
