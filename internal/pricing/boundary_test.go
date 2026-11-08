package pricing

import (
	"strings"
	"testing"

	"github.com/danushkastanley/tfviz/internal/input"
)

func TestSmallV2UnitPrices(t *testing.T) {
	// Synthetic sub-micro-unit rates must survive multiplication, not round to zero.
	for _, tc := range []struct {
		rate, quantity string
		want           int64
	}{
		{"0.0000004", "1000000", 40},
		{"0.000000000001", "1000000000", 0},
		{"0.000000005", "1000000", 1},
	} {
		t.Run(tc.rate, func(t *testing.T) {
			c, err := Read(strings.NewReader(`{"version":"2","currency":"USD","as_of":"2026-01-01","rates":[{"resource_type":"aws_sqs_queue","region":"us-east-1","selectors":{"fifo_queue":"false"},"component":"requests","unit":"request","unit_price":"` + tc.rate + `"}]}`))
			if err != nil {
				t.Fatal(err)
			}
			a := map[string]any{"region": "us-east-1", "fifo_queue": false}
			r := resource(t, []string{"update"}, a, a, nil)
			r.Type = "aws_sqs_queue"
			doc := `{"version":"1","resources":{"aws_instance.test":{"before":{"requests":"0"},"after":{"requests":"` + tc.quantity + `"}}}}`
			if err := c.ReadUsage(strings.NewReader(doc), []input.Resource{r}); err != nil {
				t.Fatal(err)
			}
			amount(t, c.Estimate(r).AfterMonthlyCents, tc.want)
		})
	}
}

func TestNestedContainerSensitivityWithholdsBothSides(t *testing.T) {
	c := expanded(t, "aws_eks_cluster", `{"version":"1.33","upgrade_policy.support_type":"STANDARD"}`, "cluster_hours", "hour")
	before := map[string]any{"region": "us-east-1", "version": "1.33", "upgrade_policy": []any{map[string]any{"support_type": "STANDARD"}, map[string]any{"support_type": "STANDARD"}}}
	after := map[string]any{"region": "us-east-1", "version": "1.33", "upgrade_policy": []any{map[string]any{"support_type": "STANDARD"}}}
	r := resource(t, []string{"update"}, before, after, map[string]any{"before_sensitive": map[string]any{"upgrade_policy": []any{map[string]any{"support_type": true}, map[string]any{}}}})
	r.Type = "aws_eks_cluster"
	cost := c.Estimate(r)
	if cost.BeforeMonthlyCents != nil || cost.AfterMonthlyCents != nil || cost.BeforeRegion != "" || cost.AfterRegion != "" {
		t.Fatal("sensitive container leaked derived evidence")
	}
}

func TestUsageRejectsIneligibleResourcesAndV1(t *testing.T) {
	a := map[string]any{"region": "us-east-1", "fifo_queue": false}
	r := resource(t, []string{"update"}, a, a, nil)
	r.Type = "aws_sqs_queue"
	doc := `{"version":"1","resources":{"aws_instance.test":{"before":{"requests":"1"},"after":{"requests":"2"}}}}`
	for _, mutate := range []func(*input.Resource){
		func(r *input.Resource) { r.Mode = "data" },
		func(r *input.Resource) { r.Provider = "other/aws" },
		func(r *input.Resource) { r.Importing = true },
	} {
		c := expanded(t, "aws_sqs_queue", `{"fifo_queue":"false"}`, "requests", "request")
		bad := r
		mutate(&bad)
		if c.ReadUsage(strings.NewReader(doc), []input.Resource{bad}) == nil {
			t.Error("accepted unusable resource usage")
		}
	}
	c := catalogue(t)
	if c.ReadUsage(strings.NewReader(doc), []input.Resource{r}) == nil {
		t.Error("v1 catalogue silently accepts v2 usage")
	}
	c = expanded(t, "aws_sqs_queue", `{"fifo_queue":"false"}`, "requests", "request")
	r.Type = "aws_unsupported"
	if c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"after":{}}}}`), []input.Resource{r}) == nil {
		t.Error("unsupported empty usage accepted")
	}
}

func TestExpandedARNEvidence(t *testing.T) {
	c := expanded(t, "aws_sqs_queue", `{"fifo_queue":"false"}`, "requests", "request")
	a := map[string]any{"region": "us-east-1", "fifo_queue": false, "arn": true}
	r := resource(t, []string{"create"}, nil, a, nil)
	r.Type = "aws_sqs_queue"
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"after":{"requests":"1"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	if c.Estimate(r).AfterMonthlyCents != nil {
		t.Error("malformed ARN was treated as absent")
	}
	c = expanded(t, "aws_s3_bucket", `{}`, "get_requests", "request")
	a = map[string]any{"region": "us-east-1", "arn": "arn:aws-us-gov:s3:::synthetic-bucket"}
	r = resource(t, []string{"create"}, nil, a, nil)
	r.Type = "aws_s3_bucket"
	if err := c.ReadUsage(strings.NewReader(`{"version":"1","resources":{"aws_instance.test":{"after":{"get_requests":"1"}}}}`), []input.Resource{r}); err != nil {
		t.Fatal(err)
	}
	amount(t, c.Estimate(r).AfterMonthlyCents, 10)
}
