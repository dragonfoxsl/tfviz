package pricing

import (
	"math/big"
	"sort"

	"github.com/danushkastanley/tfviz/internal/input"
)

func (c *Catalogue) rateKey(typ string, v input.Value) (string, string, bool) {
	region := regionOf(v, typ)
	if region == "" {
		return "", "", false
	}
	selectors := map[string]string{}
	for _, key := range selectorKeys(typ) {
		s, ok := selector(v, key)
		if !ok {
			return "", region, false
		}
		selectors[key] = s
	}
	if !validSelectors(typ, selectors) || (typ == "aws_instance" && !onDemand(v)) {
		return "", region, false
	}
	return identity(typ, region, selectors), region, true
}

// comparableUsage picks the usage each live side prices. With a fixed charge a
// metric counts only if every live side can price it, so a delta never compares
// different component sets. Metrics left out are named from the fixed vocabulary.
func (c *Catalogue) comparableUsage(r input.Resource, action string) (map[string]map[string]int64, []string) {
	sides := map[string]input.Value{"before": r.Before, "after": r.After}
	if action == "create" {
		delete(sides, "before")
	}
	if action == "delete" {
		delete(sides, "after")
	}
	spec := services[r.Type]
	included := map[string]map[string]int64{"before": {}, "after": {}}
	var missing []string
	for _, comp := range spec.usage {
		priced := 0
		for which, v := range sides {
			_, supplied := c.usage[r.Address][which][comp.name]
			key, _, known := c.rateKey(r.Type, v)
			if _, rated := c.rates[key+"|"+comp.name]; supplied && known && rated {
				priced++
			}
		}
		if priced == len(sides) {
			for which := range sides {
				included[which][comp.name] = c.usage[r.Address][which][comp.name]
			}
			continue
		}
		if spec.base.name != "" || priced == 0 {
			missing = append(missing, comp.name)
		}
	}
	return included, missing
}

func (c *Catalogue) componentSide(typ string, v input.Value, usage map[string]int64) (*int64, string, bool) {
	key, region, ok := c.rateKey(typ, v)
	if !ok {
		return nil, region, false
	}
	spec := services[typ]
	total := new(big.Int)
	if spec.base.name != "" {
		q, ok := baseQuantity(v, spec.base)
		if !ok {
			return nil, region, false
		}
		rate, ok := c.rates[key+"|"+spec.base.name]
		if !ok {
			return nil, region, false
		}
		addProduct(total, rate, q)
	} else if len(usage) == 0 {
		return nil, region, false
	}
	names := make([]string, 0, len(usage))
	for name := range usage {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rate, ok := c.rates[key+"|"+name]
		if !ok {
			return nil, region, false
		}
		addProduct(total, rate, usage[name])
	}
	n, ok := cents(total)
	if !ok {
		return nil, region, false
	}
	return n, region, len(usage) > 0
}

func onDemand(v input.Value) bool {
	for _, key := range []string{"instance_market_options", "spot_instance_request_id", "market_type"} {
		value := v.Field(key)
		s, isString := value.String()
		empty := value.Kind() == input.KindNull || (key == "instance_market_options" && value.Kind() == input.KindList && len(value.Items()) == 0) || (key != "instance_market_options" && isString && s == "")
		if !empty {
			return false
		}
	}
	return true
}
