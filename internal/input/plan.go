package input

import (
	"sort"
	"strings"
	"time"
)

// Plan format 1.2 is the newest version in the tested fixture matrix.
const knownPlanMinor = 2

type planDocument struct {
	FormatVersion    string               `json:"format_version"`
	TerraformVersion string               `json:"terraform_version"`
	Timestamp        string               `json:"timestamp"`
	Complete         *bool                `json:"complete"`
	Errored          bool                 `json:"errored"`
	ResourceChanges  []resourceChangeJSON `json:"resource_changes"`
	Configuration    *configurationJSON   `json:"configuration"`
}

type resourceChangeJSON struct {
	Address         string     `json:"address"`
	PreviousAddress string     `json:"previous_address"`
	ModuleAddress   string     `json:"module_address"`
	Mode            string     `json:"mode"`
	Type            string     `json:"type"`
	Name            string     `json:"name"`
	ProviderName    string     `json:"provider_name"`
	Deposed         string     `json:"deposed"`
	ActionReason    string     `json:"action_reason"`
	Change          changeJSON `json:"change"`
}

type changeJSON struct {
	Actions         []string `json:"actions"`
	Before          any      `json:"before"`
	After           any      `json:"after"`
	AfterUnknown    any      `json:"after_unknown"`
	BeforeSensitive any      `json:"before_sensitive"`
	AfterSensitive  any      `json:"after_sensitive"`
	ReplacePaths    [][]any  `json:"replace_paths"`
	Importing       any      `json:"importing"`
	BeforeAbsent    bool     `json:"-"`
	AfterAbsent     bool     `json:"-"`
}

func readPlan(data []byte) (*Snapshot, error) {
	var doc planDocument
	if err := decode(data, &doc); err != nil {
		return nil, err
	}
	notices, err := checkFormat(doc.FormatVersion, knownPlanMinor)
	if err != nil {
		return nil, err
	}
	if len(doc.ResourceChanges) > MaxResources {
		return nil, newError(CodeTooManyThings, "The plan has more resources than tfviz can process in one report (50,000).")
	}

	snap := &Snapshot{
		Kind:            KindPlan,
		FormatVersion:   doc.FormatVersion,
		ProducerVersion: safeVersion(doc.TerraformVersion),
		Complete:        doc.Complete,
		Errored:         doc.Errored,
		Notices:         notices,
	}
	if t, err := time.Parse(time.RFC3339, doc.Timestamp); err == nil {
		snap.Timestamp = &t
	}
	for _, rc := range doc.ResourceChanges {
		if rc.Deposed != "" {
			snap.Notices = append(snap.Notices, Notice{Code: NoticeDeposedSkipped, Address: rc.Address,
				Message: "A deposed object from an earlier interrupted replacement is not shown."})
			continue
		}
		snap.Resources = append(snap.Resources, planResource(rc))
	}
	sortResources(snap.Resources)
	snap.Producer = producerOf(doc.ResourceChanges)
	if doc.Configuration != nil {
		snap.Config = buildConfig(doc.Configuration)
		snap.Regions = doc.Configuration.constantRegions()
	}
	assignProviderKeys(snap)
	return snap, nil
}

func planResource(rc resourceChangeJSON) Resource {
	c := rc.Change
	// Both sides' sensitivity markers constrain each side: a value marked
	// sensitive before or after is withheld on both.
	sensitive := mergeMarks(c.BeforeSensitive, c.AfterSensitive)
	r := Resource{
		Address:         rc.Address,
		Module:          rc.ModuleAddress,
		Mode:            rc.Mode,
		Type:            rc.Type,
		Name:            rc.Name,
		Provider:        providerSource(rc.ProviderName),
		Actions:         c.Actions,
		ActionReason:    rc.ActionReason,
		PreviousAddress: rc.PreviousAddress,
		Importing:       c.Importing != nil,
		ReplacePaths:    flattenPaths(c.ReplacePaths),
		HasBefore:       c.Before != nil,
		HasAfter:        c.After != nil,
		BeforeAbsent:    c.BeforeAbsent,
		AfterAbsent:     c.AfterAbsent,
	}
	before, after := c.Before, c.After
	if r.HasBefore {
		var other *any
		if r.HasAfter {
			other = &after
		}
		r.Before = buildValue(before, sensitive, nil, other, c.AfterUnknown)
	}
	if r.HasAfter || containsMark(c.AfterUnknown) {
		var other *any
		if r.HasBefore {
			other = &before
		}
		r.After = buildValue(after, sensitive, c.AfterUnknown, other, nil)
		r.HasAfter = true
	}
	return r
}

// mergeMarks combines two sensitivity marker trees: a path is marked if
// either tree marks it or any ancestor of it.
func mergeMarks(a, b any) any {
	if marked(a) || marked(b) {
		return true
	}
	am, aIsMap := a.(map[string]any)
	bm, bIsMap := b.(map[string]any)
	if aIsMap || bIsMap {
		out := map[string]any{}
		for k, v := range am {
			out[k] = mergeMarks(v, bm[k])
		}
		for k, v := range bm {
			if _, done := out[k]; !done {
				out[k] = mergeMarks(nil, v)
			}
		}
		return out
	}
	al, aIsList := a.([]any)
	bl, bIsList := b.([]any)
	if aIsList || bIsList {
		n := max(len(al), len(bl))
		out := make([]any, n)
		for i := range n {
			var x, y any
			if i < len(al) {
				x = al[i]
			}
			if i < len(bl) {
				y = bl[i]
			}
			out[i] = mergeMarks(x, y)
		}
		return out
	}
	return nil
}

// providerSource strips the registry host so Terraform and OpenTofu
// addresses of the same provider match: "hashicorp/aws".
func providerSource(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) >= 3 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return name
}

func producerOf(changes []resourceChangeJSON) Producer {
	for _, rc := range changes {
		if p := producerFromName(rc.ProviderName); p != ProducerUnknown {
			return p
		}
	}
	return ProducerUnknown
}

func flattenPaths(paths [][]any) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		parts := make([]string, 0, len(path))
		for _, step := range path {
			if s, ok := step.(string); ok {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			out = append(out, strings.Join(parts, "."))
		}
	}
	sort.Strings(out)
	return out
}

func sortResources(resources []Resource) {
	sort.SliceStable(resources, func(i, j int) bool { return resources[i].Address < resources[j].Address })
}
