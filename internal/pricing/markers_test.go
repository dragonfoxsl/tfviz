package pricing

import "testing"

func TestSensitiveUnknownAndOmittedMarketEvidence(t *testing.T) {
	c := catalogue(t)
	for _, tt := range []struct {
		name               string
		actions            []string
		before             any
		sensitive, unknown any
	}{
		{"create sensitive unknown ARN", []string{"create"}, nil, map[string]any{"arn": true}, map[string]any{"arn": true}},
		{"create sensitive unknown AMI", []string{"create"}, nil, map[string]any{"ami": true}, map[string]any{"ami": true}},
		{"update empty sensitive market block", []string{"update"}, attrs("us-east-1"), map[string]any{"instance_market_options": []any{map[string]any{"market_type": true}}}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := attrs("us-east-1")
			a["instance_market_options"] = []any{}
			if before, ok := tt.before.(map[string]any); ok {
				before["instance_market_options"] = []any{}
			}
			cost := c.Estimate(resource(t, tt.actions, tt.before, a, map[string]any{"after_sensitive": tt.sensitive, "after_unknown": tt.unknown}))
			if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil || cost.BeforeRegion != "" || cost.AfterRegion != "" || cost.Reason != sensitive {
				t.Fatalf("sensitive evidence exported: %+v", cost)
			}
		})
	}
	a := attrs("us-east-1")
	a["instance_market_options"] = []any{}
	cost := c.Estimate(resource(t, []string{"update"}, attrs("us-east-1"), a, map[string]any{"after_unknown": map[string]any{"instance_market_options": []any{map[string]any{"market_type": true}}}}))
	if cost.AfterMonthlyCents != nil || cost.DeltaMonthlyCents != nil {
		t.Fatal("unknown market option estimated")
	}
}
