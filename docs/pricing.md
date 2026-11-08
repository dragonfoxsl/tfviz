# Offline change estimates

Pricing is opt-in and supported for `plan` reports only:

```bash
terraform show -json tfplan | tfviz plan --input - --output review.html \
  --pricing-file rates.json --usage-file usage.json --offline
```

`--usage-file` is optional. Without it, resources that have a fixed charge still
show that charge, with a warning that the price can vary with usage.

Supply your own versioned local JSON catalogue. tfviz does not fetch AWS prices,
validate rates against AWS, upload plans, or calculate your actual bill. The rate
date is supplied by you, not a claim that a rate is current. Review the catalogue
against your purchase terms before using an estimate for a decision.

## Catalogue version 2

This example is **synthetic test data, not current or historical AWS pricing**:

```json
{
  "version": "2",
  "currency": "USD",
  "as_of": "2026-01-01",
  "rates": [
    {
      "resource_type": "aws_nat_gateway",
      "region": "us-east-1",
      "selectors": { "connectivity_type": "public" },
      "component": "gateway_hours",
      "unit": "hour",
      "unit_price": "0.1"
    },
    {
      "resource_type": "aws_nat_gateway",
      "region": "us-east-1",
      "selectors": { "connectivity_type": "public" },
      "component": "processed_gb",
      "unit": "GB",
      "unit_price": "0.05"
    }
  ]
}
```

Each rate names one component of one resource type. It must have exactly the
selectors listed below, and they must match known plan values exactly. Booleans
are written as the strings `"true"` and `"false"`.

| Resource type | Selectors | Fixed component (unit) | Usage components (unit) |
| --- | --- | --- | --- |
| `aws_instance` | `instance_type`, `ami`, `tenancy` | `instance_hours` (hour) | — |
| `aws_db_instance` | `instance_class`, `engine`, `multi_az`, `license_model` | `instance_hours` (hour) | — |
| `aws_ebs_volume` | `type` | `storage_gb_month` (GB-month), × `size` | — |
| `aws_nat_gateway` | `connectivity_type` | `gateway_hours` (hour) | `processed_gb` (GB) |
| `aws_lb` | `load_balancer_type` | `load_balancer_hours` (hour) | `capacity_unit_hours` (capacity-unit-hour) |
| `aws_eks_cluster` | `version`, `upgrade_policy.support_type` | `cluster_hours` (hour) | — |
| `aws_kms_key` | `key_usage`, `customer_master_key_spec`, `multi_region` | `key_month` (key-month) | `requests` (request) |
| `aws_secretsmanager_secret` | none | `secret_month` (secret-month) | `api_requests` (request) |
| `aws_elasticache_cluster` | `node_type`, `engine`, `engine_version` | `node_hours` (node-hour), × `num_cache_nodes` | — |
| `aws_redshift_cluster` | `node_type`, `cluster_type` | `node_hours` (node-hour), × `number_of_nodes` | — |
| `aws_vpc_endpoint` | `vpc_endpoint_type`, `service_name` | `endpoint_hours` (endpoint-hour), × subnets | `processed_gb` (GB) |
| `aws_s3_bucket` | none | none | `standard_storage_gb_month` (GB-month), `get_requests` (request), `put_requests` (request) |
| `aws_lambda_function` | `architecture`, `package_type`, `memory_size` | none | `requests` (request), `compute_gb_seconds` (GB-second) |
| `aws_sqs_queue` | `fifo_queue` | none | `requests` (request) |
| `aws_sns_topic` | `fifo_topic` | none | `publish_requests` (request), `http_deliveries` (delivery) |
| `aws_cloudwatch_log_group` | `log_group_class` | none | `ingested_gb` (GB), `stored_gb_month` (GB-month) |

AMI matching is intentional: the rate supplier must account for that image's
operating system and licensing. Only EC2 default tenancy, without Spot/market
options, is supported. The RDS hourly rate must already cover the selected
Single-AZ or Multi-AZ deployment; tfviz does not multiply it again. Only
Interface VPC endpoints are estimated. Aurora cluster instances and resource
types not listed are not estimated.

