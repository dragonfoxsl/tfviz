// Package model defines the sanitised report model, version 1.
//
// schema/report.v1.schema.json is the source of truth; these types mirror it
// exactly. The model deliberately has no generic attribute map: every value
// that reaches a report passes through an approved MetadataField.
package model

import "time"

// SchemaVersion is the report schema version these types implement.
const SchemaVersion = "1.0"

type Mode string

const (
	ModePlan  Mode = "plan"
	ModeState Mode = "state"
)

type Disclosure string

const (
	DisclosureInternal  Disclosure = "internal"
	DisclosureSafeShare Disclosure = "safe_share"
)

type Report struct {
	SchemaVersion string                `json:"schema_version"`
	Mode          Mode                  `json:"mode"`
	Disclosure    Disclosure            `json:"disclosure"`
	Title         string                `json:"title"`
	InitialView   View                  `json:"initial_view,omitempty"`
	GeneratedAt   time.Time             `json:"generated_at"`
	Tool          Tool                  `json:"tool"`
	Producer      Producer              `json:"producer"`
	Source        Source                `json:"source"`
	Completeness  Completeness          `json:"completeness"`
	Summary       Summary               `json:"summary"`
	Coverage      Coverage              `json:"coverage"`
	Resources     []Resource            `json:"resources"`
	Relationships []Relationship        `json:"relationships"`
	Unresolved    []UnresolvedReference `json:"unresolved"`
	Groups        []Group               `json:"groups"`
	Warnings      []Warning             `json:"warnings"`
	Icons         *Icons                `json:"icons,omitempty"`
	Pricing       *Pricing              `json:"pricing,omitempty"`
}

// Icons come from an icon pack the user supplied with --icons. tfviz does not
// ship provider icons; it embeds the user's copies as data URIs so the
// report stays offline.
type Icons struct {
	// Images maps an opaque icon id to a data:image/svg+xml URI.
	Images map[string]string `json:"images"`
	// ByType maps a resource type to the id of its icon.
	ByType map[string]string `json:"by_type"`
}

type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ProducerName string

const (
	ProducerTerraform ProducerName = "terraform"
	ProducerOpenTofu  ProducerName = "opentofu"
	ProducerUnknown   ProducerName = "unknown"
)

type Producer struct {
	Name          ProducerName `json:"name"`
	Version       string       `json:"version,omitempty"`
	FormatVersion string       `json:"format_version"`
}

type SourceKind string

const (
	SourceFile  SourceKind = "file"
	SourceStdin SourceKind = "stdin"
	SourceS3    SourceKind = "s3"
)

type TimestampStatus string

const (
	TimestampKnown       TimestampStatus = "known"
	TimestampUnavailable TimestampStatus = "unavailable"
)

// Source describes the input. Its timestamp describes the input document,
// never the live state of the infrastructure.
type Source struct {
	Kind            SourceKind      `json:"kind"`
	TimestampStatus TimestampStatus `json:"timestamp_status"`
	Timestamp       *time.Time      `json:"timestamp,omitempty"`
	ObjectVersion   string          `json:"object_version,omitempty"`
}

type CompletenessStatus string

const (
	CompletenessComplete    CompletenessStatus = "complete"
	CompletenessIncomplete  CompletenessStatus = "incomplete"
	CompletenessNotReported CompletenessStatus = "not_reported"
)

type Completeness struct {
	Status CompletenessStatus `json:"status"`
}

// Summary counts resource changes. A replacement is one replace, not an
// addition plus a destruction.
type Summary struct {
	Create      int `json:"create"`
	Update      int `json:"update"`
	Delete      int `json:"delete"`
	Replace     int `json:"replace"`
	Read        int `json:"read"`
	Forget      int `json:"forget"`
	NoOp        int `json:"no_op"`
	Unsupported int `json:"unsupported"`
}

type Coverage struct {
	ResourcesTotal     int `json:"resources_total"`
	ResourcesSupported int `json:"resources_supported"`
	ResourcesGeneric   int `json:"resources_generic"`
}

type WarningCode string

const (
	WarningUnsupportedAction       WarningCode = "unsupported_action"
	WarningUnsupportedResource     WarningCode = "unsupported_resource"
	WarningUnknownFormatMinor      WarningCode = "unknown_format_minor"
	WarningCompletenessNotReported WarningCode = "completeness_not_reported"
	WarningIncompletePlan          WarningCode = "incomplete_plan"
	WarningUnresolvedRelationship  WarningCode = "unresolved_relationship"
	WarningDeposedObjectSkipped    WarningCode = "deposed_object_skipped"
	WarningLimitReached            WarningCode = "limit_reached"
)

type Warning struct {
	Code     WarningCode `json:"code"`
	Message  string      `json:"message"`
	Resource string      `json:"resource,omitempty"`
}
