<!-- Copyright (c) Barrule Medical Group Limited -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# AGENTS.md

Guidance for AI coding agents working in this repository.

## What this repository is

This is an internal fork of [hashicorp/terraform-provider-googleworkspace](https://github.com/hashicorp/terraform-provider-googleworkspace).
HashiCorp has **archived** the upstream project and no longer maintains or
supports it. This fork (`BarruleMedicalGroupLimited/terraform-provider-googleworkspace`)
is maintained internally by **Barrule Medical Group Limited**. It is not
affiliated with, published by, or supported by HashiCorp.

## Attribution rules (do not skip this when touching license/copyright text)

This code is licensed under MPL-2.0, inherited from the upstream project.
When editing files or writing new ones:

- **Existing files** (essentially all of `internal/provider/*.go`, the example
  `.tf` files, etc.) carry a `Copyright (c) HashiCorp, Inc.` /
  `SPDX-License-Identifier: MPL-2.0` header. **Leave that header as-is** when
  modifying an existing file, even substantially — MPL-2.0 does not allow
  stripping or altering the original copyright notice, and the file is still
  majority HashiCorp-authored code even after your edits. Do not add your
  own copyright line to an existing file's header.
- **Wholly new files** you create (a new resource, a new test file with no
  HashiCorp precedent, etc.) should carry `Copyright (c) Barrule Medical
  Group Limited` / `SPDX-License-Identifier: MPL-2.0` instead — that's the
  accurate attribution, since HashiCorp never touched that file.
- **`LICENSE`**: contains both `Copyright (c) 2021 HashiCorp, Inc.` and
  `Portions Copyright (c) 2026 Barrule Medical Group Limited`. Never remove
  the HashiCorp line. Never rewrite the MPL-2.0 legal text itself — that's
  not something to edit casually or "clean up."
- **`CHANGELOG.md`**: a historical record of real PRs merged against the
  actual upstream HashiCorp repository. Don't rewrite or rebrand old
  entries; only append new ones for this fork's own changes going forward.
- Docs/READMEs/branding should point at `BarruleMedicalGroupLimited/terraform-provider-googleworkspace`
  and this repo's own `docs/` folder, not `hashicorp/terraform-provider-googleworkspace`
  or the public Terraform Registry — this provider isn't published there.

If a task involves copyright/license/attribution decisions beyond the rules
above, ask before proceeding rather than guessing — these have real legal
consequences.

## Build, test, lint

```sh
make build      # go install
make test       # fast unit tests, no network
make testacc    # full acceptance suite - TF_ACC=1, hits the real Google
                 # Workspace Admin SDK API, needs real credentials, costs
                 # real API quota, can create/destroy real resources
gofmt -l .       # must be empty
go vet ./...     # must be clean (exit 0) - keep it that way; fix findings
                 # rather than suppressing them
```

Acceptance tests (`TestAcc*` in `internal/provider/*_test.go`) require env
vars documented in `.github/CONTRIBUTING.md` (`GOOGLEWORKSPACE_CUSTOMER_ID`,
`GOOGLEWORKSPACE_DOMAIN`, `GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL`,
credentials, etc.) and are skipped automatically without them. No CI
workflow runs them - they're a local/manual responsibility until a
dedicated test environment exists for this fork.

`.github/workflows/`:
- `test.yml` runs on every push to `main` and every PR: build, `gofmt`,
  `go vet`, `golangci-lint`, unit tests (`make test`, which never sets
  `TF_ACC` so acceptance tests self-skip), and `govulncheck`.
  - `golangci-lint` is scoped to `only-new-issues: true`. That scoping is
    deliberate - there's a backlog of pre-existing findings from before
    this workflow existed; only newly introduced issues fail the build.
    Don't quietly widen that scope without either fixing the backlog
    first or getting a green light to add a temporary `//nolint` inline
    instead.
  - `.golangci.yml` is v2-schema, and the lint action is pinned to a v7.x
    release (which defaults to golangci-lint v2). This isn't optional:
    go.mod targets `go 1.25.0` (needed to pick up CVE-fixed golang.org/
    x/net releases - see govulncheck below), and golangci-lint v1's final
    release was built with go1.24, which refuses outright to analyze a
    newer-targeting module. If go.mod's `go` directive ever drops back
    below what the pinned golangci-lint's own Go toolchain supports,
    that's the failure mode to look for.
  - `govulncheck` (`golang/govulncheck-action`) checks reachable-vulnerability
    exposure via the Go vulnerability database. If it fails, read the
    trace it prints - it tells you which of your own call sites reach the
    vulnerable code, which is usually enough to know whether a dependency
    bump is required or the finding is a false positive for this binary.
- `release.yml` runs on `v*` tag pushes: builds, signs (GPG), and publishes
  a GitHub Release via GoReleaser. Needs `GPG_PRIVATE_KEY`/`PASSPHRASE`
  repo secrets to succeed - see `.github/CONTRIBUTING.md`'s "Releasing"
  section.

## Working conventions

- Push to a branch, never directly to `main`. Don't open a PR unless asked.
- Keep `go vet ./...`, `gofmt -l .`, and `go build ./...` clean before
  committing Go changes.
- When fixing a suspected bug in `internal/provider`, prefer writing a
  small isolated test that reproduces it first, confirm it fails against
  the current code, then fix and confirm it passes — this repo's existing
  tests (e.g. `eventual_consistency_test.go`) already follow that style.
- `internal/provider/logging_transport.go` scrubs sensitive fields (e.g.
  `accessToken`) out of `TF_LOG=DEBUG` request/response dumps, recursively
  through nested objects and arrays (`obfuscateValues`/`obfuscateValue`).
  If you add a new sensitive field name to a request/response body, add it
  to `getValuesToScrub()` and cover it in `logging_transport_test.go`.
- Provider-level secret fields (`access_token`, `credentials` in
  `provider.go`) are marked `Sensitive: true`, matching resource-level
  secrets like `googleworkspace_user.password`. Keep that convention for
  any new secret-bearing schema field.
