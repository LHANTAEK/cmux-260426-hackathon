package loadtest

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxRecentLogLines = 8

func RunTUI(cfg Config, dryRun bool, noInstall bool) error {
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
	command := ShellString(cmd)
	if dryRun {
		fmt.Print(BoardString(cfg, command, "READY", 0, []string{"dry-run: Locust was not started"}))
		return nil
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	lines := make(chan string, 64)
	done := make(chan error, 1)
	go scanLines(stdout, lines)
	go scanLines(stderr, lines)
	go func() {
		done <- cmd.Wait()
	}()

	startedAt := time.Now()
	recent := []string{"Locust started"}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	render := func(status string) {
		fmt.Print("\033[H\033[2J")
		fmt.Print(BoardString(cfg, command, status, time.Since(startedAt), recent))
	}
	render("RUNNING")

	for {
		select {
		case line := <-lines:
			line = strings.TrimSpace(line)
			if line != "" {
				recent = appendRecent(recent, line)
			}
		case <-ticker.C:
			render("RUNNING")
		case err := <-done:
			status := "PASS"
			if err != nil {
				status = "FAIL"
				recent = appendRecent(recent, err.Error())
			}
			render(status)
			return err
		}
	}
}

func BoardString(cfg Config, command string, status string, elapsed time.Duration, recent []string) string {
	htmlPath := filepath.Join(cfg.Locust.ResultsDir, "report.html")
	csvPrefix := filepath.Join(cfg.Locust.ResultsDir, "stats")
	var b strings.Builder
	fmt.Fprintf(&b, "Agent Sail Load Test\n")
	fmt.Fprintf(&b, "====================\n\n")
	fmt.Fprintf(&b, "Status: %s\n", status)
	if elapsed > 0 {
		fmt.Fprintf(&b, "Elapsed: %s\n", elapsed.Truncate(time.Second))
	}
	fmt.Fprintf(&b, "Target: %s%s\n", strings.TrimRight(cfg.Target.BaseURL, "/"), cfg.Target.ChatPath)
	fmt.Fprintf(&b, "Profile: %d users, spawn-rate %d/s, runtime %s\n", cfg.Locust.Users, cfg.Locust.SpawnRate, cfg.Locust.RunTime)
	fmt.Fprintf(&b, "Memory: %s\n\n", MemorySummary(cfg))

	fmt.Fprintf(&b, "SLOs\n")
	fmt.Fprintf(&b, "  ttft_seconds p95 < 1.5s\n")
	fmt.Fprintf(&b, "  inter_token_latency_seconds p95 < 0.08s\n")
	fmt.Fprintf(&b, "  total_response_seconds p95 < 10s\n")
	fmt.Fprintf(&b, "  llm_errors_total / llm_requests_total < 1%%\n")
	fmt.Fprintf(&b, "  container_memory_working_set_bytes < resources.memory.alert_at of resources.memory.limit\n\n")

	fmt.Fprintf(&b, "Leading Signals\n")
	fmt.Fprintf(&b, "  request_queue_depth, concurrent_llm_calls, concurrent_sessions\n\n")

	fmt.Fprintf(&b, "Artifacts\n")
	fmt.Fprintf(&b, "  HTML: %s\n", htmlPath)
	fmt.Fprintf(&b, "  CSV:  %s_*.csv\n\n", csvPrefix)

	fmt.Fprintf(&b, "Command\n")
	fmt.Fprintf(&b, "  %s\n\n", command)

	fmt.Fprintf(&b, "Recent Locust Output\n")
	if len(recent) == 0 {
		fmt.Fprintf(&b, "  waiting for output...\n")
	} else {
		for _, line := range recent {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}
	return b.String()
}

func MemorySummary(cfg Config) string {
	limitBytes, err := ParseMemoryBytes(cfg.Resources.Memory.Limit)
	if err != nil {
		return cfg.Resources.Memory.Limit + " limit, alert " + cfg.Resources.Memory.AlertAt
	}
	alertRatio, err := ParsePercent(cfg.Resources.Memory.AlertAt)
	if err != nil {
		return cfg.Resources.Memory.Limit + " limit, alert " + cfg.Resources.Memory.AlertAt
	}
	limitGB := float64(limitBytes) / 1000 / 1000 / 1000
	alertGB := limitGB * alertRatio
	return fmt.Sprintf("%s limit (%.2f GB), alert at %s (%.2f GB)", cfg.Resources.Memory.Limit, limitGB, cfg.Resources.Memory.AlertAt, alertGB)
}

func scanLines(reader io.Reader, lines chan<- string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
}

func appendRecent(recent []string, line string) []string {
	recent = append(recent, line)
	if len(recent) > maxRecentLogLines {
		return recent[len(recent)-maxRecentLogLines:]
	}
	return recent
}
