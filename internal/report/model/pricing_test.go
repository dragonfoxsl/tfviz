package model_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/pricing"
	"github.com/danushkastanley/tfviz/internal/report/build"
	"github.com/danushkastanley/tfviz/internal/report/model"
	"github.com/danushkastanley/tfviz/internal/report/safeshare"
	"github.com/danushkastanley/tfviz/internal/testutil"
)

func TestGeneratedPricingReportMatchesSchema(t *testing.T) {
	data, err := os.ReadFile(repoPath(t, "testdata/pricing/plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	snap, err := input.Read(data, input.KindPlan)
	if err != nil {
		t.Fatal(err)
	}
	c, err := pricing.Load(repoPath(t, "testdata/pricing/synthetic-rates.json"))
	if err != nil {
		t.Fatal(err)
	}
	schema := compileSchema(t)
	report := build.Report(snap, build.Options{Pricing: c, Source: model.SourceFile, GeneratedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), ToolVersion: "test"})
	if report.Pricing == nil || report.Pricing.EstimatedResources != 1 || report.Pricing.UnestimatedResources != 1 {
		t.Fatal("pricing missing or counts wrong")
	}
	if *report.Pricing.Summary.DeltaMonthlyCents != 36 {
		t.Fatal("subtotal")
	}
	for _, r := range []*model.Report{report, safeshare.Apply(report)} {
		out, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if err := validate(t, schema, out); err != nil {
			t.Fatal(err)
		}
		testutil.AssertNoCanaries(t, "pricing report", out)
	}
	without := build.Report(snap, build.Options{})
	if without.Pricing != nil || without.Resources[0].Cost != nil {
		t.Fatal("not opt in")
	}
	snap.Kind = input.KindState
	state := build.Report(snap, build.Options{Pricing: c})
	if state.Pricing != nil || state.Resources[0].Cost != nil {
		t.Fatal("state pricing")
	}
}
