package pricing

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/danushkastanley/tfviz/internal/input"
)

func TestServiceComponents(t *testing.T) {
	cases := []struct {
		typ    string
		values map[string]any
		cents  int64
	}{
		{"aws_instance", attrs("us-east-1"), 7300},
		{"aws_db_instance", map[string]any{"instance_class": "db.test", "engine": "postgres", "multi_az": true, "license_model": "postgresql-license"}, 7300},
		{"aws_ebs_volume", map[string]any{"type": "gp3", "size": 20}, 200},
		{"aws_nat_gateway", map[string]any{"connectivity_type": "public"}, 7300},
		{"aws_lb", map[string]any{"load_balancer_type": "application"}, 7300},
		{"aws_eks_cluster", map[string]any{"version": "1.33", "upgrade_policy": []any{map[string]any{"support_type": "STANDARD"}}}, 7300},
		{"aws_kms_key", map[string]any{"key_usage": "ENCRYPT_DECRYPT", "customer_master_key_spec": "SYMMETRIC_DEFAULT", "multi_region": false}, 10},
		{"aws_secretsmanager_secret", map[string]any{}, 10},
		{"aws_elasticache_cluster", map[string]any{"node_type": "cache.test", "engine": "memcached", "engine_version": "1.6", "num_cache_nodes": 2}, 14600},
		{"aws_redshift_cluster", map[string]any{"node_type": "test", "cluster_type": "multi-node", "number_of_nodes": 2}, 14600},
		{"aws_vpc_endpoint", map[string]any{"vpc_endpoint_type": "Interface", "service_name": "com.amazonaws.us-east-1.test", "subnet_ids": []any{"subnet-a", "subnet-b"}}, 14600},
		{"aws_s3_bucket", map[string]any{}, 0},
		{"aws_lambda_function", map[string]any{"architectures": []any{"arm64"}, "package_type": "Zip", "memory_size": 128}, 0},
		{"aws_sqs_queue", map[string]any{"fifo_queue": false}, 0},
		{"aws_sns_topic", map[string]any{"fifo_topic": false}, 0},
		{"aws_cloudwatch_log_group", map[string]any{"log_group_class": "STANDARD"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.typ, func(t *testing.T) {
			tc.values["region"] = "us-east-1"
			r := resource(t, []string{"update"}, tc.values, tc.values, nil)
			r.Type = tc.typ
			spec := services[tc.typ]
			selectors := map[string]string{}
			for _, key := range spec.selectors {
				s, ok := selector(r.After, key)
				if !ok {
					t.Fatal("test selector", key)
				}
				selectors[key] = s
			}
			rates := []map[string]any{}
			for _, comp := range append([]component{spec.base}, spec.usage...) {
				if comp.name != "" {
					rates = append(rates, map[string]any{"resource_type": tc.typ, "region": "us-east-1", "selectors": selectors, "component": comp.name, "unit": comp.unit, "unit_price": "0.1"})
				}
			}
			data, _ := json.Marshal(map[string]any{"version": "2", "currency": "USD", "as_of": "2026-01-01", "rates": rates})
			c, err := Read(strings.NewReader(string(data)))
			if err != nil {
				t.Fatal(err)
			}
			if spec.base.name != "" {
				amount(t, c.Estimate(r).AfterMonthlyCents, tc.cents)
			} else if c.Estimate(r).AfterMonthlyCents != nil {
				t.Fatal("invented usage-free estimate")
			}
			// All supported usage metrics are independent monthly quantities.
			if len(spec.usage) > 0 {
				metrics := map[string]string{}
				for _, comp := range spec.usage {
					metrics[comp.name] = "10"
				}
				doc, _ := json.Marshal(map[string]any{"version": "1", "resources": map[string]any{r.Address: map[string]any{"before": metrics, "after": metrics}}})
				if err := c.ReadUsage(strings.NewReader(string(doc)), []input.Resource{r}); err != nil {
					t.Fatal(err)
				}
				cost := c.Estimate(r)
				amount(t, cost.AfterMonthlyCents, tc.cents+int64(len(spec.usage))*100)
				if !cost.UsageIncluded {
					t.Fatal("usage flag absent")
				}
			}
			// Every selector and base quantity must be known, and sensitivity wins on both sides.
			keys := append(append([]string{}, spec.selectors...), spec.base.quantity, "region", "arn")
			for _, key := range keys {
				if key == "" {
					continue
				}
				root := strings.SplitN(key, ".", 2)[0]
				if root == "architecture" {
					root = "architectures"
				}
				marked := resource(t, []string{"update"}, tc.values, tc.values, map[string]any{"before_sensitive": map[string]any{root: true}})
				marked.Type = tc.typ
				cost := c.Estimate(marked)
				if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil || cost.BeforeRegion != "" || cost.AfterRegion != "" {
					t.Fatal("sensitivity leak", key)
				}
				if key != "arn" {
					marked = resource(t, []string{"update"}, tc.values, tc.values, map[string]any{"after_unknown": map[string]any{root: true}})
					marked.Type = tc.typ
					if c.Estimate(marked).AfterMonthlyCents != nil {
						t.Fatal("unknown evidence priced", key)
					}
				}
			}
		})
	}
}

func TestUsageSideComparisonAndOverflow(t *testing.T) {
	c := expanded(t, "aws_sqs_queue", `{"fifo_queue":"false"}`, "requests", "request")
	a := map[string]any{"region": "us-east-1", "fifo_queue": false}
	r := resource(t, []string{"update"}, a, a, nil)
	r.Type = "aws_sqs_queue"
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"after":{"requests":"1"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	cost := c.Estimate(r)
	if cost.BeforeMonthlyCents != nil || cost.DeltaMonthlyCents != nil {
		t.Fatal("missing usage became zero or delta")
	}
	amount(t, cost.AfterMonthlyCents, 10)
	for _, value := range []string{"1000000000.000001", "1e3", "-1", "0.0000001"} {
		doc := `{"version":"1","resources":{"aws_instance.test":{"after":{"requests":"` + value + `"}}}}`
		if c.ReadUsage(strings.NewReader(doc), []input.Resource{r}) == nil {
			t.Fatal("invalid quantity accepted")
		}
	}
	// Valid but extremely large rate × quantity must fail closed at the JS-safe bound.
	c.rates[identity(r.Type, "us-east-1", map[string]string{"fifo_queue": "false"})+"|requests"] = 1000000000000000000
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"before":{"requests":"1000000000"},"after":{"requests":"1000000000"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	if c.Estimate(r).AfterMonthlyCents != nil {
		t.Fatal("unsafe amount exported")
	}
}
