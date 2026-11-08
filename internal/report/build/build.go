// Package build assembles the sanitised report model from a snapshot.
package build

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/danushkastanley/tfviz/internal/graph"
	"github.com/danushkastanley/tfviz/internal/input"
	"github.com/danushkastanley/tfviz/internal/pricing"
	"github.com/danushkastanley/tfviz/internal/projection"
	"github.com/danushkastanley/tfviz/internal/report/model"
)

// Options are the caller-supplied report details. Time is injected so
// output is reproducible.
type Options struct {
	Title string
	// View is the structure the interface opens in; empty means architecture.
	View        model.View
	Source      model.SourceKind
	GeneratedAt time.Time
	ToolVersion string
	// SourceTime and SourceVersion describe the retrieved object (for
	// example an S3 object's last modification), when known.
	SourceTime    *time.Time
	SourceVersion string
	Pricing       *pricing.Catalogue
}

// Report builds the report model. It never copies raw attribute values:
// metadata comes only from approved fields via projection.
func Report(snap *input.Snapshot, opts Options) *model.Report {
	g := graph.Build(snap)
	report := &model.Report{
		SchemaVersion: model.SchemaVersion,
		Mode:          model.ModeState,
		Disclosure:    model.DisclosureInternal,
		Title:         title(opts.Title, snap),
		InitialView:   opts.View,
		GeneratedAt:   opts.GeneratedAt.UTC(),
		Tool:          model.Tool{Name: "tfviz", Version: opts.ToolVersion},
		Producer:      producer(snap),
		Source:        model.Source{Kind: opts.Source, TimestampStatus: model.TimestampUnavailable},
		Completeness:  completeness(snap),
		Relationships: g.Relationships,
		Unresolved:    g.Unresolved,
		Groups:        g.Groups,
		Resources:     make([]model.Resource, 0, len(g.Nodes)),
	}
	if g.Plan {
		report.Mode = model.ModePlan
	}
	// A plan's own timestamp says when it was made; otherwise use the
	// retrieved object's time. Neither describes the live infrastructure.
	for _, ts := range []*time.Time{snap.Timestamp, opts.SourceTime} {
		if ts != nil {
			utc := ts.UTC()
			report.Source.Timestamp = &utc
			report.Source.TimestampStatus = model.TimestampKnown
			break
		}
	}
	report.Source.ObjectVersion = clip(opts.SourceVersion, 1024)
	for _, n := range g.Nodes {
		res := resource(n, g)
		if g.Plan && opts.Pricing != nil {
			res.Cost = opts.Pricing.Estimate(n.Resource)
		}
		report.Resources = append(report.Resources, res)
		count(&report.Summary, n.Change.Action)
		report.Coverage.ResourcesTotal++
		if n.Supported {
			report.Coverage.ResourcesSupported++
		} else {
			report.Coverage.ResourcesGeneric++
		}
	}
	report.Warnings = warnings(snap, g)
	if g.Plan && opts.Pricing != nil {
		report.Pricing = opts.Pricing.Summarize(report.Resources)
	}
	// Collections are always arrays, never null, even when empty.
	report.Relationships = nonNil(report.Relationships)
	report.Unresolved = nonNil(report.Unresolved)
	report.Groups = nonNil(report.Groups)
	return report
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func resource(n *graph.Node, g *graph.Graph) model.Resource {
	r := model.Resource{
		ID:       n.ID,
		Address:  clip(n.Resource.Address, 1024),
		Label:    n.Label,
		Type:     n.Resource.Type,
		Family:   model.FamilyOther,
		Role:     model.RoleEntity,
		Mode:     model.ResourceMode(n.Resource.Mode),
		Module:   clip(n.Resource.Module, 1024),
		Provider: clip(n.Resource.Provider, 128),
		Support:  model.SupportGeneric,
		Change:   n.Change,
		Groups:   g.Placement[n.ID],
		Metadata: []model.MetadataField{},
	}
	if n.Supported {
		r.Family, r.Role, r.Support = n.Adapter.Family, n.Adapter.Role, model.SupportSupported
		r.Metadata = projection.Project(n.Resource, n.Adapter.Fields, g.Plan)
	}
	return r
}

func count(s *model.Summary, action model.Action) {
	switch action {
	case model.ActionCreate:
		s.Create++
	case model.ActionUpdate:
		s.Update++
	case model.ActionDelete:
		s.Delete++
	case model.ActionReplace:
		s.Replace++
	case model.ActionRead:
		s.Read++
	case model.ActionForget:
		s.Forget++
	case model.ActionNoOp:
		s.NoOp++
	default:
		s.Unsupported++
	}
}

func title(requested string, snap *input.Snapshot) string {
	if t := strings.TrimSpace(requested); t != "" {
		return clip(t, 200)
	}
	if snap.Kind == input.KindPlan {
		return "Infrastructure plan review"
	}
	return "Recorded infrastructure state"
}

func producer(snap *input.Snapshot) model.Producer {
	return model.Producer{Name: model.ProducerName(snap.Producer), Version: snap.ProducerVersion, FormatVersion: snap.FormatVersion}
}

func completeness(snap *input.Snapshot) model.Completeness {
	switch {
	case snap.Kind == input.KindState:
		return model.Completeness{Status: model.CompletenessComplete}
	case snap.Errored || (snap.Complete != nil && !*snap.Complete):
		return model.Completeness{Status: model.CompletenessIncomplete}
	case snap.Complete == nil:
		return model.Completeness{Status: model.CompletenessNotReported}
	default:
		return model.Completeness{Status: model.CompletenessComplete}
	}
}

func warnings(snap *input.Snapshot, g *graph.Graph) []model.Warning {
	out := []model.Warning{}
	switch completeness(snap).Status {
	case model.CompletenessIncomplete:
		out = append(out, model.Warning{Code: model.WarningIncompletePlan, Message: "The producer reported this plan as incomplete or errored. Some changes may be missing."})
	case model.CompletenessNotReported:
		out = append(out, model.Warning{Code: model.WarningCompletenessNotReported, Message: "The producer did not report whether this plan is complete."})
	}
	for _, n := range snap.Notices {
		code := model.WarningUnknownFormatMinor
		if n.Code == input.NoticeDeposedSkipped {
			code = model.WarningDeposedObjectSkipped
		}
		out = append(out, model.Warning{Code: code, Message: n.Message})
	}
	generic := map[string]int{}
	for _, n := range g.Nodes {
		if !n.Supported {
			generic[n.Resource.Type]++
		}
		if n.Change.Action == model.ActionUnsupported {
			out = append(out, model.Warning{Code: model.WarningUnsupportedAction, Resource: n.ID,
				Message: "This change uses an action tfviz does not recognise. Review it in the producer's own plan output."})
		}
	}
	types := make([]string, 0, len(generic))
	for t := range generic {
		types = append(types, t)
	}
	sort.Strings(types)
	for _, t := range types {
		out = append(out, model.Warning{Code: model.WarningUnsupportedResource,
			Message: fmt.Sprintf("%s is not supported yet (%d %s). It is shown with limited detail and no placement.", clip(t, 128), generic[t], plural(generic[t], "resource", "resources"))})
	}
	return out
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func clip(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := limit
	for cut > 0 && (s[cut]&0xC0) == 0x80 {
		cut--
	}
	return s[:cut]
}