Currency is one three-uppercase-letter code for the entire catalogue. Amounts
use two decimal places (one displayed minor unit is 1/100 catalogue currency
unit); this is not a currency-conversion or currency-specific rounding service.
Unit prices are nonnegative decimal strings with at most twelve fractional
places and a maximum of 1,000,000 currency units per unit. No exponents,
duplicate keys/rates, unknown fields or extra selectors are accepted. Catalogues
are limited to 4 MiB and 10,000 rates and must be regular local files.

Version 1 catalogues still work. They price only EC2 and RDS instance-hours,
with `hourly_price` in place of `component`, `unit` and `unit_price`, and at
most six fractional places. Usage files need a version 2 catalogue.

## Usage file

Usage is keyed by Terraform address. Quantities are **monthly**, and each side
of the change is separate: tfviz never copies `before` to `after`.

```json
{
  "version": "1",
  "resources": {
    "aws_nat_gateway.main": {
      "before": { "processed_gb": "120" },
      "after": { "processed_gb": "150.5" }
    },
    "aws_sqs_queue.orders": {
      "after": { "requests": "2500000" }
    }
  }
}
```

Quantities are decimal strings with at most six fractional places, up to
1,000,000,000 per month. Every address must be a managed AWS resource in the
plan with a type listed above, and every metric a usage component of that type.
Anything else rejects the whole file. Files are limited to 4 MiB and 10,000
resources and must be regular local files. tfviz never exports usage quantities
or addresses from this file, only the amounts computed from them.

## When usage is missing

| Situation | Report shows |
| --- | --- |
| Fixed charge, all usage supplied | Fixed charge plus usage |
| Fixed charge, some or no usage supplied | The fixed charge, and any usage that can be compared, with a warning naming the metrics left out |
| Fixed charge, nothing tfviz can take as usage | The fixed charge, with a warning that other usage can change the price |
| Usage-only resource, no usage supplied | **Not estimated**, with a warning naming the metrics to supply |
| Usage-only resource, some usage supplied | The supplied usage, with a warning naming the rest |

With a fixed charge, a metric counts only if every side of the change can price
it. An update with usage for `after` alone therefore shows the fixed charge on
both sides and names that metric in the warning, so the delta never compares
different sets of charges. Supply both `before` and `after` to include it. The
same applies to usage supplied without a matching rate.

The header counts how many resources have usage charges left out.

## Reading the estimates

- The **Cost period** control in the header and inspector switches between
  daily (24 hours), weekly (168 hours) and monthly (730 hours). Daily and weekly
  are normalised averages of the monthly figure, not calendar bills.
- Each resource shows before, after and signed delta beside its change details.
  Fixed hourly charges use `unit_price × 730`, and the resource's components are
  summed and rounded half-up to two decimal places once per resource, using
  fixed-point arithmetic.
- Regions are resolved independently for each side from explicit resource
  `region` or a compatible resource ARN. No region is guessed from an AZ,
  another resource, your default profile or the state bucket's region.
- Unknown, sensitive or missing required evidence, missing rates and unsupported
  resource/action types display **Not estimated**, not a zero-cost claim.
  Sensitivity on either side suppresses both sides' estimates and any usage
  warning derived from them.
- A confirmed create has a zero before cost; a confirmed delete has a zero
  after cost. Imports, data sources, reads and removal from state without
  destruction do not imply savings and are not estimated.
- The header is a **resource-cost subtotal**, not a full bill. Only resources
  with comparable estimates for both sides contribute. Coverage counts show
  how many resources are excluded, even if one side has a known cost.
- Replacements compare steady-state costs. Create-first replacements warn
  that temporary overlap is additional and **not estimated**. Neither overlap
  duration nor partial-month runtime can be inferred from a plan.
- Charges that have no component above, such as RDS storage, IOPS, backups and
  data transfer, along with discounts, reservations, Savings Plans, credits,
  free tiers and taxes, are excluded.
- Targeted/incomplete plans still describe only their reported scope. Existing
  completeness warnings remain authoritative; the subtotal is never a whole
  account or whole-estate forecast.

Reports remain self-contained and offline. Only computed amounts, broad regions,
fixed explanation text, fixed usage metric names, currency and date reach the
sanitised report model—not catalogue selectors, AMIs, rate entries, usage
quantities or file paths. `--safe-share` retains the estimates and regions;
costs and topology may themselves be confidential. Omit `--pricing-file` to
preserve the existing unpriced output.

`--strict` still checks infrastructure interpretation, not pricing coverage.
A missing price is a visible coverage gap rather than a failed plan report.
