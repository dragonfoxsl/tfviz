# ADR 0005: Multi-service estimates with optional usage

- **Status:** Accepted
- **Amends:** [ADR 0004](0004-offline-pricing.md), which covered EC2 and RDS instance-hours and deferred storage and usage

## Context

Reviewers want cost context for more than EC2 and RDS, and many services bill
partly or wholly by traffic, storage or requests. A plan does not record usage,
and most teams will not have usage figures for every resource.

## Decision

Add a version 2 catalogue with a fixed vocabulary of per-type components: a
fixed charge, usage charges, or both. Catalogues cannot define expressions.
Add `plan --usage-file <local-json>` for optional monthly quantities per
resource and side. Version 1 catalogues keep working unchanged.

Missing usage never hides a fixed charge. The report shows the charge and warns
that the price can vary, naming the left-out metrics from the fixed vocabulary
in an optional typed `missing_usage` list. A usage-only resource with no usage
stays **Not estimated**: missing usage is never zero.

With a fixed charge, a metric counts only if every live side of the change can
price it, so a delta never compares different sets of charges. The report model
gains `missing_usage` and a `usage_omitted_resources` count. Neither carries
user input. Sensitivity withholds the estimate and its warning.

The report converts monthly figures to daily (24 hours) and weekly (168 hours)
views in the browser. These are normalised averages, using the 730-hour month.

## Consequences

- Usage-based charges are as accurate as the operator's rates and quantities.
- One-sided usage on an update is left out when a fixed charge exists. The
  warning names it; supplying both sides includes it.
- Charges with no component, such as RDS storage, remain excluded, and the
  report says other usage can change the price.
- State and explore reports remain unpriced.

See [pricing.md](../pricing.md) for the catalogue and usage contract.
