// Package safeshare produces the --safe-share disclosure of a report (plan
// §11): identifying values are replaced with consistent, report-local
// stand-ins while service types, topology, broad regions and change
// categories stay. The mapping lives only in memory and never reaches the
// artefact. The result is topology-oriented, not anonymous.
package safeshare

import (
	"fmt"
	"strings"

	"github.com/danushkastanley/tfviz/internal/provider/aws"
	"github.com/danushkastanley/tfviz/internal/report/model"
)

// descriptive values describe a kind of thing, not a specific thing, and
// are kept as they are.
var descriptive = map[string]bool{
	"engine": true, "engine_version": true, "instance_class": true, "instance_type": true, "instance_types": true,
	"broker_node_group_info.instance_type": true, "node_type": true, "kafka_version": true, "kafka_versions": true,
	"load_balancer_type": true, "target_type": true, "connectivity_type": true, "billing_mode": true,
	"launch_type": true, "package_type": true, "network_mode": true, "requires_compatibilities": true,
	"architectures": true, "protocol": true, "ports": true, "domain": true, "instance_tenancy": true,
	"key_usage": true, "encryption_info.client_broker": true, "type": true, "tier": true, "availability_zone": true,
	"ssl_policy": true, "version": true, "at_rest_encryption_enabled": true, "cpu": true, "memory": true,
	"runtime": true, "encrypted": true, "cluster_type": true, "vpc_endpoint_type": true,
}

// summarised values embed several identifiers in one string; they are
// withheld rather than replaced with a meaningless token.
var summarised = map[string]bool{"routes": true, "ingress": true, "egress": true}

// Apply returns a safe-share copy of the report. The input is not changed.
func Apply(r *model.Report) *model.Report {
	p := newPseudonyms()
	out := *r
	out.Disclosure = model.DisclosureSafeShare
	out.Title = "Infrastructure plan review (safe-share)"
	if r.Mode == model.ModeState {
		out.Title = "Recorded infrastructure state (safe-share)"
	}
	out.Source.ObjectVersion = ""

	labels := map[string]string{}
	out.Resources = make([]model.Resource, len(r.Resources))
	perType := map[string]int{}
	for i, res := range r.Resources {
		perType[res.Type]++
		noun := nounOf(res.Type)
		n := perType[res.Type]
		res.Label = fmt.Sprintf("%s %d", noun, n)
		res.Module = p.module(res.Module)
		// Neutral local names: a noun-based name could coincide with the original.
		res.Address = joinAddress(res.Module, string(res.Mode), res.Type, fmt.Sprintf("resource_%d", n))
		res.Change.PreviousAddress = ""
		res.Metadata = scrubMetadata(res.Metadata, p)
		labels[res.ID] = res.Label
		out.Resources[i] = res
	}
	out.Groups = make([]model.Group, len(r.Groups))
	for i, g := range r.Groups {
		g.Label = groupLabel(g, labels, p)
		out.Groups[i] = g
	}
	return &out
}

func scrubMetadata(fields []model.MetadataField, p *pseudonyms) []model.MetadataField {
	out := make([]model.MetadataField, len(fields))
	for i, f := range fields {
		f.Before = scrubValue(f.Key, f.Before, p)
		f.After = scrubValue(f.Key, f.After, p)
		out[i] = f
	}
	return out
}

func scrubValue(key string, v *model.FieldValue, p *pseudonyms) *model.FieldValue {
	if v == nil || v.Status() != model.StatusKnown {
		return v
	}
	var replaced model.FieldValue
	switch value := v.Value().(type) {
	case float64, bool:
		return v
	case string:
		switch {
		case descriptive[key]:
			return v
		case summarised[key]:
			replaced = model.Omitted()
		default:
			replaced = model.KnownString(p.value(value))
		}
	case []string:
		switch {
		case descriptive[key]:
			return v
		case summarised[key]:
			replaced = model.Omitted()
		default:
			items := make([]string, len(value))
			for j, s := range value {
				items[j] = p.value(s)
			}
			replaced = model.KnownList(items)
		}
	default:
		replaced = model.Omitted()
	}
	return &replaced
}

func groupLabel(g model.Group, labels map[string]string, p *pseudonyms) string {
	switch g.Kind {
	case model.GroupAccount:
		if account, ok := strings.CutPrefix(g.Label, "Account "); ok && account != "not recorded" {
			return "Account " + p.value(account)
		}
		return g.Label
	case model.GroupModule:
		if g.Label == "Root module" {
			return g.Label
		}
		return p.module(g.Label)
	case model.GroupVPC, model.GroupSubnet:
		if g.Resource == "" {
			if id, ok := strings.CutPrefix(g.Label, "VPC "); ok && strings.HasSuffix(id, " (outside this report)") {
				return "VPC " + p.value(strings.TrimSuffix(id, " (outside this report)")) + " (outside this report)"
			}
			return g.Label
		}
		label := labels[g.Resource]
		// Availability zones are broad location, like regions, and are kept.
		if _, az, ok := strings.Cut(g.Label, " · "); ok && g.Kind == model.GroupSubnet {
			label += " · " + az
		}
		return label
	default:
		// Regions, regional services and unplaced groups have generic labels.
		return g.Label
	}
}

func nounOf(resourceType string) string {
	if a, ok := aws.Lookup(resourceType); ok && a.Noun != "" {
		return a.Noun
	}
	noun := strings.ReplaceAll(strings.TrimPrefix(resourceType, "aws_"), "_", " ")
	if noun == "" {
		return "Resource"
	}
	return strings.ToUpper(noun[:1]) + noun[1:]
}

func joinAddress(module, mode, typ, name string) string {
	local := typ + "." + name
	if mode == string(model.ResourceData) {
		local = "data." + local
	}
	if module == "" {
		return local
	}
	return module + "." + local
}
