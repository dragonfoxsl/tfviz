# ADR 0004: Opt-in offline instance-hour change estimates

- **Status:** Accepted, amended by [ADR 0005](0005-usage-aware-estimates.md)
- **Amends:** Implementation plan §3, which deferred all cost estimation

## Context

Reviewers want a cost delta alongside infrastructure changes. A full cloud bill
cannot be derived from a plan: usage, billing terms and replacement duration are
not recorded. A hosted estimator or live price lookup would also change tfviz's
privacy and offline contract.

## Decision

Add `plan --pricing-file <local-json>` with no new dependencies. Start with
EC2 default-tenancy instance-hours and RDS DB instance-hours, matched against
an explicit user-supplied version-1 catalogue. Match every required selector
and each side's recorded region, never guess unknowns or send input externally.
Use fixed-point arithmetic, 730 hours/month and half-up per-resource rounding.

Extend the version-1 sanitised report schema with optional typed `pricing` and
resource `cost` fields. No raw attributes or catalogue records cross the
renderer boundary. Required sensitive evidence suppresses estimates on both
sides. Existing reports without the flag retain their previous model/output.

Show before/after/delta in the inspector and a comparable-resource subtotal
with excluded-resource counts in the header. Never present missing costs as
zero. Confirmed absence is distinguished from omitted producer data before
assigning zero to a create/delete side. Imports, forgotten resources, reads and
data sources do not imply billable additions or savings. Create-first
replacements explicitly exclude temporary overlap costs.

## Consequences

- Rates are operator-maintained estimates, not verified AWS quotes.
- Coverage is intentionally narrow; storage, usage, billing adjustments and
  transient costs remain deferred. State/explore pricing is not implemented.
- AMI-specific EC2 rates keep OS/licensing assumptions with the rate supplier.
- Safe-share retains computed costs and broad regions, which may be sensitive
  business information even without infrastructure identifiers.
- Removal of the flag returns to unpriced reports; there are no network,
  infrastructure or persisted-state migrations to roll back.

See [pricing.md](../pricing.md) for the catalogue and estimate contract.
