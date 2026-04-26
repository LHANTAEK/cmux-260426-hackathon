package evidence

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func StateDir() string {
	return ".agentsail"
}

func Collect(customer string) error {
	src := filepath.Join("fixtures", "agentsail", customer)
	dst := filepath.Join(StateDir(), "cache", customer)
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("fixture context for %q not found: %w", customer, err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := copyFile(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func WriteJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func ReadJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func NextRunPath(customer string) string {
	dir := filepath.Join(StateDir(), "runs")
	_ = os.MkdirAll(dir, 0o755)
	entries, _ := os.ReadDir(dir)
	prefix := customer + "-run-"
	max := 0
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".json")
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(strings.TrimPrefix(name, prefix), "%03d", &n); err == nil && n > max {
			max = n
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s-run-%03d.json", customer, max+1))
}

func LatestRunPath(customer string) (string, error) {
	dir := filepath.Join(StateDir(), "runs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), customer+"-run-") && strings.HasSuffix(entry.Name(), ".json") {
			matches = append(matches, filepath.Join(dir, entry.Name()))
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no run artifact for %q", customer)
	}
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
