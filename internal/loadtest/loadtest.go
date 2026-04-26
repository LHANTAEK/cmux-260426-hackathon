package loadtest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/braincrew/agentsail/internal/assets"
	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "agentsail.loadtest.yaml"

type Config struct {
	Version int    `yaml:"version"`
	Name    string `yaml:"name"`
	Target  struct {
		BaseURL       string `yaml:"base_url"`
		ChatPath      string `yaml:"chat_path"`
		MetricsURL    string `yaml:"metrics_url"`
		PrometheusURL string `yaml:"prometheus_url"`
	} `yaml:"target"`
	Resources struct {
		Memory struct {
			Limit   string `yaml:"limit"`
			AlertAt string `yaml:"alert_at"`
		} `yaml:"memory"`
	} `yaml:"resources"`
	Locust struct {
		Locustfile string `yaml:"locustfile"`
		ResultsDir string `yaml:"results_dir"`
		Users      int    `yaml:"users"`
		SpawnRate  int    `yaml:"spawn_rate"`
		RunTime    string `yaml:"run_time"`
	} `yaml:"locust"`
	Runtime struct {
		AutoInstall bool     `yaml:"auto_install"`
		Installer   string   `yaml:"installer"`
		Python      string   `yaml:"python"`
		Venv        string   `yaml:"venv"`
		Packages    []string `yaml:"packages"`
	} `yaml:"runtime"`
	SLOs           map[string]MetricSpec `yaml:"slos"`
	LeadingSignals map[string]MetricSpec `yaml:"leading_signals"`
	PromQL         map[string]string     `yaml:"promql"`
}

type MetricSpec struct {
	Metric                  string `yaml:"metric"`
	Unit                    string `yaml:"unit"`
	DisplayUnit             string `yaml:"display_unit"`
	Threshold               any    `yaml:"threshold"`
	Of                      string `yaml:"of"`
	CandidateThreshold      any    `yaml:"candidate_threshold"`
	CandidateThresholdRatio any    `yaml:"candidate_threshold_ratio"`
	Meaning                 string `yaml:"meaning"`
}

func InitConfig(path string) error {
	if path == "" {
		path = DefaultConfigPath
	}
	if err := copyTemplateFile("templates/agentsail.loadtest.yaml", path, 0o644); err != nil {
		return err
	}
	if err := copyTemplateDir("templates/locust", "locust"); err != nil {
		return err
	}
	return nil
}

