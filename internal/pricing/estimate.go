package pricing

import (
	"strings"

	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/report/model"
)

const unavailable = "Estimate unavailable: unsupported action, resource, purchase option, or incomplete pricing evidence."
const sensitive = "Estimate withheld because pricing evidence is sensitive."
const partial = "Partial steady-state subtotal of comparable resources only; excludes unestimated resources and unprovided usage."
const resourcePartial = "Partial steady-state estimate of supplied components only; excludes unprovided usage, other charges and transient replacement costs."
const usageOnly = "Estimate unavailable: this resource is billed only by usage and no usage was supplied."

// Estimate reads only sensitivity-safe input.Value trees, never raw attributes.
func (c *Catalogue) Estimate(r input.Resource) *model.CostEstimate {
	out := &model.CostEstimate{Reason: unavailable, Partial: true}
	keys, supported := fields[r.Type]
	if c.version == "2" {
		_, supported = services[r.Type]
		keys = selectorKeys(r.Type)
	}
	if !supported || r.Mode != "managed" || r.Provider != "hashicorp/aws" || r.Importing {
		return out
	}
	relevant := append(append([]string{}, keys...), "region", "arn")
	if r.Type == "aws_instance" {
		relevant = append(relevant, "instance_market_options", "spot_instance_request_id", "market_type")
	}
	if c.version == "2" && services[r.Type].base.quantity != "" {
		relevant = append(relevant, services[r.Type].base.quantity)
	}
	for _, key := range relevant {
		root := strings.SplitN(key, ".", 2)[0]
		if evidence(r.Before, root).Contains(input.KindSensitive) || evidence(r.After, root).Contains(input.KindSensitive) {
			out.Reason = sensitive
			return out
		}
	}
	if len(r.Actions) == 0 || len(r.Actions) > 2 {
		return out
	}
	for _, a := range r.Actions {
		if strings.Contains(a, ",") {
			return out
		}
	}
	action := strings.Join(r.Actions, ",")
	included, missing := c.comparableUsage(r, action)
	fixed := services[r.Type].base.name != ""
	priced := false
	side := func(v input.Value, which string) (*int64, string) {
		if c.version != "2" {
			return c.side(r.Type, v)
		}
		usage := included[which]
		if !fixed {
			// Nothing to fall back to: an omitted monthly quantity is not zero.
			for _, metrics := range c.usage[r.Address] {
				for name := range metrics {
					if _, ok := c.usage[r.Address][which][name]; !ok {
						return nil, regionOf(v, r.Type)
					}
				}
			}
			usage = c.usage[r.Address][which]
		}
		amount, region, used := c.componentSide(r.Type, v, usage)
		out.UsageIncluded = out.UsageIncluded || used
		priced = priced || amount != nil
		return amount, region
	}
	switch action {
	case "create":
		if !r.BeforeAbsent || r.HasBefore || !r.HasAfter {
			return out
		}
		out.BeforeMonthlyCents = ptr(0)
		out.AfterMonthlyCents, out.AfterRegion = side(r.After, "after")
	case "delete":
		if !r.AfterAbsent || r.HasAfter || !r.HasBefore {
			return out
		}
		out.BeforeMonthlyCents, out.BeforeRegion = side(r.Before, "before")
		out.AfterMonthlyCents = ptr(0)
	case "update", "no-op", "delete,create", "create,delete":
		if !r.HasBefore || !r.HasAfter {
			return out
		}
		out.BeforeMonthlyCents, out.BeforeRegion = side(r.Before, "before")
		out.AfterMonthlyCents, out.AfterRegion = side(r.After, "after")
		out.ReplacementOverlap = action == "create,delete"
	default:
		return out
	}
	if !fixed || priced {
		out.MissingUsage = missing
	}
	if !fixed && len(c.usage[r.Address]) == 0 {
		out.Reason = usageOnly
	}
	if out.BeforeMonthlyCents != nil && out.AfterMonthlyCents != nil {
		out.DeltaMonthlyCents = ptr(*out.AfterMonthlyCents - *out.BeforeMonthlyCents)
		out.Reason = resourcePartial
	}
	return out
}

