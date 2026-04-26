package loadtest

import (
	"strings"
	"testing"
)

func TestParseMemoryBytes(t *testing.T) {
	tests := map[string]uint64{
		"512m": 512000000,
		"1g":   1000000000,
		"2gb":  2000000000,
		"1GiB": 1073741824,
	}
	for input, want := range tests {
		got, err := ParseMemoryBytes(input)
		if err != nil {
			t.Fatalf("ParseMemoryBytes(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseMemoryBytes(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestParsePercent(t *testing.T) {
	got, err := ParsePercent("80%")
	if err != nil {
		t.Fatal(err)
	}
	if got != 0.8 {
		t.Fatalf("ParsePercent(80%%) = %f, want 0.8", got)
	}
}

func TestMemorySummaryUsesGB(t *testing.T) {
	var cfg Config
	cfg.Resources.Memory.Limit = "1g"
	cfg.Resources.Memory.AlertAt = "80%"
	got := MemorySummary(cfg)
	if !strings.Contains(got, "1.00 GB") || !strings.Contains(got, "0.80 GB") {
		t.Fatalf("MemorySummary() = %q, want GB limit and alert values", got)
	}
}

func TestBoardStringShowsLoadtestTUIContent(t *testing.T) {
	var cfg Config
	cfg.Target.BaseURL = "http://localhost:8000"
	cfg.Target.ChatPath = "/chat"
	cfg.Resources.Memory.Limit = "1g"
	cfg.Resources.Memory.AlertAt = "80%"
	cfg.Locust.ResultsDir = ".agentsail/loadtests/default"
	cfg.Locust.Users = 32
	cfg.Locust.SpawnRate = 1
	cfg.Locust.RunTime = "12m"

	got := BoardString(cfg, "locust -f locust/agentsail/locustfile.py", "READY", 0, []string{"dry-run"})
	for _, want := range []string{
		"Agent Sail Load Test",
		"Status: READY",
		"Profile: 32 users",
		"Memory: 1g limit (1.00 GB), alert at 80% (0.80 GB)",
		"container_memory_working_set_bytes",
		".agentsail/loadtests/default/report.html",
		"dry-run",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("BoardString() missing %q:\n%s", want, got)
		}
	}
}

func TestRuntimeInstallerHonorsPipOverride(t *testing.T) {
	var cfg Config
	cfg.Runtime.Installer = "pip"
	if got := RuntimeInstaller(cfg); got != "pip" {
		t.Fatalf("RuntimeInstaller() = %q, want pip", got)
	}
}
