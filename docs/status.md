# Implementation status

This page tracks progress against [the implementation plan](implementation-plan.md). Each item points to its evidence. Items that need people, accounts or a published release are listed separately and are not counted as done.

## Milestones

| Milestone | State | Notes |
| --- | --- | --- |
| M0: validate design and risky assumptions | Done | Fixtures generated without AWS. Offline `file://` report under a strict CSP. Design reviewed. ELK replaced by a structural layout ([ADR 0001](decisions/0001-structural-architecture-layout.md)) |
| M1: plan-to-report slice | Done | Reader, projection, AWS adapters, graph, report builder and CLI. Golden reports and canary tests |
| M2: state exploration and interaction | Done | State JSON, raw v4 state, module view, collapsible groups, search through collapsed groups, focus with breadcrumbs, change and domain filters, keyboard list, richer MSK associations |
| M3: S3 and AWS authentication | Done, except live verification | Exact-object GetObject through the SDK credential chain, SSO guidance and verified permissions ([s3.md](s3.md)). Tested against a local S3-compatible server |
| M4: coverage, hardening, public alpha | Engineering done; alpha pending | Tier A and B adapters, safe-share, local explorer, fuzzing, performance, release pipeline, notices and docs |
| M5: v1 after feedback | Not started | Depends on alpha feedback |

Since M4: official AWS icons through a user-supplied pack (`--icons`, [ADR 0002](decisions/0002-user-supplied-provider-icons.md)). Relationship lines are routed around cards ([ADR 0003](decisions/0003-orthogonal-edge-routing.md)).

Offline pricing: opt-in plan estimates for 16 AWS resource types from explicit local rate catalogues, with optional monthly usage ([pricing.md](pricing.md), [ADR 0004](decisions/0004-offline-pricing.md), [ADR 0005](decisions/0005-usage-aware-estimates.md)). These are partial user-rate estimates, not live AWS quotes or full bills.

## Release checklist (plan §19)

| Criterion | State | Evidence |
| --- | --- | --- |
| Terraform and OpenTofu fixtures cover supported formats, modules, `count`/`for_each`, data sources and references | Done | Two stacks × two producers; nested-module reference test; golden reports |
| Before/after and create/update/delete/replace classifications match; unknown values preserved | Done | Graph and projection tests; golden reports |
| Unknown new IDs use only resolvable evidence | Done | Configuration references only for specific instances; unresolved references reported |
| Multi-subnet services counted once; regional services outside VPCs | Done | Graph placement tests; proxy markers in the interface |
| Unsupported coverage visible; `--strict` predictable | Done | Generic cards, warnings, CLI strict tests |
| Canary secrets absent from reports, JSON and logs | Done | Canaries in marked, unmarked, nested, user-data, environment and opaque values; scanned in every report, stdout and stderr |
| Safe-share removes identifying metadata everywhere | Done | Leak tests over data and visible HTML, plus a browser test |
| Hostile labels, script tags, executable URLs, oversized and deep input handled | Done | Renderer and browser tests; limits; fuzzing |
| Local API unavailable to unrelated origins | Done | Server tests; cross-origin test in three browsers |
| S3 tests: profiles, SSO, workload role, failures, versions, KMS | Partly | Local server and isolated environment cover the request, versions, errors, expired SSO and missing credentials. **A real bucket with an SSO profile and a CI workload role is still to do** |
| Downloaded HTML works from `file://` offline in Chrome, Firefox and Safari | Partly | Tested in Playwright's Chromium, Firefox and WebKit builds. **Branded browsers on separate devices are still to do** |
| No telemetry or remote requests; reduced motion and keyboard paths work | Done | Request logging in e2e; accessibility tests |
| A Jenkins run archives only the report, and it works after the CLI exits | **To do** | The pipeline is documented ([ci.md](ci.md)); not yet run on a Jenkins controller |
| Bundled licences recorded; releases have checksums | Done | Generated notices with licence allowlists; draft release with checksums, SBOMs and provenance |

## Needs a maintainer or external people

- Publishing the first release (push a tag, review the draft, publish).
- S3 verification against a real test bucket, with SSO and a CI workload role.
- A Jenkins run, and branded-browser checks on real devices.
- M4 exit: three external engineers complete the workflows without help.
- Deciding when to add Azure and GCP (deferred beyond v1 in the plan).
