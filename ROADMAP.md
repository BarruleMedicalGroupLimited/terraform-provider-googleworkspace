<!-- Copyright (c) Barrule Medical Group Limited -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Roadmap: Calendar, Drive, Reports, and Security Alerts

Adds four new areas to the provider: **Calendar** resources (rooms/buildings), **Drive** (shared drives), **Reports** (audit/usage), and **Security Alerts** (Alert Center). Scoped deliberately narrow per the brief: enough to *configure, certify, and review* the setup for us and our clients - not full API coverage. See `RESOURCE_COVERAGE.md` for what the provider already covers.

> **Research caveat**: `developers.google.com` is unreachable from this sandbox (egress-blocked), so the reference links and scope/field details below come from search snippets and `pkg.go.dev` (the Go client library docs, which *were* reachable), not a direct read of each reference page. Everything here matches known API shape, but skim the linked pages directly before writing schema code, in case of a recent field addition/deprecation.

## API reference notes

### Calendar resources (rooms, buildings) - Admin SDK Directory API

Already the API this provider is built on (`google.golang.org/api/admin/directory/v1`, already imported). **No new dependency.**

| | |
|---|---|
| Docs | [REST Resource: resources.calendars](https://developers.google.com/workspace/admin/directory/reference/rest/v1/resources.calendars), [REST Resource: resources.buildings](https://developers.google.com/workspace/admin/directory/reference/rest/v1/resources.buildings), [REST Resource: resources.features](https://developers.google.com/workspace/admin/directory/reference/rest/v1/resources.features) (features not fetched directly, but same family) |
| Go client | `admin.ResourcesCalendarsService`, `admin.ResourcesBuildingsService` in the already-imported package. Confirmed via pkg.go.dev: both have `List`/`Get`/`Insert`/`Update`/`Patch`/`Delete`, matching the CRUD pattern every existing resource in this provider already uses. |
| Scope | `https://www.googleapis.com/auth/admin.directory.resource.calendar` (read-write), `.../admin.directory.resource.calendar.readonly` (read-only, for data sources) - confirmed the same scope family covers both calendars and buildings. |
| Key fields (CalendarResource) | `resourceName`, `resourceType`, `capacity`, `buildingId`, `floorName`, `resourceEmail` (read-only, generated), `generatedResourceName` (read-only) |
| Key fields (Building) | `buildingId`, `buildingName`, `floorNames` (ordered list, must have >=1 entry), `address`, `coordinates` |

### Shared Drives - Google Drive API v3

New API for this provider. **New dependency**: `google.golang.org/api/drive/v3`.

| | |
|---|---|
| Docs | [REST Resource: drives](https://developers.google.com/drive/api/v3/reference/drives), [Manage shared drives guide](https://developers.google.com/workspace/drive/api/guides/manage-shareddrives) |
| Go client | `drive.DrivesService` - confirmed via pkg.go.dev: `Create(requestId, drive)`, `Get`, `List`, `Update`, `Delete`, `Hide`, `Unhide`. `Create` requires a client-generated `requestId` for idempotency - a pattern this provider doesn't currently have anywhere, worth a helper. |
| Scope | `https://www.googleapis.com/auth/drive` only - **there is no shared-drive-scoped-only OAuth scope**. This is the one real scoping compromise in this whole roadmap: the scope that lets you administer shared drives is the same one that grants read/write access to file *content* across the domain. Flagged as an open decision below. |
| Key fields (Drive) | `name`, `themeId`, `restrictions` (`adminManagedRestrictions`, `copyRequiresWriterPermission`, `domainUsersOnly`, `driveMembersOnly`, `sharingFoldersRequiresOrganizerPermission` - confirm exact field list against the live docs before writing schema, noted as reconstructed above), `capabilities` (read-only) |

### Reports - Admin SDK Reports API

New API for this provider. **New dependency**: `google.golang.org/api/admin/reports/v1`.

| | |
|---|---|
| Docs | [Admin SDK: Reports API reference](https://developers.google.com/workspace/admin/reports/reference/rest), [activities.list](https://developers.google.com/admin-sdk/reports/reference/rest/v1/activities/list) |
| Go client | Confirmed via pkg.go.dev: `ActivitiesService.List(userKey, applicationName)` (audit events - login, admin, drive, calendar, groups, etc., filterable by actor/time range), `CustomerUsageReportsService.Get(date)`, `UserUsageReportService.Get(userKey, date)`, `EntityUsageReportsService.Get(entityType, entityKey, date)`. |
| Scope | `AdminReportsAuditReadonlyScope` = `.../admin.reports.audit.readonly`, `AdminReportsUsageReadonlyScope` = `.../admin.reports.usage.readonly` - both **read-only by design**, there's no write scope because this API has no write operations. Lowest-risk API in this whole roadmap. |
| Shape | Purely a reporting/query API - no create/update/delete anywhere. Maps to **data sources only**, not resources. |

### Security Alerts - Google Workspace Alert Center API (v1beta1)

New API for this provider. **New dependency**: `google.golang.org/api/alertcenter/v1beta1`. Note the API is still `v1beta1` at the API level (confirmed via the reference URL path) even though the Go client library itself is stable/released - worth a mention in our own resource's docs that the upstream API hasn't graduated to GA.

| | |
|---|---|
| Docs | [Google Workspace Alert Center API reference](https://developers.google.com/workspace/admin/alertcenter/reference/rest), [Choose Alert Center API scopes](https://developers.google.com/workspace/admin/alertcenter/guides/auth), [Alert types reference](https://developers.google.com/workspace/admin/alertcenter/reference/alert-types) |
| Go client | Confirmed via pkg.go.dev: `AlertsService.List/Get/Delete/Undelete/BatchDelete/BatchUndelete/GetMetadata`, `AlertsFeedbackService.Create/List`. |
| Scope | **Exactly one scope exists for this whole API**: `https://www.googleapis.com/auth/apps.alerts` - "See and delete your domain's Google Workspace alerts, and send alert feedback." There's no read-only variant. Since the scope itself permits delete, that's a reason to keep our own resource/data-source surface narrower than the API allows (see design below), not a reason to avoid the API. |
| Key fields (Alert) | `alertId`, `type`, `source`, `createTime`, `startTime`, `endTime`, `data` (type-specific payload), `securityInvestigationToolLink`, `deleted` |
| Shape | `AlertsFeedback` (type e.g. `NOT_RECOMMENDED`, `PERFORMED_ACTION`) is how you record that a human reviewed and acted on an alert - the natural fit for "certify" in the brief. |

## Proposed resource/data source design

Deliberately narrow, per "not full suite control, just enough to configure, certify, and review":

| New resource | New data source | Notes |
|---|---|---|
| `googleworkspace_building` | `googleworkspace_building` (single), `googleworkspace_buildings` (list) | Standard CRUD, mirrors `googleworkspace_org_unit`'s shape closely. |
| `googleworkspace_calendar_resource` | `googleworkspace_calendar_resource` (single), `googleworkspace_calendar_resources` (list) | Standard CRUD. `buildingId` references the building resource above. |
| `googleworkspace_shared_drive` | `googleworkspace_shared_drive` (single), `googleworkspace_shared_drives` (list) | CRUD except intentionally **no support for permanently deleting a non-empty drive via `Delete`'s force option** - require the drive to already be empty, matching Drive API's own default safety behavior, so this provider can't be used to accidentally mass-delete drive contents. |
| *(none)* | `googleworkspace_report_activities`, `googleworkspace_usage_report_customer` | Read-only by nature of the upstream API - no resource makes sense here. This is the "review" half of the brief almost verbatim. |
| `googleworkspace_alert_feedback` (create-only; feedback can't be un-submitted upstream, so no update/delete) | `googleworkspace_alerts` (list, filterable), `googleworkspace_alert` (single) | **Deliberately excludes `alerts.delete`/`batchDelete`** even though the API and its one scope both allow it - an alert deleted via Terraform destroys incident evidence, and "review" doesn't require destroy capability. `googleworkspace_alert_feedback` is how "certify you reviewed this" gets represented in state. |

## Open decisions before implementation starts

1. **Drive's `drive` scope is broad.** It's the only option for shared-drive administration, but if `oauth_scopes` is configured with it, the service account/impersonated user gains read/write to file *content* domain-wide, not just shared-drive metadata. Worth deciding: do we accept that (it's already implied by whoever has the Drive admin role in the console today), or restrict which identity/service account is granted this specific scope so it's not bundled with the broader provider config used for everything else? This is a real security-boundary question, not a code question - answer it before writing `provider.go` changes.
2. **Alert Center's single scope covers delete.** Decided above to just not expose delete in our schema - confirm that's the right call for "certify and review" rather than incident response tooling (if you *do* want Terraform-driven alert dismissal later, that's a deliberate follow-up, not something to slip in accidentally).
3. **Test environment.** All four areas need real data to test against in a dedicated Workspace environment (per `AGENTS.md`, this doesn't exist yet for any acceptance tests). Calendar/Drive/Buildings are easy to seed (create test rooms/drives, tear down via sweepers). Reports and Alert Center are harder - audit log entries and alerts arise from real activity, not something you can `Insert` on demand. Plan for those two to have acceptance tests that assert on the *shape* of whatever data exists (schema validation, pagination, filtering by date range) rather than deterministic fixtures, and treat manufacturing a real test alert (e.g. via a deliberate policy violation in the test tenant) as a nice-to-have, not a blocker.

## Test suite plan

Every new resource/data source gets both layers, matching this repo's existing convention (see `AGENTS.md`):

- **Unit tests**: schema validation (`TestProvider`'s `InternalValidate()` picks up new schema automatically), and dedicated tests for any flatten/expand helpers (e.g. `Drive.Restrictions` <-> Terraform nested block, `Alert.Data` <-> whatever shape we expose). No live API needed - these run in every CI build from day one.
- **Acceptance tests** (`TestAcc*`, gated on `TF_ACC=1`, need the dedicated test environment from decision #3 above):
  - `googleworkspace_building` / `googleworkspace_calendar_resource`: full create/read/update/import/delete cycle, same shape as existing `TestAccResourceOrgUnit_*`.
  - `googleworkspace_shared_drive`: create/read/update/import/delete, plus a negative test confirming delete is refused on a non-empty drive.
  - `googleworkspace_report_activities` / `googleworkspace_usage_report_customer`: read-path tests against the test tenant's real activity - assert on response shape and filtering, not exact content.
  - `googleworkspace_alerts` / `googleworkspace_alert_feedback`: read-path test against whatever alerts exist; feedback-create test if/once we can reliably produce a test alert.
- **Sweepers**: `googleworkspace_building`, `googleworkspace_calendar_resource`, `googleworkspace_shared_drive` all create real, cleanup-needed objects - add sweepers for each, matching `googleworkspace_sweeper_test.go`'s existing `tf-test` prefix convention. Reports/Alerts need no sweeper (nothing created).
- **CI**: no workflow changes needed - `test.yml`'s `unit-test`/`lint`/`vet`/`govulncheck`/`build` jobs already cover any new Go code automatically. `make testacc` stays a manual/local step per `CONTRIBUTING.md` until decision #3's test environment exists; that's the one piece of CI this roadmap doesn't solve by itself.

## Phased roadmap to internal production

Ordered by dependency and risk, not by importance - Reports first because it's genuinely the lowest-risk place to prove out "new API area end-to-end" before touching anything with write access.

**Phase 0 - Decide the open questions above.** Scope-breadth call on Drive, confirm the Alert Center delete-exclusion design, and stand up (or at least plan) the dedicated test Workspace environment. Nothing below should start until the Drive scope question has an answer, since it affects `provider.go`'s scope handling.

**Phase 1 - Reports data sources.** Lowest risk (read-only scope, no new write surface), and a clean way to validate the "add a new Google API family to this provider" mechanics (new `NewXService()` client constructor in `provider_config.go`, new import, `go.mod` update, `govulncheck`/`lint` passing on the new dependency) before anything touches write access.

**Phase 2 - Calendar resources (buildings + calendar resources).** No new dependency, standard CRUD, closely mirrors existing resources (`googleworkspace_org_unit` is the closest template). Good second phase because the implementation pattern is already fully proven in this codebase - it's the least novel phase.

**Phase 3 - Security Alerts.** New dependency, but narrow surface (list/get/feedback only, no delete). Build the delete-exclusion deliberately into the schema/resource design from the start, not as an afterthought.

**Phase 4 - Shared Drives.** Last, because it has the broadest scope and the most consequential write operations (a mis-configured `restrictions` block or an unintended delete has real impact on client data). Build the non-empty-drive delete guard and get a second pair of eyes on the schema before merging.

**Phase 5 - Documentation, security review, certification.** Run this repo's existing security-review posture (see prior session's `AGENTS.md`-documented conventions: `Sensitive` flags on anything secret-shaped, recursive log scrubbing coverage if any new field is ever sensitive, `go vet`/`gofmt`/`lint`/`govulncheck` clean) against all four areas together. Update `RESOURCE_COVERAGE.md` to reflect the new coverage once each phase lands.

**Phase 6 - Internal production rollout.** Dogfood against Barrule's own tenant first (using the areas for our own certification/review needs), then extend to client tenants once the internal dogfooding period surfaces no correctness issues. This is the actual "internal production" bar - not merged-to-main, but "we're using it to run our own compliance reviews without surprises."
