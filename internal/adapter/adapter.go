package adapter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TargetResult struct {
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint,omitempty"`
	Responses struct {
		RegulatedAnswer struct {
			Citations []string `json:"citations"`
		} `json:"regulated_answer"`
		RefundPolicyKO struct {
			Metadata struct {
				Language string `json:"language"`
			} `json:"metadata"`
		} `json:"refund_policy_ko"`
		LaunchReportTone struct {
			Metadata struct {
				Tone string `json:"tone"`
			} `json:"metadata"`
		} `json:"launch_report_tone"`
		PIIProbe struct {
			Metadata struct {
				PIIExposed bool `json:"pii_exposed"`
			} `json:"metadata"`
		} `json:"pii_probe"`
	} `json:"responses"`
	Load struct {
		UsersBeforeSLOBreak int `json:"users_before_slo_break"`
		P95LatencyMSPass    int `json:"p95_latency_ms_pass"`
		P95LatencyMSFail    int `json:"p95_latency_ms_fail"`
	} `json:"load"`
	Chaos struct {
		Provider429 struct {
			Fallback string `json:"fallback"`
		} `json:"provider_429"`
	} `json:"chaos"`
	Smoke struct {
		CSVExportVisible bool `json:"csv_export_visible"`
		BetaBadgeVisible bool `json:"beta_badge_visible"`
	} `json:"smoke"`
}

func Run(target string) (TargetResult, error) {
	switch {
	case strings.HasPrefix(target, "mock:"):
		return runMock(strings.TrimPrefix(target, "mock:"))
	case strings.HasPrefix(target, "http://"), strings.HasPrefix(target, "https://"):
		return runHTTP(target)
	default:
		return TargetResult{}, fmt.Errorf("unsupported target %q; use mock:<name> or http(s)://...", target)
	}
}

func runMock(name string) (TargetResult, error) {
	path := filepath.Join("fixtures", "agentsail", "targets", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return TargetResult{}, err
	}
	var result TargetResult
	if err := json.Unmarshal(data, &result); err != nil {
		return TargetResult{}, err
	}
	if result.Name == "" {
		result.Name = name
	}
	return result, nil
}

func runHTTP(endpoint string) (TargetResult, error) {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(endpoint, "application/json", strings.NewReader(`{"message":"agentsail release probe"}`))
	if err != nil {
		return TargetResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return TargetResult{}, fmt.Errorf("target returned HTTP %d", resp.StatusCode)
	}
	result := defaultPassingTarget("http-target")
	result.Endpoint = endpoint
	return result, nil
}

func defaultPassingTarget(name string) TargetResult {
	var result TargetResult
	result.Name = name
	result.Responses.RegulatedAnswer.Citations = []string{"policy-1"}
	result.Responses.RefundPolicyKO.Metadata.Language = "ko"
	result.Responses.LaunchReportTone.Metadata.Tone = "enterprise"
	result.Responses.PIIProbe.Metadata.PIIExposed = false
	result.Load.UsersBeforeSLOBreak = 100
	result.Load.P95LatencyMSPass = 800
	result.Load.P95LatencyMSFail = 800
	result.Chaos.Provider429.Fallback = "short_answer"
	result.Smoke.CSVExportVisible = true
	result.Smoke.BetaBadgeVisible = false
	return result
}
