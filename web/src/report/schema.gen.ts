/* Generated from schema/report.v1.schema.json by scripts/generate-types.mjs. Do not edit. */

/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Count".
 */
export type Count = number;
/**
 * Opaque report-local identifier.
 *
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Id".
 */
export type Id = string;
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Family".
 */
export type Family =
  | "network"
  | "security"
  | "compute"
  | "load_balancing"
  | "database"
  | "streaming"
  | "secrets"
  | "encryption"
  | "observability"
  | "storage"
  | "messaging"
  | "configuration"
  | "other";
/**
 * A value carries a payload only when its status is known. Sensitive values never carry a payload, a hash or a fingerprint.
 *
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "FieldValue".
 */
export type FieldValue =
  | {
      status: "known";
      value: DisplayValue;
    }
  | {
      status: "sensitive" | "unknown" | "absent" | "omitted";
    };
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "DisplayValue".
 */
export type DisplayValue = string | number | boolean | string[];
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "RelationshipType".
 */
export type RelationshipType =
  | "subnet_membership"
  | "security_group_attachment"
  | "security_group_rule"
  | "routing"
  | "route_association"
  | "service_reference"
  | "secret_association"
  | "encryption_key"
  | "log_destination";

/**
 * Sanitised tfviz report model, version 1. Only approved, projected metadata appears here; there is deliberately no generic attribute map.
 */
