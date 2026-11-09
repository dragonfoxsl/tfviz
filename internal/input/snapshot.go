// Package input reads Terraform and OpenTofu JSON exports into a normalised
// snapshot. It is the only package that sees raw, secret-bearing input:
// sensitive values are reduced to markers here and never leave the reader.
package input

import "time"

// SnapshotKind distinguishes a proposed change from recorded state.
type SnapshotKind string

const (
	KindPlan  SnapshotKind = "plan"
	KindState SnapshotKind = "state"
)

// Producer identifies the tool that exported the document.
type Producer string

const (
	ProducerTerraform Producer = "terraform"
	ProducerOpenTofu  Producer = "opentofu"
	ProducerUnknown   Producer = "unknown"
)

// Snapshot is a normalised plan or state.
type Snapshot struct {
	Kind            SnapshotKind
	Producer        Producer
	ProducerVersion string
	FormatVersion   string
	// Timestamp is when the producer generated the plan, when it says so.
	Timestamp *time.Time
	// Complete is nil when the producer does not report completeness.
	Complete *bool
	Errored  bool
	// Resources are ordered by address for determinism.
	Resources []Resource
	// Regions are constant provider regions by provider configuration key
	// ("aws" or "aws.alias"); computed regions are absent.
	Regions map[string]string
	Config  *Config
	Notices []Notice
}

// Resource is one resource instance and, for plans, its proposed change.
type Resource struct {
	Address string
	// Module is the module instance address; empty for the root module.
	Module   string
	Mode     string // "managed" or "data"
	Type     string
	Name     string
	Provider string // namespace/type, e.g. "hashicorp/aws"
	// ProviderKey is the configuration key, e.g. "aws" or "aws.west".
	ProviderKey string

	// Actions are the producer's raw plan actions; nil for state.
	Actions         []string
	ActionReason    string
	ReplacePaths    []string
	Importing       bool
	PreviousAddress string

	Before    Value
	After     Value
	HasBefore bool
	HasAfter  bool
	// Explicit null sides distinguish confirmed absence from omitted data.
	BeforeAbsent bool
	AfterAbsent  bool
}

// NoticeCode classifies a non-fatal reading observation.
type NoticeCode string

const (
	NoticeUnknownFormatMinor NoticeCode = "unknown_format_minor"
	NoticeDeposedSkipped     NoticeCode = "deposed_skipped"
)

// Notice records something the report should disclose.
type Notice struct {
	Code    NoticeCode
	Message string
	Address string
}
