package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/testutil"
)

func TestPricingCLI(t *testing.T) {
	rates := testutil.RepoPath("testdata/pricing/synthetic-rates.json")
	plan := testutil.RepoPath("testdata/pricing/plan.json")
	for _, safe := range []bool{false, true} {
		output := filepath.Join(t.TempDir(), "report.html")
		args := []string{"plan", "--input", plan, "--output", output, "--pricing-file", rates, "--offline"}
		if safe {
			args = append(args, "--safe-share")
		}
		result := run(t, nil, args...)
		if result.code != ExitOK {
			t.Fatalf("%d: %s", result.code, result.stderr)
		}
		page, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		for name, b := range map[string][]byte{"html": page, "stdout": []byte(result.stdout), "stderr": []byte(result.stderr)} {
			testutil.AssertNoCanaries(t, name, b)
		}
		if !bytes.Contains(page, []byte(`"delta_monthly_cents":36`)) {
			t.Fatal("cost not rendered")
		}
		for _, secret := range []string{rates, plan, "hourly_price", "selectors"} {
			if bytes.Contains(page, []byte(secret)) {
				t.Fatalf("catalogue/input evidence leaked: %s", secret)
			}
		}
		if safe && bytes.Contains(page, []byte("private_example")) {
			t.Fatal("safe-share identity leak")
		}
	}
	// Exercise the actual shared CLI build path before/after safe-share, not a reimplementation.
	var stderr bytes.Buffer
	env := Env{Stdin: strings.NewReader(""), Stderr: &stderr, Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }, Version: "test"}
	opts := reportOptions{input: plan, pricingFile: rates, view: "architecture"}
	original, err := buildReport(input.KindPlan, opts, env)
	if err != nil {
		t.Fatal(err)
	}
	opts.safeShare = true
	shared, err := buildReport(input.KindPlan, opts, env)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(original.Pricing)
	b, _ := json.Marshal(shared.Pricing)
	if !bytes.Equal(a, b) {
		t.Fatal("safe share changed price")
	}
}

func TestPricingCLIFailures(t *testing.T) {
	for _, command := range []string{"state", "explore"} {
		result := run(t, nil, command, "--pricing-file", "private-path-do-not-echo")
		if result.code != ExitUnsupported {
			t.Fatalf("%s: %d", command, result.code)
		}
		if strings.Contains(result.stderr, "private-path-do-not-echo") {
			t.Fatal("path echoed")
		}
	}
	dir := t.TempDir()
	bad := filepath.Join(dir, "private-pricing-path.json")
	if err := os.WriteFile(bad, []byte(`{"secret":"CANARY-DB-PASSWORD-a41c"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{bad, filepath.Join(dir, "missing"), dir, "s3://private/rates"} {
		output := filepath.Join(dir, "never-written.html")
		result := run(t, nil, "plan", "--input", testutil.RepoPath("testdata/pricing/plan.json"), "--output", output, "--pricing-file", path)
		if result.code != ExitFailure {
			t.Fatalf("exit %d: %s", result.code, result.stderr)
		}
		testutil.AssertNoCanaries(t, "stderr", []byte(result.stderr))
		if strings.Contains(result.stderr, path) {
			t.Fatal("path echoed")
		}
		if _, err := os.Stat(output); !os.IsNotExist(err) {
			t.Fatal("output written on failure")
		}
	}
}