func ReadConfig(path string) (Config, error) {
	if path == "" {
		path = DefaultConfigPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Target.BaseURL == "" {
		return Config{}, fmt.Errorf("target.base_url is required")
	}
	if cfg.Target.ChatPath == "" {
		cfg.Target.ChatPath = "/chat"
	}
	if cfg.Resources.Memory.Limit == "" {
		cfg.Resources.Memory.Limit = "1g"
	}
	if cfg.Resources.Memory.AlertAt == "" {
		cfg.Resources.Memory.AlertAt = "80%"
	}
	if _, err := ParseMemoryBytes(cfg.Resources.Memory.Limit); err != nil {
		return Config{}, fmt.Errorf("resources.memory.limit: %w", err)
	}
	if _, err := ParsePercent(cfg.Resources.Memory.AlertAt); err != nil {
		return Config{}, fmt.Errorf("resources.memory.alert_at: %w", err)
	}
	if cfg.Locust.Locustfile == "" {
		return Config{}, fmt.Errorf("locust.locustfile is required")
	}
	if cfg.Locust.ResultsDir == "" {
		cfg.Locust.ResultsDir = ".agentsail/loadtests/default"
	}
	if cfg.Locust.Users == 0 {
		cfg.Locust.Users = 1
	}
	if cfg.Locust.SpawnRate == 0 {
		cfg.Locust.SpawnRate = 1
	}
	if cfg.Locust.RunTime == "" {
		cfg.Locust.RunTime = "1m"
	}
	if cfg.Runtime.Python == "" {
		cfg.Runtime.Python = "python3"
	}
	if cfg.Runtime.Installer == "" {
		cfg.Runtime.Installer = "uv"
	}
	if cfg.Runtime.Venv == "" {
		cfg.Runtime.Venv = ".agentsail/loadtests/.venv"
	}
	if len(cfg.Runtime.Packages) == 0 {
		cfg.Runtime.Packages = []string{"locust>=2.42.0", "httpx>=0.28.0"}
	}
	return cfg, nil
}

func Command(cfg Config, locustPath string) *exec.Cmd {
	htmlPath := filepath.Join(cfg.Locust.ResultsDir, "report.html")
	csvPrefix := filepath.Join(cfg.Locust.ResultsDir, "stats")
	args := []string{
		"-f", cfg.Locust.Locustfile,
		"--host", cfg.Target.BaseURL,
		"--headless",
		"--users", fmt.Sprintf("%d", cfg.Locust.Users),
		"--spawn-rate", fmt.Sprintf("%d", cfg.Locust.SpawnRate),
		"--run-time", cfg.Locust.RunTime,
		"--csv", csvPrefix,
		"--html", htmlPath,
	}
	cmd := exec.Command(locustPath, args...)
	cmd.Env = append(os.Environ(),
		"TARGET_HOST="+cfg.Target.BaseURL,
		"CHAT_PATH="+cfg.Target.ChatPath,
	)
	return cmd
}

func Run(cfg Config, dryRun bool, noInstall bool) error {
	if err := os.MkdirAll(cfg.Locust.ResultsDir, 0o755); err != nil {
		return err
	}
	locustPath, err := ResolveLocust(cfg, dryRun || noInstall)
	if err != nil {
		return err
	}
	if !dryRun && !noInstall {
		if err := EnsureLocust(cfg, locustPath); err != nil {
			return err
		}
	}
	cmd := Command(cfg, locustPath)
	fmt.Println(ShellString(cmd))
	if dryRun {
		return nil
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Install(cfg Config) error {
	locustPath := venvBinary(cfg, "locust")
	return EnsureLocust(cfg, locustPath)
}

func Doctor(cfg Config) error {
	path, err := ResolveLocust(cfg, true)
	if err != nil {
		return err
	}
	fmt.Println("loadtest config: ok")
	fmt.Println("locust runner:", path)
	fmt.Println("auto_install:", cfg.Runtime.AutoInstall)
	fmt.Println("installer:", RuntimeInstaller(cfg))
	fmt.Println("venv:", cfg.Runtime.Venv)
	fmt.Println("packages:", strings.Join(cfg.Runtime.Packages, ", "))
	return nil
}

func ResolveLocust(cfg Config, planningOnly bool) (string, error) {
	if path, err := exec.LookPath("locust"); err == nil {
		return path, nil
	}
	venvLocust := venvBinary(cfg, "locust")
	if fileExists(venvLocust) {
		return venvLocust, nil
	}
	if cfg.Runtime.AutoInstall || planningOnly {
		return venvLocust, nil
	}
	return "", fmt.Errorf("locust not found; set runtime.auto_install=true or run agentsail loadtest install")
}

func EnsureLocust(cfg Config, locustPath string) error {
	if fileExists(locustPath) {
		return nil
	}
	if !cfg.Runtime.AutoInstall {
		return fmt.Errorf("locust not found at %s and runtime.auto_install=false", locustPath)
	}
	installer := RuntimeInstaller(cfg)
	fmt.Printf("Installing Locust runtime into %s with %s\n", cfg.Runtime.Venv, installer)
	if installer == "uv" {
		return installWithUV(cfg)
	}
	if installer != "pip" {
		return fmt.Errorf("unsupported runtime.installer %q; use uv, pip, or auto", cfg.Runtime.Installer)
	}
	return installWithPip(cfg)
}

func RuntimeInstaller(cfg Config) string {
	switch strings.ToLower(strings.TrimSpace(cfg.Runtime.Installer)) {
	case "", "uv":
		return "uv"
	case "auto":
		if _, err := exec.LookPath("uv"); err == nil {
			return "uv"
		}
		return "pip"
	case "pip", "python":
		return "pip"
	default:
		return cfg.Runtime.Installer
	}
}

func installWithUV(cfg Config) error {
	if err := runCommand("uv", "venv", "--python", cfg.Runtime.Python, cfg.Runtime.Venv); err != nil {
		return err
	}
	args := append([]string{"pip", "install", "--python", venvPython(cfg)}, cfg.Runtime.Packages...)
	return runCommand("uv", args...)
}

func installWithPip(cfg Config) error {
	if err := runCommand(cfg.Runtime.Python, "-m", "venv", cfg.Runtime.Venv); err != nil {
		return err
	}
	pip := venvBinary(cfg, "pip")
	args := append([]string{"install", "--upgrade", "pip"}, cfg.Runtime.Packages...)
	return runCommand(pip, args...)
}

func Explain() {
	fmt.Println("Agent Sail load-test metrics")
	fmt.Println()
	fmt.Println("SLO metrics from llm-apps-monitoring-0424:")
	fmt.Println("  ttft_seconds: p95 time-to-first-token at API edge; pass < 1.5s")
	fmt.Println("  inter_token_latency_seconds: p95 SSE inter-token gap; pass < 0.08s")
	fmt.Println("  total_response_seconds: p95 request-to-last-token wall clock; pass < 10s")
	fmt.Println("  llm_errors_total / llm_requests_total: error rate; pass < 1%")
	fmt.Println("  container_memory_working_set_bytes: memory working set; pass < resources.memory.alert_at of resources.memory.limit")
	fmt.Println()
	fmt.Println("Leading autoscaling signals:")
	fmt.Println("  request_queue_depth: requests waiting on the concurrency semaphore")
	fmt.Println("  concurrent_llm_calls: in-flight LLM calls holding semaphore slots")
	fmt.Println("  concurrent_sessions: active SSE sessions")
	fmt.Println()
	fmt.Println("Human unit conversion:")
	fmt.Println("  Configure memory like Docker Compose: 512m, 1g, 2g")
	fmt.Println("  Agent Sail converts those values internally; users do not write byte values")
	fmt.Println()
	fmt.Println("Runtime:")
	fmt.Println("  agentsail loadtest run auto-installs Locust/httpx with uv into .agentsail/loadtests/.venv when missing")
	fmt.Println("  agentsail loadtest install pre-installs that runtime explicitly")
}

func ParsePercent(value string) (float64, error) {
	raw := strings.TrimSpace(value)
	raw = strings.TrimSuffix(raw, "%")
	if raw == "" {
		return 0, fmt.Errorf("empty percent")
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid percent %q", value)
	}
	if n < 0 {
		return 0, fmt.Errorf("percent must be non-negative")
	}
	if n > 1 {
		n = n / 100
	}
	return n, nil
}

func ParseMemoryBytes(value string) (uint64, error) {
	raw := strings.ToLower(strings.TrimSpace(value))
	if raw == "" {
		return 0, fmt.Errorf("empty memory value")
	}
	units := []struct {
		suffix string
		mult   float64
	}{
		{"gib", 1024 * 1024 * 1024},
		{"gb", 1000 * 1000 * 1000},
		{"g", 1000 * 1000 * 1000},
		{"mib", 1024 * 1024},
		{"mb", 1000 * 1000},
		{"m", 1000 * 1000},
		{"kib", 1024},
		{"kb", 1000},
		{"k", 1000},
	}
	mult := float64(1)
	number := raw
	for _, unit := range units {
		if strings.HasSuffix(raw, unit.suffix) {
			mult = unit.mult
			number = strings.TrimSpace(strings.TrimSuffix(raw, unit.suffix))
			break
		}
	}
	n, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory value %q", value)
	}
	if n < 0 {
		return 0, fmt.Errorf("memory must be non-negative")
	}
	return uint64(n * mult), nil
}

func ShellString(cmd *exec.Cmd) string {
	parts := append([]string{cmd.Path}, cmd.Args[1:]...)
	for i, part := range parts {
		if strings.ContainsAny(part, " \t\n\"'") {
			parts[i] = fmt.Sprintf("%q", part)
		}
	}
	return strings.Join(parts, " ")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func venvBinary(cfg Config, name string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(cfg.Runtime.Venv, "Scripts", name+".exe")
	}
	return filepath.Join(cfg.Runtime.Venv, "bin", name)
}

func venvPython(cfg Config) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(cfg.Runtime.Venv, "Scripts", "python.exe")
	}
	return filepath.Join(cfg.Runtime.Venv, "bin", "python")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func copyTemplateDir(srcRoot string, dstRoot string) error {
	entries, err := assets.Harness.ReadDir(srcRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		src := srcRoot + "/" + entry.Name()
		dst := filepath.Join(dstRoot, entry.Name())
		if entry.IsDir() {
			if err := copyTemplateDir(src, dst); err != nil {
				return err
			}
			continue
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(entry.Name(), ".sh") {
			mode = 0o755
		}
		if err := copyTemplateFile(src, dst, mode); err != nil {
			return err
		}
	}
	return nil
}

func copyTemplateFile(src string, dst string, mode os.FileMode) error {
	data, err := assets.Harness.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, mode)
}