func (c *Catalogue) side(typ string, v input.Value) (*int64, string) {
	region := regionOf(v, typ)
	if region == "" {
		return nil, ""
	}
	selectors := map[string]string{}
	for _, key := range fields[typ] {
		if key == "multi_az" {
			b, ok := v.Field(key).Bool()
			if !ok {
				return nil, region
			}
			if b {
				selectors[key] = "true"
			} else {
				selectors[key] = "false"
			}
		} else {
			s, ok := v.Field(key).String()
			if !ok || s == "" {
				return nil, region
			}
			selectors[key] = s
		}
	}
	if typ == "aws_instance" && (selectors["tenancy"] != "default" || !onDemand(v)) {
		return nil, region
	}
	micro, ok := c.rates[identity(typ, region, selectors)]
	if !ok {
		return nil, region
	}
	amount, ok := monthly(micro)
	if !ok {
		return nil, region
	}
	return ptr(amount), region
}

func regionOf(v input.Value, typ string) string {
	explicit := v.Field("region")
	region, _ := explicit.String()
	if explicit.Kind() != input.KindNull && (!regionPattern.MatchString(region) || len(region) > 64) {
		return ""
	}
	arn := v.Field("arn")
	if arn.Kind() == input.KindSensitive {
		return ""
	}
	if arn.Kind() == input.KindUnknown {
		return region
	}
	if arn.Kind() != input.KindNull && arn.Kind() != input.KindString {
		return ""
	}
	if s, ok := arn.String(); ok && s != "" {
		parts := strings.SplitN(s, ":", 6)
		service, prefix := services[typ].arnService, services[typ].arnPrefix
		if typ == "aws_s3_bucket" {
			// Bucket ARNs are global; they cannot supply regional price evidence.
			if len(parts) != 6 || parts[0] != "arn" || (parts[1] != "aws" && parts[1] != "aws-cn" && parts[1] != "aws-us-gov") || parts[2] != "s3" || parts[3] != "" || parts[4] != "" || parts[5] == "" {
				return ""
			}
			return region
		}
		if len(parts) != 6 || parts[0] != "arn" || (parts[1] != "aws" && parts[1] != "aws-cn" && parts[1] != "aws-us-gov") || parts[2] != service || !regionPattern.MatchString(parts[3]) || len(parts[3]) > 64 || len(parts[4]) != 12 || strings.Trim(parts[4], "0123456789") != "" || !strings.HasPrefix(parts[5], prefix) || len(parts[5]) <= len(prefix) {
			return ""
		}
		if region != "" && region != parts[3] {
			return ""
		}
		region = parts[3]
	}
	return region
}

func ptr(n int64) *int64 { return &n }

// Summarize never adds one-sided costs to the comparable subtotal.
func (c *Catalogue) Summarize(resources []model.Resource) *model.Pricing {
	p := &model.Pricing{UsageProvided: c.usage != nil, Currency: c.currency, AsOf: c.asOf, HoursPerMonth: 730, Summary: model.CostEstimate{Reason: partial, Partial: true}}
	var before, after int64
	for _, r := range resources {
		cost := r.Cost
		if cost != nil && len(cost.MissingUsage) > 0 {
			p.UsageOmittedResources++
		}
		if cost == nil || cost.BeforeMonthlyCents == nil || cost.AfterMonthlyCents == nil {
			p.UnestimatedResources++
			continue
		}
		if before > maxSafe-*cost.BeforeMonthlyCents || after > maxSafe-*cost.AfterMonthlyCents {
			// Fail closed if a future resource limit exceeds the current safe bound.
			p.EstimatedResources = 0
			p.UnestimatedResources = len(resources)
			return p
		}
		p.Summary.UsageIncluded = p.Summary.UsageIncluded || cost.UsageIncluded
		before += *cost.BeforeMonthlyCents
		after += *cost.AfterMonthlyCents
		p.EstimatedResources++
	}
	if p.EstimatedResources > 0 {
		p.Summary.BeforeMonthlyCents, p.Summary.AfterMonthlyCents = ptr(before), ptr(after)
		p.Summary.DeltaMonthlyCents = ptr(after - before)
	}
	return p
}
