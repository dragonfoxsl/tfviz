# tfviz

Turn Terraform and OpenTofu plans and state into calm, interactive infrastructure diagrams: a single, self-contained HTML file you can open offline or download from CI.

> **Status: alpha candidate.** Plan reports, state reports (local, raw `.tfstate` or S3), the local explorer and safe-share work end to end for the [supported AWS resource types](docs/support.md). They are tested against synthetic Terraform 1.16 and OpenTofu 1.13 fixtures. No release has been published yet. See [docs/status.md](docs/status.md).

## Principles

- **Local and private.** No account, no hosted service and no telemetry. Local-file workflows run entirely offline.
- **Sanitised by construction.** The renderer only ever receives an approved, projected report model. Raw Terraform attributes, secrets and input variables never reach the HTML.
- **Honest semantics.** State is a recorded snapshot and a plan is a proposal. Relationships show recorded associations, not observed traffic.

## Usage

```bash
terraform plan -out=tfplan
terraform show -json tfplan | tfviz plan --input - --output infra-report.html
```

Open `infra-report.html` in a browser. It works offline, with no server. You can also explore state through a private local server (`tfviz explore --state terraform.tfstate`), or read state straight from S3. To show official AWS service icons, pass the icon pack you downloaded with `--icons` ([docs/icons.md](docs/icons.md)). See [docs/cli.md](docs/cli.md), [docs/s3.md](docs/s3.md) and [docs/support.md](docs/support.md).

## Development

Requirements: Go 1.27+, Node 26+, pnpm 12+. Fixture regeneration also needs Terraform and OpenTofu.

```bash
make test     # Go and web checks (builds the web bundle first)
make build    # web bundle plus the Go binary in bin/
make e2e      # renders reports and opens them offline via file:// in three browsers
```

```bash
make build
bin/tfviz plan --input testdata/producer/terraform-1.16/plan.json --output out/report.html
```

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## Documentation

- [Installing and verifying releases](docs/install.md)
- [Command line](docs/cli.md)
- [Offline change cost estimates](docs/pricing.md)
- [Using tfviz in CI](docs/ci.md)
- [Reading state from S3](docs/s3.md)
- [Supported resource types](docs/support.md)
- [Tested compatibility](docs/compatibility.md)
- [Security model](docs/security-model.md)
- [Implementation status](docs/status.md)

## Licence

Apache-2.0. See [LICENSE](LICENSE).
