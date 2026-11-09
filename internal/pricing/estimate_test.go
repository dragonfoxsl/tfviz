package pricing

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/report/model"
)

func attrs(region string) map[string]any {
	return map[string]any{"instance_type": "t3.micro", "ami": "ami-synthetic", "tenancy": "default", "region": region}
}

func resource(t *testing.T, actions []string, before, after any, extras map[string]any) input.Resource {
	t.Helper()
	change := map[string]any{"actions": actions, "before": before, "after": after}
	for k, v := range extras {
		change[k] = v
	}
	doc := map[string]any{"format_version": "1.2", "resource_changes": []any{map[string]any{
		"address": "aws_instance.test", "mode": "managed", "type": "aws_instance", "provider_name": "registry.terraform.io/hashicorp/aws", "change": change,
	}}}
	data, _ := json.Marshal(doc)
	snap, err := input.Read(data, input.KindPlan)
	if err != nil {
		t.Fatal(err)
	}
	return snap.Resources[0]
}

func catalogue(t *testing.T) *Catalogue {
	t.Helper()
	c, err := Load("../../testdata/pricing/synthetic-rates.json")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func amount(t *testing.T, p *int64, n int64) {
	t.Helper()
	if p == nil || *p != n {
		t.Fatalf("amount %v, want %d", p, n)
	}
}

func TestActionsAndSides(t *testing.T) {
	c := catalogue(t)
	for _, tt := range []struct {
		actions       []string
		before, after any
		b, a          int64
		overlap       bool
	}{
		{[]string{"create"}, nil, attrs("us-east-1"), 0, 37, false},
		{[]string{"delete"}, attrs("us-east-1"), nil, 37, 0, false},
		{[]string{"update"}, attrs("us-east-1"), attrs("us-west-2"), 37, 73, false},
		{[]string{"no-op"}, attrs("us-east-1"), attrs("us-east-1"), 37, 37, false},
		{[]string{"delete", "create"}, attrs("us-east-1"), attrs("us-west-2"), 37, 73, false},
		{[]string{"create", "delete"}, attrs("us-east-1"), attrs("us-west-2"), 37, 73, true},
	} {
		t.Run(strings.Join(tt.actions, ","), func(t *testing.T) {
			cost := c.Estimate(resource(t, tt.actions, tt.before, tt.after, nil))
			amount(t, cost.BeforeMonthlyCents, tt.b)
			amount(t, cost.AfterMonthlyCents, tt.a)
			amount(t, cost.DeltaMonthlyCents, tt.a-tt.b)
			if cost.ReplacementOverlap != tt.overlap {
				t.Fatal("overlap")
			}
			if tt.before != nil && cost.BeforeRegion != "us-east-1" {
				t.Fatal("before region")
			}
		})
	}
}

func TestUnavailableEvidence(t *testing.T) {
	c := catalogue(t)
	for _, key := range []string{"instance_type", "ami", "tenancy", "region"} {
		t.Run("missing "+key, func(t *testing.T) {
			after := attrs("us-east-1")
			delete(after, key)
			cost := c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), after, nil))
			if cost.AfterMonthlyCents != nil || cost.DeltaMonthlyCents != nil {
				t.Fatal("invented estimate")
			}
		})
		for _, mark := range []string{"after_unknown", "before_sensitive", "after_sensitive"} {
			t.Run(mark+key, func(t *testing.T) {
				cost := c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), attrs("us-east-1"), map[string]any{mark: map[string]any{key: true}}))
				if cost.AfterMonthlyCents != nil || cost.DeltaMonthlyCents != nil {
					t.Fatal("marked estimate")
				}
				if mark != "after_unknown" && cost.BeforeMonthlyCents != nil {
					t.Fatal("opposite sensitivity")
				}
			})
		}
	}
	for _, key := range []string{"arn", "instance_market_options", "spot_instance_request_id", "market_type"} {
		cost := c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), attrs("us-east-1"), map[string]any{"after_sensitive": map[string]any{key: true}}))
		if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil {
			t.Fatal("omitted sensitivity", key)
		}
		after := attrs("us-east-1")
		after[key] = "sensitive"
		cost = c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), after, map[string]any{"after_sensitive": map[string]any{key: true}}))
		if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil {
			t.Fatal("sensitivity", key)
		}
	}
	for _, modify := range []func(*input.Resource){
		func(r *input.Resource) { r.Actions = []string{"read"} },
		func(r *input.Resource) { r.Actions = []string{"forget"} },
		func(r *input.Resource) { r.Actions = []string{"unknown"} },
		func(r *input.Resource) { r.Actions = []string{"delete,create"} },
		func(r *input.Resource) { r.Importing = true },
		func(r *input.Resource) { r.Mode = "data" },
		func(r *input.Resource) { r.Type = "aws_s3_bucket" },
		func(r *input.Resource) { r.Provider = "other/aws" },
	} {
		r := resource(t, []string{"update"}, attrs("us-east-1"), attrs("us-east-1"), nil)
		modify(&r)
		cost := c.Estimate(r)
		if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil {
			t.Fatal("unsupported estimate")
		}
	}
	r := resource(t, []string{"create"}, nil, attrs("us-east-1"), nil)
	r.BeforeAbsent = false
	if c.Estimate(r).BeforeMonthlyCents != nil {
		t.Fatal("missing before is not confirmed absent")
	}
	r = resource(t, []string{"delete"}, attrs("us-east-1"), nil, nil)
	r.AfterAbsent = false
	if c.Estimate(r).AfterMonthlyCents != nil {
		t.Fatal("missing after is not confirmed absent")
	}
}

