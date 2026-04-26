package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDemoVerdicts(t *testing.T) {
	root := testRepoRoot(t)
	tmp := t.TempDir()
	if err := copyTree(filepath.Join(root, "fixtures"), filepath.Join(tmp, "fixtures")); err != nil {
		t.Fatal(err)
	}
	wd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	cases := []struct {
		customer string
		exitCode int
	}{
		{customer: "finbank", exitCode: 0},
		{customer: "retailco", exitCode: 0},
		{customer: "acme-bank", exitCode: 0},
	}
	for _, tc := range cases {
		code := Run([]string{"ci", "--customer", tc.customer, "--target", "mock:support_agent_v12", "--report", "--soft-exit"})
		if code != tc.exitCode {
			t.Fatalf("%s exit = %d, want %d", tc.customer, code, tc.exitCode)
		}
	}
	assertContains(t, filepath.Join(tmp, ".agentsail", "runs", "finbank-run-001.json"), `"verdict": "HOLD"`)
	assertContains(t, filepath.Join(tmp, ".agentsail", "runs", "retailco-run-001.json"), `"verdict": "SHIP"`)
	assertContains(t, filepath.Join(tmp, ".agentsail", "runs", "acme-bank-run-001.json"), `"verdict": "BLOCK"`)
	assertContains(t, filepath.Join(tmp, ".agentsail", "reports", "acme-bank-run-001.html"), "missing CSV export")
	assertContains(t, filepath.Join(tmp, ".agentsail", "reports", "acme-bank-run-001.html"), "beta badge exposed")
	assertContains(t, filepath.Join(tmp, ".agentsail", "reports", "acme-bank-run-001.html"), "tone drift")
}

func TestInitInstallsHarnessIntoEmptyProject(t *testing.T) {
	tmp := t.TempDir()
	code := Run([]string{"init", tmp})
	if code != 0 {
		t.Fatalf("init exit = %d, want 0", code)
	}
	for _, path := range []string{
		".claude-plugin/plugin.json",
		".claude/commands/agentsail/ci.md",
		".claude/commands/agentsail/init.md",
		".claude/commands/agentsail/loadtest.md",
		".claude/skills/agentsail-ci/SKILL.md",
		".claude/agents/agentsail/verdict-engine.md",
		".codex-plugin/plugin.json",
		"commands/agentsail.md",
		"commands/loadtest.md",
		"skills/agentsail/SKILL.md",
		".codex/commands/agentsail/init.md",
		".codex/commands/agentsail/loadtest.md",
		"fixtures/agentsail/targets/support_agent_v12.json",
		"agentsail.loadtest.yaml",
		"locust/agentsail/locustfile.py",
		"AGENTS.md",
		".gitignore",
	} {
		if _, err := os.Stat(filepath.Join(tmp, path)); err != nil {
			t.Fatalf("expected installed path %s: %v", path, err)
		}
	}
	assertContains(t, filepath.Join(tmp, "AGENTS.md"), "execute the terminal CLI directly")
	assertContains(t, filepath.Join(tmp, "commands", "agentsail.md"), "# /agentsail")
	assertContains(t, filepath.Join(tmp, "commands", "loadtest.md"), "# /agentsail:loadtest")
	assertContains(t, filepath.Join(tmp, "agentsail.loadtest.yaml"), "ttft_seconds")
	assertContains(t, filepath.Join(tmp, "agentsail.loadtest.yaml"), "request_queue_depth")
	assertContains(t, filepath.Join(tmp, "agentsail.loadtest.yaml"), "limit: 1g")
	assertContains(t, filepath.Join(tmp, "agentsail.loadtest.yaml"), "threshold: 80%")

	wd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	code = Run([]string{"ci", "--customer", "acme-bank", "--target", "mock:support_agent_v12", "--report", "--soft-exit"})
	if code != 0 {
		t.Fatalf("initialized project demo exit = %d, want 0", code)
	}
	assertContains(t, filepath.Join(tmp, ".agentsail", "runs", "acme-bank-run-001.json"), `"verdict": "BLOCK"`)

	code = Run([]string{"loadtest", "run", "--config", "agentsail.loadtest.yaml", "--dry-run"})
	if code != 0 {
		t.Fatalf("loadtest dry-run exit = %d, want 0", code)
	}
}

func TestLoadtestInitInstallsYamlAndLocustTemplate(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	code := Run([]string{"loadtest", "init", "--config", "agentsail.loadtest.yaml"})
	if code != 0 {
		t.Fatalf("loadtest init exit = %d, want 0", code)
	}
	assertContains(t, filepath.Join(tmp, "agentsail.loadtest.yaml"), "inter_token_latency_seconds")
	assertContains(t, filepath.Join(tmp, "locust", "agentsail", "locustfile.py"), "Agent Sail SSE-aware Locust")
	code = Run([]string{"loadtest", "run", "--config", "agentsail.loadtest.yaml", "--dry-run"})
	if code != 0 {
		t.Fatalf("loadtest dry-run exit = %d, want 0", code)
	}
}

func testRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "fixtures", "agentsail", "targets", "support_agent_v12.json")); err == nil {
			return wd
		}
		next := filepath.Dir(wd)
		if next == wd {
			t.Fatal("repo root not found")
		}
		wd = next
	}
}

func assertContains(t *testing.T, path string, needle string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), needle) {
		t.Fatalf("%s does not contain %q", path, needle)
	}
}

func copyTree(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
