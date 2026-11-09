package model

type Family string

const (
	FamilyNetwork       Family = "network"
	FamilySecurity      Family = "security"
	FamilyCompute       Family = "compute"
	FamilyLoadBalancing Family = "load_balancing"
	FamilyDatabase      Family = "database"
	FamilyStreaming     Family = "streaming"
	FamilySecrets       Family = "secrets"
	FamilyEncryption    Family = "encryption"
	FamilyObservability Family = "observability"
	FamilyStorage       Family = "storage"
	FamilyMessaging     Family = "messaging"
	FamilyConfiguration Family = "configuration"
	FamilyOther         Family = "other"
)

type Role string

const (
	// RoleAssociation resources exist to connect two others; the interface
	// may draw them as connectors, but they are always counted.
	RoleEntity      Role = "entity"
	RoleAssociation Role = "association"
)

type ResourceMode string

const (
	ResourceManaged ResourceMode = "managed"
	ResourceData    ResourceMode = "data"
)

type Support string

const (
	// SupportGeneric resources get a deliberately limited card with no metadata.
	SupportSupported Support = "supported"
	SupportGeneric   Support = "generic"
)

type Resource struct {
	ID       string          `json:"id"`
	Address  string          `json:"address"`
	Label    string          `json:"label"`
	Type     string          `json:"type"`
	Family   Family          `json:"family"`
	Role     Role            `json:"role"`
	Mode     ResourceMode    `json:"mode"`
	Module   string          `json:"module"`
	Provider string          `json:"provider"`
	Support  Support         `json:"support"`
	Change   Change          `json:"change"`
	Groups   ResourceGroups  `json:"groups"`
	Metadata []MetadataField `json:"metadata"`
	Cost     *CostEstimate   `json:"cost,omitempty"`
}

// ResourceGroups is the primary placement in each view. Membership of several
// subnets is expressed through relationships, never by duplicating a resource.
type ResourceGroups struct {
	Architecture string `json:"architecture"`
	Modules      string `json:"modules"`
}

type Action string

const (
	ActionNoOp    Action = "no_op"
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionReplace Action = "replace"
	ActionRead    Action = "read"
	// ActionForget removes a resource from state without destroying it.
	ActionForget      Action = "forget"
	ActionUnsupported Action = "unsupported"
)

type ReplaceOrder string

const (
	ReplaceDeleteFirst ReplaceOrder = "delete_first"
	ReplaceCreateFirst ReplaceOrder = "create_first"
)

type Annotation string

const (
	AnnotationImported  Annotation = "imported"
	AnnotationMoved     Annotation = "moved"
	AnnotationForgotten Annotation = "forgotten"
)

type Change struct {
	Action               Action       `json:"action"`
	ReplaceOrder         ReplaceOrder `json:"replace_order,omitempty"`
	ProducerActions      []string     `json:"producer_actions"`
	Reason               string       `json:"reason,omitempty"`
	ReplaceFields        []string     `json:"replace_fields,omitempty"`
	ReplaceFieldsOmitted int          `json:"replace_fields_omitted,omitempty"`
	Annotations          []Annotation `json:"annotations,omitempty"`
	PreviousAddress      string       `json:"previous_address,omitempty"`
}

type ChangeStatus string

const (
	ChangeUnchanged ChangeStatus = "unchanged"
	ChangeChanged   ChangeStatus = "changed"
	ChangeAdded     ChangeStatus = "added"
	ChangeRemoved   ChangeStatus = "removed"
	// ChangeUnknown means the comparison could not be established.
	ChangeUnknown ChangeStatus = "unknown"
)

type MetadataField struct {
	Key          string       `json:"key"`
	Label        string       `json:"label"`
	Before       *FieldValue  `json:"before,omitempty"`
	After        *FieldValue  `json:"after,omitempty"`
	ChangeStatus ChangeStatus `json:"change_status,omitempty"`
}