export interface Report {
  schema_version: "1.0";
  mode: "plan" | "state";
  /**
   * safe_share replaces identifying values with report-local pseudonyms.
   */
  disclosure: "internal" | "safe_share";
  title: string;
  /**
   * The structure the interface opens in.
   */
  initial_view?: "architecture" | "modules";
  generated_at: string;
  tool: Tool;
  producer: Producer;
  source: Source;
  completeness: Completeness;
  summary: Summary;
  coverage: Coverage;
  resources: Resource[];
  relationships: Relationship[];
  unresolved: UnresolvedReference[];
  groups: Group[];
  warnings: Warning[];
  icons?: Icons;
  pricing?: Pricing;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Tool".
 */
export interface Tool {
  name: "tfviz";
  version: string;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Producer".
 */
export interface Producer {
  name: "terraform" | "opentofu" | "unknown";
  version?: string;
  format_version: string;
}
/**
 * Where the input came from. The timestamp describes the input, never the live infrastructure.
 *
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Source".
 */
export interface Source {
  kind: "file" | "stdin" | "s3";
  timestamp_status: "known" | "unavailable";
  timestamp?: string;
  /**
   * S3 object version, when the input was read from a versioned bucket.
   */
  object_version?: string;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Completeness".
 */
export interface Completeness {
  /**
   * not_reported means the producer gave no completeness signal; it must not be shown as complete.
   */
  status: "complete" | "incomplete" | "not_reported";
}
/**
 * Counts of resource changes. Replacements are counted separately, not as an addition plus a destruction.
 *
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Summary".
 */
export interface Summary {
  create: Count;
  update: Count;
  delete: Count;
  replace: Count;
  read: Count;
  forget: Count;
  no_op: Count;
  unsupported: Count;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Coverage".
 */
export interface Coverage {
  resources_total: Count;
  resources_supported: Count;
  resources_generic: Count;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Resource".
 */
export interface Resource {
  id: Id;
  /**
   * Full Terraform address, including count/for_each keys and module instances.
   */
  address: string;
  label: string;
  type: string;
  family: Family;
  /**
   * association resources exist to connect two others (route table associations, rule and attachment resources). They are counted like any resource and may be drawn as connectors.
   */
  role: "entity" | "association";
  mode: "managed" | "data";
  /**
   * Module instance address; empty for the root module.
   */
  module: string;
  /**
   * Provider namespace/type, without the registry host.
   */
  provider: string;
  /**
   * generic resources receive a deliberately limited card with no projected metadata.
   */
  support: "supported" | "generic";
  change: Change;
  groups: ResourceGroups;
  metadata: MetadataField[];
  cost?: CostEstimate;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Change".
 */
export interface Change {
  /**
   * forget removes a resource from state without destroying it.
   */
  action: "no_op" | "create" | "update" | "delete" | "replace" | "read" | "forget" | "unsupported";
  replace_order?: "delete_first" | "create_first";
  /**
   * @maxItems 4
   */
  producer_actions:
    [] | [string] | [string, string] | [string, string, string] | [string, string, string, string];
  reason?: string;
  /**
   * Approved field keys that force replacement; unapproved paths are counted, not named.
   */
  replace_fields?: string[];
  replace_fields_omitted?: Count;
  annotations?: ("imported" | "moved" | "forgotten")[];
  previous_address?: string;
}
/**
 * Primary placement in each view. Multi-subnet membership is expressed through relationships, never by duplicating the resource.
 *
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "ResourceGroups".
 */
export interface ResourceGroups {
  architecture: Id;
  modules: Id;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "MetadataField".
 */
export interface MetadataField {
  key: string;
  label: string;
  before?: FieldValue;
  after?: FieldValue;
  /**
   * Present in plan mode. unknown means the comparison could not be established.
   */
  change_status?: "unchanged" | "changed" | "added" | "removed" | "unknown";
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "CostEstimate".
 */
export interface CostEstimate {
  before_monthly_cents?: number;
  after_monthly_cents?: number;
  delta_monthly_cents?: number;
  before_region?: string;
  after_region?: string;
  reason?:
    | "Estimate unavailable: unsupported action, resource, purchase option, or incomplete pricing evidence."
    | "Estimate withheld because pricing evidence is sensitive."
    | "Partial steady-state subtotal of comparable resources only; excludes unestimated resources and usage costs."
    | "Steady-state instance-hour estimate only; storage, usage and transient replacement costs are not estimated."
    | "Partial steady-state subtotal of comparable resources only; excludes unestimated resources and unprovided usage."
    | "Partial steady-state estimate of supplied components only; excludes unprovided usage, other charges and transient replacement costs."
    | "Estimate unavailable: this resource is billed only by usage and no usage was supplied.";
  replacement_overlap?: boolean;
  partial?: boolean;
  usage_included?: boolean;
  /**
   * @maxItems 12
   */
  missing_usage?:
    | []
    | [
        | "processed_gb"
        | "capacity_unit_hours"
        | "requests"
        | "api_requests"
        | "standard_storage_gb_month"
        | "get_requests"
        | "put_requests"
        | "compute_gb_seconds"
        | "publish_requests"
        | "http_deliveries"
        | "ingested_gb"
        | "stored_gb_month"
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ]
    | [
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        ),
        (
          | "processed_gb"
          | "capacity_unit_hours"
          | "requests"
          | "api_requests"
          | "standard_storage_gb_month"
          | "get_requests"
          | "put_requests"
          | "compute_gb_seconds"
          | "publish_requests"
          | "http_deliveries"
          | "ingested_gb"
          | "stored_gb_month"
        )
      ];
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Relationship".
 */
export interface Relationship {
  id: Id;
  source: Id;
  target: Id;
  type: RelationshipType;
  /**
   * Whether the association exists before the change, after it, or in both.
   */
  presence: "before" | "after" | "both";
  evidence: Evidence;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Evidence".
 */
export interface Evidence {
  /**
   * attribute: a recorded attribute value matched the target's identifier. configuration: an unambiguous configuration reference. association_resource: a dedicated association resource.
   */
  kind: "attribute" | "configuration" | "association_resource";
  field: string;
  via?: Id;
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "UnresolvedReference".
 */
export interface UnresolvedReference {
  source: Id;
  type: RelationshipType;
  field: string;
  reason: "unknown_until_apply" | "external" | "ambiguous" | "sensitive";
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Group".
 */
export interface Group {
  id: Id;
  view: "architecture" | "modules";
  kind:
    | "account"
    | "region"
    | "vpc"
    | "availability_zone"
    | "subnet"
    | "regional_services"
    | "global_services"
    | "unplaced"
    | "module";
  label: string;
  parent?: Id;
  /**
   * unresolved groups collect resources whose placement could not be established.
   */
  placement: "known" | "unresolved";
  /**
   * Opaque report-local identifier.
   */
  resource?: string;
  /**
   * Subnet classification, labelled public or private only when route-table evidence supports it.
   */
  classification?: "public" | "private" | "unknown";
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Warning".
 */
export interface Warning {
  code:
    | "unsupported_action"
    | "unsupported_resource"
    | "unknown_format_minor"
    | "completeness_not_reported"
    | "incomplete_plan"
    | "unresolved_relationship"
    | "deposed_object_skipped"
    | "limit_reached";
  message: string;
  resource?: Id;
}
/**
 * Images from an icon pack the user supplied with --icons. tfviz does not ship provider icons.
 *
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Icons".
 */
export interface Icons {
  images: {
    [k: string]: string;
  };
  /**
   * Resource type to icon id.
   */
  by_type: {
    [k: string]: Id;
  };
}
/**
 * This interface was referenced by `Report`'s JSON-Schema
 * via the `definition` "Pricing".
 */
export interface Pricing {
  currency: string;
  as_of: string;
  hours_per_month: 730;
  summary: CostEstimate;
  usage_provided?: boolean;
  estimated_resources: Count;
  unestimated_resources: Count;
  usage_omitted_resources?: Count;
}
