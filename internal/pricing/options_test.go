package pricing

import (
	"testing"
)

func TestKnownEmptyMarketIdentifiers(t *testing.T) {
	c := catalogue(t)
	a := attrs("us-east-1")
	a["instance_market_options"] = []any{}
	a["spot_instance_request_id"] = ""
	a["market_type"] = ""
	cost := c.Estimate(resource(t, []string{"update"}, a, a, nil))
	amount(t, cost.AfterMonthlyCents, 37)
}
