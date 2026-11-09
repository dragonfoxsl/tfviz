# Command line

tfviz reads the JSON that Terraform or OpenTofu exports with `show -json`, and writes one self-contained HTML report. The report opens directly from disk and makes no network requests.

## Generate a plan review

```bash
terraform plan -out=tfplan
terraform show -json tfplan | tfviz plan --input - --output infra-report.html
```

OpenTofu works the same way:

```bash
tofu plan -out=tfplan
tofu show -json tfplan | tfviz plan --input - --output infra-report.html
```

Piping avoids writing the exported JSON to disk. The export can contain plaintext secrets: tfviz removes them from the report, but it cannot sanitise the producer's own plan or state files.

## Generate a state report

From an exported state:

```bash
terraform show -json | tfviz state --input - --output state-report.html
```

Or directly from a version-4 state file (Terraform 0.12 and later, and OpenTofu):

```bash
tfviz state --input terraform.tfstate --output state-report.html
```

Both routes produce the same report. Root module outputs are never read or shown.

Or straight from S3 (see [s3.md](s3.md)):

```bash
tfviz state --input s3://example-state-bucket/prod/terraform.tfstate --aws-profile work --aws-region eu-west-1 --output state-report.html
```

A state report shows Terraform's recorded snapshot. It does not prove that the infrastructure in AWS is currently identical.

## Explore state locally

```bash
tfviz explore --state terraform.tfstate
tfviz explore --state s3://example-state-bucket/prod/terraform.tfstate --aws-profile work --aws-region eu-west-1
```

`explore` serves the same interface from a private server on this machine and prints a link to open. It reads the state once, and again only when you press **Refresh**. Press Ctrl+C to stop.

The explorer is built so that only you, through the printed link, can use it:

- It listens on `127.0.0.1` only, and refuses requests for any other host name, which defends against DNS rebinding.
- The link carries a one-time access token. The token is exchanged for a session cookie (HttpOnly, SameSite=Strict) and removed from the address bar.
- Responses are never cached, cannot be framed, and send no CORS headers, so other websites cannot read them.
- Refresh needs a same-origin request carrying a token embedded in the page.
- It serves only the sanitised report. It never serves raw input, never reads other files, and never performs AWS operations for the browser.

`explore` takes `--state`, `--title`, `--view`, `--safe-share`, `--offline`, `--icons`, the S3 options, `--port` (default: any free port) and `--open` (open your browser).

## Options

| Option | Commands | Meaning |
| --- | --- | --- |
| `--input <file\|->` | plan, state | Exported JSON (or, for `state`, a version-4 `.tfstate` file), or `-` for standard input. Required. |
| `--output <file>` | plan, state | HTML report to write. Required. Written atomically with permissions `0600`. |
| `--title <text>` | plan, state | Report title. |
| `--view architecture\|modules` | plan, state | The structure the report opens in. Viewers can switch at any time. Defaults to `architecture`. |
| `--force` | plan, state | Replace the output file if it exists. Without it, an existing file is never touched. |
| `--safe-share` | plan, state | Replace identifying details with consistent stand-ins. See [Sharing reports](#sharing-reports). |
| `--offline` | plan, state | Refuse any input that would need network access (`s3://`). |
| `--pricing-file <file>` | plan | Opt-in offline cost estimates from your rate catalogue. Shows partial before/after/delta and excluded-resource counts, not a full bill. See [pricing.md](pricing.md). |
| `--usage-file <file>` | plan | Optional monthly usage per resource, such as traffic or storage. Needs a version 2 `--pricing-file`. Without it, fixed charges still show, with a warning that the price can vary. |
| `--icons <folder>` | plan, state, explore | Show official AWS service icons from a pack you downloaded. See [icons.md](icons.md). Without it, reports use tfviz's own symbols. |
| `--aws-profile`, `--aws-region`, `--s3-version`, `--expected-bucket-owner` | state | Credentials and object selection for `s3://` inputs. See [s3.md](s3.md). |
| `--strict` | plan, state | Fail if anything cannot be fully interpreted: unsupported resource types, unrecognised actions, a plan the producer reported as incomplete, skipped deposed objects, or a newer JSON format. No report is written. |

`tfviz version` prints the version.

## Sharing reports

A normal report shows approved names, IDs, ARNs and network details, because reviewers need them. Treat it like any other infrastructure artefact, with your usual access and retention controls.

`--safe-share` replaces identifying details with stand-ins that stay consistent within the report: names, Terraform addresses, module names, resource IDs, ARNs, account IDs, DNS names, IP addresses, CIDR ranges and bucket names. The title becomes generic, and summaries that embed several identifiers (routes and security group rules) are withheld. Service types, topology, regions, availability zones, descriptive settings (such as engine or instance class) and change categories stay. The mapping between stand-ins and real values is never written to the report.

Safe-share reduces disclosure; it does not make a report anonymous. The shape of an architecture can itself be confidential, so share safe-share reports deliberately.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | The report was written. A plan that changes infrastructure is still a success. |
| `1` | Processing or retrieval failed: an unreadable input file, S3 access denied or not found, expired credentials, or an output that could not be written. |
| `2` | Invalid arguments, unsupported input, an output that already exists without `--force`, or a `--strict` failure. |

## What tfviz prints

On success it prints one line to standard error with counts only. For example:

```text
Wrote infra-report.html: 39 resources (38 fully supported, 1 with limited detail). Changes: 3 to add, 6 to change, 2 to replace, 1 to destroy, 1 read during apply. 1 unresolved reference(s).
```

It never prints input content. Error messages explain what to do next and never echo the document.

## Inputs it rejects, with guidance

- **State files in formats other than version 4.** Export them with `terraform show -json` instead.
- **Terraform's streaming `-json` log output.** Save the plan with `-out` and export it instead.
- **Encrypted OpenTofu state or plans.** Export them with `tofu show -json` using your encryption configuration.
- **A plan passed to `state`, or state (exported or raw) passed to `plan`.**
- **JSON format major versions other than 1.**
- **Inputs over the size limits.** These are 512 MiB, nesting deeper than 128 levels, or more than 50,000 resources.
