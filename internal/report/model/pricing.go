package model

// Pricing is an explicitly partial, comparable-resource steady-state subtotal.
type Pricing struct {
	UsageProvided        bool         `json:"usage_provided,omitempty"`
	Currency             string       `json:"currency"`
	AsOf                 string       `json:"as_of"`
	HoursPerMonth        int          `json:"hours_per_month"`
	Summary              CostEstimate `json:"summary"`
	EstimatedResources   int          `json:"estimated_resources"`
	UnestimatedResources int          `json:"unestimated_resources"`
	// UsageOmittedResources counts resources whose MissingUsage is not empty.
	UsageOmittedResources int `json:"usage_omitted_resources,omitempty"`
}

// Pointer amounts preserve known zero, distinct from an unavailable estimate.
type CostEstimate struct {
	Partial       bool `json:"partial,omitempty"`
	UsageIncluded bool `json:"usage_included,omitempty"`
	// MissingUsage names metrics, from a fixed vocabulary, left out of the amounts.
	MissingUsage       []string `json:"missing_usage,omitempty"`
	BeforeMonthlyCents *int64   `json:"before_monthly_cents,omitempty"`
	AfterMonthlyCents  *int64   `json:"after_monthly_cents,omitempty"`
	DeltaMonthlyCents  *int64   `json:"delta_monthly_cents,omitempty"`
	BeforeRegion       string   `json:"before_region,omitempty"`
	AfterRegion        string   `json:"after_region,omitempty"`
	Reason             string   `json:"reason,omitempty"`
	ReplacementOverlap bool     `json:"replacement_overlap,omitempty"`
}
