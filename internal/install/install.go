package install

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/braincrew/agentsail/internal/assets"
)

type Summary struct {
	ProjectRoot string
	Files       int
	Dirs        int
}

func Init(projectRoot string) (Summary, error) {
	if projectRoot == "" {
		projectRoot = "."
	}
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return Summary{}, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return Summary{}, err
	}
	summary := Summary{ProjectRoot: abs}
	if err := copyEmbeddedDir("templates/claude", filepath.Join(abs, ".claude"), &summary); err != nil {
		return Summary{}, err
	}
	if err := copyEmbeddedDir("templates/claude-plugin", filepath.Join(abs, ".claude-plugin"), &summary); err != nil {
		return Summary{}, err
	}
	if err := copyEmbeddedDir("templates/codex", filepath.Join(abs, ".codex"), &summary); err != nil {
		return Summary{}, err
	}
	if err := copyEmbeddedDir("templates/codex-plugin", abs, &summary); err != nil {
		return Summary{}, err
	}
	if err := copyEmbeddedDir("templates/fixtures", filepath.Join(abs, "fixtures"), &summary); err != nil {
		return Summary{}, err
	}
	if err := copyEmbeddedDir("templates/locust", filepath.Join(abs, "locust"), &summary); err != nil {
		return Summary{}, err
	}
	if err := copyEmbeddedFile("templates/agentsail.loadtest.yaml", filepath.Join(abs, "agentsail.loadtest.yaml"), 0o644, &summary); err != nil {
		return Summary{}, err
	}
	for _, dir := range []string{
		filepath.Join(abs, ".agentsail", "cache"),
		filepath.Join(abs, ".agentsail", "contracts"),
		filepath.Join(abs, ".agentsail", "runs"),
		filepath.Join(abs, ".agentsail", "reports"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Summary{}, err
		}
		summary.Dirs++
	}
	if err := ensureGitignore(abs); err != nil {
		return Summary{}, err
	}
	if err := ensureAgentsMarkdown(abs); err != nil {
		return Summary{}, err
	}
	return summary, nil
}

func copyEmbeddedDir(srcRoot string, dstRoot string, summary *Summary) error {
	return fs.WalkDir(assets.Harness, srcRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dstRoot, 0o755)
		}
		dst := filepath.Join(dstRoot, rel)
		if entry.IsDir() {
			summary.Dirs++
			return os.MkdirAll(dst, 0o755)
		}
		data, err := assets.Harness.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		mode := fs.FileMode(0o644)
		if strings.HasSuffix(path, ".sh") {
			mode = 0o755
		}
		summary.Files++
		return os.WriteFile(dst, data, mode)
	})
}

func ensureGitignore(projectRoot string) error {
	path := filepath.Join(projectRoot, ".gitignore")
	block := []string{
		"",
		"# Agent Sail local artifacts",
		".agentsail/",
		"bin/agentsail",
		"fixtures/agentsail/",
		".codex/commands/agentsail/",
		".codex-plugin/",
		"commands/agentsail*.md",
		"skills/agentsail/",
		"locust/agentsail/",
		"agentsail.loadtest.yaml",
		".gocache/",
		".gomodcache/",
	}
	existing := map[string]bool{}
	if file, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			existing[scanner.Text()] = true
		}
		file.Close()
	}
	var missing []string
	for _, line := range block {
		if line == "" || !existing[line] {
			missing = append(missing, line)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(missing, "\n") + "\n")
	return err
}

func ensureAgentsMarkdown(projectRoot string) error {
	path := filepath.Join(projectRoot, "AGENTS.md")
	const marker = "<!-- agentsail:init -->"
	block := marker + `

## Agent Sail

Agent Sail is installed for this project.

- Terminal CLI: ` + "`agentsail`" + `
- Claude Code commands: ` + "`/agentsail:init`, `/agentsail:collect`, `/agentsail:compile`, `/agentsail:check`, `/agentsail:verdict`, `/agentsail:ci`, `/agentsail:report`, `/agentsail:loadtest`, `/agentsail:doctor`, `/agentsail:version`" + `
- Codex plugin commands: ` + "`/agentsail`, `/agentsail:init`, `/agentsail:ci`, `/agentsail:loadtest`" + `
- Codex skill: ` + "`skills/agentsail/SKILL.md`" + `
- State/evidence directory: ` + "`.agentsail/`" + `
- Load test config: ` + "`agentsail.loadtest.yaml`" + `
- Primary demo:
  ` + "`agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert`" + `
- Load test dry run:
  ` + "`agentsail loadtest run --config agentsail.loadtest.yaml --dry-run`" + `

Codex should use the Agent Sail plugin command docs and skill, then execute the terminal CLI directly.
`
	data, err := os.ReadFile(path)
	if err == nil && strings.Contains(string(data), marker) {
		return nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}
	_, err = file.WriteString(block)
	return err
}

func InstallDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "agentsail")
	}
	return filepath.Join(os.Getenv("HOME"), ".local", "bin")
}

func PrintSummary(summary Summary) {
	fmt.Printf("Agent Sail initialized in %s\n", summary.ProjectRoot)
	fmt.Printf("Installed %d harness files and prepared %d directories\n", summary.Files, summary.Dirs)
	fmt.Println("Next:")
	fmt.Println("  agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert")
	fmt.Println("  agentsail loadtest run --config agentsail.loadtest.yaml --dry-run")
	fmt.Println("  Claude Code: /agentsail:init, /agentsail:ci, /agentsail:loadtest")
	fmt.Println("  Codex: /agentsail, /agentsail:init, /agentsail:ci, /agentsail:loadtest")
}

func copyEmbeddedFile(src string, dst string, mode fs.FileMode, summary *Summary) error {
	data, err := assets.Harness.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	summary.Files++
	return os.WriteFile(dst, data, mode)
}