func TestRegionAndPurchaseEvidence(t *testing.T) {
	c := catalogue(t)
	for _, tt := range []struct {
		key   string
		value any
	}{
		{"tenancy", "dedicated"}, {"instance_market_options", []any{map[string]any{"market_type": "spot"}}},
		{"spot_instance_request_id", "sir-test"}, {"market_type", "spot"}, {"region", "unknown"},
		{"arn", "arn:aws:rds:us-east-1:123456789012:db:wrong-service"},
		{"arn", "arn:aws:ec2:us-west-2:123456789012:instance/i-test"},
	} {
		a := attrs("us-east-1")
		a[tt.key] = tt.value
		cost := c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), a, nil))
		if cost.AfterMonthlyCents != nil {
			t.Fatal("unsupported evidence", tt.key)
		}
	}
	a := attrs("us-east-1")
	delete(a, "region")
	a["arn"] = "arn:aws:ec2:us-east-1:123456789012:instance/i-test"
	cost := c.Estimate(resource(t, []string{"update"}, a, a, nil))
	amount(t, cost.AfterMonthlyCents, 37)
	delete(a, "arn")
	if c.Estimate(resource(t, []string{"update"}, a, a, nil)).AfterMonthlyCents != nil {
		t.Fatal("guessed region")
	}
}

func TestRDSAndSubtotal(t *testing.T) {
	c := catalogue(t)
	a := map[string]any{"region": "us-east-1", "instance_class": "db.t3.micro", "engine": "postgres", "multi_az": true, "license_model": "postgresql-license"}
	r := resource(t, []string{"update"}, a, a, nil)
	r.Type = "aws_db_instance"
	cost := c.Estimate(r)
	amount(t, cost.AfterMonthlyCents, 7300) // no Multi-AZ multiplier
	good := c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), attrs("us-west-2"), nil))
	bad := attrs("us-west-2")
	delete(bad, "ami")
	oneSide := c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), bad, nil))
	p := c.Summarize([]model.Resource{{Cost: good}, {Cost: oneSide}, {Cost: &model.CostEstimate{Reason: unavailable}}})
	if p.EstimatedResources != 1 || p.UnestimatedResources != 2 {
		t.Fatal("counts", p)
	}
	amount(t, p.Summary.BeforeMonthlyCents, 37)
	amount(t, p.Summary.AfterMonthlyCents, 73)
	amount(t, p.Summary.DeltaMonthlyCents, 36)
	p = c.Summarize([]model.Resource{{Cost: oneSide}})
	if p.Summary.BeforeMonthlyCents != nil || p.Summary.AfterMonthlyCents != nil || p.Summary.DeltaMonthlyCents != nil {
		t.Fatal("empty subtotal invented")
	}
}

func TestProducerFixtureMissingTenancyIsNotEstimated(t *testing.T) {
	for _, producer := range []string{"terraform-1.16-platform", "opentofu-1.13-platform"} {
		t.Run(producer, func(t *testing.T) {
			data, err := os.ReadFile("../../testdata/producer/" + producer + "/plan.json")
			if err != nil {
				t.Fatal(err)
			}
			snap, err := input.Read(data, input.KindPlan)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range snap.Resources {
				if r.Type != "aws_instance" || r.Mode != "managed" {
					continue
				}
				c := &Catalogue{currency: "USD", asOf: "2026-01-01", rates: map[string]int64{}}
				for _, side := range []input.Value{r.Before, r.After} {
					selectors := map[string]string{}
					for _, key := range fields[r.Type] {
						selectors[key], _ = side.Field(key).String()
					}
					// Synthetic rate, matched to actual producer-recorded selectors.
					c.rates[identity(r.Type, regionOf(side, r.Type), selectors)] = 100000
				}
				// These producer fixtures record tenancy as an empty string.
				// Even a matching test rate must not make that evidence known.
				for _, side := range []input.Value{r.Before, r.After} {
					tenancy, _ := side.Field("tenancy").String()
					if tenancy != "" {
						t.Fatal("fixture no longer exercises missing tenancy")
					}
				}
				cost := c.Estimate(r)
				if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil || cost.DeltaMonthlyCents != nil {
					t.Fatal("incomplete producer evidence was estimated")
				}
				return
			}
			t.Fatal("producer fixture has no managed EC2 instance")
		})
	}
}
