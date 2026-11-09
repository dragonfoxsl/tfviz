package pricing

import (
	"encoding/json"
	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/report/model"
	"github.com/danushkastanley/tfviz/internal/testutil"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func expanded(t *testing.T, typ, selectors, component, unit string) *Catalogue {
	t.Helper()
	c, err := Read(strings.NewReader(`{"version":"2","currency":"USD","as_of":"2026-01-01","rates":[{"resource_type":"` + typ + `","region":"us-east-1","selectors":` + selectors + `,"component":"` + component + `","unit":"` + unit + `","unit_price":"0.1"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestExpandedBaseAndUsage(t *testing.T) {
	c := expanded(t, "aws_ebs_volume", `{"type":"gp3"}`, "storage_gb_month", "GB-month")
	a := map[string]any{"region": "us-east-1", "type": "gp3", "size": 20}
	r := resource(t, []string{"update"}, a, a, nil)
	r.Type = "aws_ebs_volume"
	cost := c.Estimate(r)
	amount(t, cost.AfterMonthlyCents, 200)
	if !cost.Partial {
		t.Fatal("base must be partial")
	}
	summary := c.Summarize([]model.Resource{{Cost: cost}})
	amount(t, summary.Summary.AfterMonthlyCents, 200)
	c = expanded(t, "aws_sqs_queue", `{"fifo_queue":"false"}`, "requests", "request")
	a = map[string]any{"region": "us-east-1", "fifo_queue": false}
	r = resource(t, []string{"update"}, a, a, nil)
	r.Type = "aws_sqs_queue"
	if c.Estimate(r).AfterMonthlyCents != nil {
		t.Fatal("missing usage is not free")
	}
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"before":{"requests":"0"},"after":{"requests":"10"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	cost = c.Estimate(r)
	amount(t, cost.BeforeMonthlyCents, 0)
	amount(t, cost.AfterMonthlyCents, 100)
	if !cost.UsageIncluded {
		t.Fatal("usage flag")
	}
	r = resource(t, []string{"update"}, a, a, map[string]any{"before_sensitive": map[string]any{"fifo_queue": true}})
	r.Type = "aws_sqs_queue"
	cost = c.Estimate(r)
	if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil || cost.AfterRegion != "" {
		t.Fatal("sensitive evidence leaked")
	}
}
func natCatalogue(t *testing.T, withUsageRate bool) (*Catalogue, input.Resource) {
	t.Helper()
	rates := `{"resource_type":"aws_nat_gateway","region":"us-east-1","selectors":{"connectivity_type":"public"},"component":"gateway_hours","unit":"hour","unit_price":"0.1"}`
	if withUsageRate {
		rates += `,{"resource_type":"aws_nat_gateway","region":"us-east-1","selectors":{"connectivity_type":"public"},"component":"processed_gb","unit":"GB","unit_price":"1"}`
	}
	c, err := Read(strings.NewReader(`{"version":"2","currency":"USD","as_of":"2026-01-01","rates":[` + rates + `]}`))
	if err != nil {
		t.Fatal(err)
	}
	a := map[string]any{"region": "us-east-1", "connectivity_type": "public"}
	r := resource(t, []string{"update"}, a, a, nil)
	r.Type = "aws_nat_gateway"
	return c, r
}

func TestFixedChargeShownWithoutUsage(t *testing.T) {
	missing := []string{"processed_gb"}
	c, r := natCatalogue(t, true)
	cost := c.Estimate(r)
	amount(t, cost.BeforeMonthlyCents, 7300)
	amount(t, cost.AfterMonthlyCents, 7300)
	if !reflect.DeepEqual(cost.MissingUsage, missing) || cost.UsageIncluded {
		t.Fatalf("no usage: %+v", cost)
	}
	if got := c.Summarize([]model.Resource{{Cost: cost}}).UsageOmittedResources; got != 1 {
		t.Fatalf("omitted count %d", got)
	}
	// One-sided usage must not turn into a delta between different component sets.
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"after":{"processed_gb":"10"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	cost = c.Estimate(r)
	amount(t, cost.AfterMonthlyCents, 7300)
	amount(t, cost.DeltaMonthlyCents, 0)
	if !reflect.DeepEqual(cost.MissingUsage, missing) || cost.UsageIncluded {
		t.Fatalf("one-sided usage: %+v", cost)
	}
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"before":{"processed_gb":"5"},"after":{"processed_gb":"10"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	cost = c.Estimate(r)
	amount(t, cost.BeforeMonthlyCents, 7800)
	amount(t, cost.AfterMonthlyCents, 8300)
	if len(cost.MissingUsage) != 0 || !cost.UsageIncluded {
		t.Fatalf("full usage: %+v", cost)
	}
	// Usage without a matching rate falls back to the fixed charge too.
	c, r = natCatalogue(t, false)
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"before":{"processed_gb":"5"},"after":{"processed_gb":"10"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	cost = c.Estimate(r)
	amount(t, cost.AfterMonthlyCents, 7300)
	if !reflect.DeepEqual(cost.MissingUsage, missing) {
		t.Fatalf("unrated usage: %+v", cost)
	}
}

func TestUsageWarningRespectsEvidence(t *testing.T) {
	c := expanded(t, "aws_sqs_queue", `{"fifo_queue":"false"}`, "requests", "request")
	a := map[string]any{"region": "us-east-1", "fifo_queue": false}
	r := resource(t, []string{"update"}, a, a, nil)
	r.Type = "aws_sqs_queue"
	cost := c.Estimate(r)
	if cost.AfterMonthlyCents != nil || cost.Reason != usageOnly || !reflect.DeepEqual(cost.MissingUsage, []string{"requests"}) {
		t.Fatalf("usage-only: %+v", cost)
	}
	// Sensitivity wins: no warning derived from a withheld estimate.
	r = resource(t, []string{"update"}, a, a, map[string]any{"before_sensitive": map[string]any{"fifo_queue": true}})
	r.Type = "aws_sqs_queue"
	if cost = c.Estimate(r); cost.MissingUsage != nil || cost.Reason != sensitive {
		t.Fatalf("sensitive: %+v", cost)
	}
	// A fixed charge with no rate has nothing to caveat.
	nat, r := natCatalogue(t, true)
	wrong := map[string]any{"region": "eu-west-1", "connectivity_type": "public"}
	r = resource(t, []string{"update"}, wrong, wrong, nil)
	r.Type = "aws_nat_gateway"
	if cost = nat.Estimate(r); cost.MissingUsage != nil {
		t.Fatalf("unpriced fixed charge: %+v", cost)
	}
}

func TestUsageVocabularyMatchesSchema(t *testing.T) {
	data, err := os.ReadFile(testutil.RepoPath("schema/report.v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Defs struct {
			CostEstimate struct {
				Properties struct {
					MissingUsage struct {
						MaxItems int `json:"maxItems"`
						Items    struct {
							Enum []string `json:"enum"`
						} `json:"items"`
					} `json:"missing_usage"`
				} `json:"properties"`
			} `json:"CostEstimate"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	missing := schema.Defs.CostEstimate.Properties.MissingUsage
	for typ, spec := range services {
		if len(spec.usage) > missing.MaxItems {
			t.Errorf("%s exceeds maxItems", typ)
		}
		for _, comp := range spec.usage {
			if !slices.Contains(missing.Items.Enum, comp.name) {
				t.Errorf("%s not in schema", comp.name)
			}
		}
	}
}

func TestUsageValidation(t *testing.T) {
	c := expanded(t, "aws_sqs_queue", `{"fifo_queue":"false"}`, "requests", "request")
	r := resource(t, []string{"update"}, attrs("us-east-1"), attrs("us-east-1"), nil)
	r.Type = "aws_sqs_queue"
	for _, doc := range []string{
		`{"version":"1","resources":null}`,
		`{"version":"1","resources":{"unknown":{"after":{"requests":"1"}}}}`,
		`{"version":"1","resources":{"aws_instance.test":{"after":{"requests":"-1"}}}}`,
		`{"version":"1","resources":{"aws_instance.test":{"after":{"requests":1}}}}`,
		`{"version":"1","resources":{"aws_instance.test":{"after":null}}}`,
		`{"version":"1","resources":{"aws_instance.test":{"after":{"requests":"1","requests":"2"}}}}`,
		`{"version":"1","resources":{}} {}`,
	} {
		if c.ReadUsage(strings.NewReader(doc), []input.Resource{r}) == nil {
			t.Fatalf("accepted invalid usage")
		}
	}
}
