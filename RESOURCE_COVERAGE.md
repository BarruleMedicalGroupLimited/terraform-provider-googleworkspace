<!-- Copyright (c) Barrule Medical Group Limited -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Google Workspace resource coverage

An inventory of what this provider actually manages today, and what it doesn't. Written for planning purposes - "can we manage X with this provider" - not as end-user Terraform documentation (see `docs/` for that).

Current counts: **15 resources**, **15 data sources**, backed by **5** of Google's Workspace-related APIs.

## Resources

| Resource | Manages | Backing API |
|---|---|---|
| `googleworkspace_user` | User accounts: profile, name, password, org unit, aliases, custom schema field values, recovery email/phone, suspension, IMAP/POP settings, language, keywords, external IDs, relations, addresses, phones, etc. | Admin SDK Directory API |
| `googleworkspace_user_delegate` | Gmail delegated-access grants on a user's mailbox | Gmail API |
| `googleworkspace_group` | Group accounts: email, name, description | Admin SDK Directory API |
| `googleworkspace_group_member` | A single group membership (one resource per member) | Admin SDK Directory API |
| `googleworkspace_group_members` | All memberships of a group, managed as one resource | Admin SDK Directory API |
| `googleworkspace_group_settings` | Group access/moderation settings (who can post, join, view; moderation; spam handling) | Groups Settings API |
| `googleworkspace_dynamic_group` | Groups whose membership is computed by a query (e.g. `user.organizations.exists(...)`) rather than managed manually | Cloud Identity API |
| `googleworkspace_org_unit` | Organizational unit tree (create/move/rename OUs) | Admin SDK Directory API |
| `googleworkspace_domain` | Primary/secondary domains on the account | Admin SDK Directory API |
| `googleworkspace_domain_alias` | Domain aliases | Admin SDK Directory API |
| `googleworkspace_role` | Custom admin roles and their privilege grants | Admin SDK Directory API |
| `googleworkspace_role_assignment` | Assigning a role to a user or group, at customer or OU scope | Admin SDK Directory API |
| `googleworkspace_schema` | Custom user schema definitions (the field definitions themselves, not the values - values are set via `googleworkspace_user`'s `custom_schemas`) | Admin SDK Directory API |
| `googleworkspace_chrome_policy` | Chrome browser/OS policy values applied to an OU or group | Chrome Policy API |
| `googleworkspace_gmail_send_as_alias` | Gmail "send mail as" aliases (including SMTP relay config for the alias) | Gmail API |

## Data sources

| Data source | Reads | Backing API |
|---|---|---|
| `googleworkspace_user` | A single user | Admin SDK Directory API |
| `googleworkspace_users` | Multiple users, with query filtering | Admin SDK Directory API |
| `googleworkspace_group` | A single group | Admin SDK Directory API |
| `googleworkspace_groups` | Multiple groups, with query filtering | Admin SDK Directory API |
| `googleworkspace_group_member` | A single group membership | Admin SDK Directory API |
| `googleworkspace_group_members` | All memberships of a group (optionally including derived/nested membership) | Admin SDK Directory API |
| `googleworkspace_group_settings` | A group's access/moderation settings | Groups Settings API |
| `googleworkspace_dynamic_group` | A single dynamic group, by email or ID | Cloud Identity API |
| `googleworkspace_org_unit` | A single OU, by path or ID | Admin SDK Directory API |
| `googleworkspace_domain` | A single domain | Admin SDK Directory API |
| `googleworkspace_domain_alias` | A single domain alias | Admin SDK Directory API |
| `googleworkspace_role` | A single custom or built-in role, by name or ID | Admin SDK Directory API |
| `googleworkspace_privileges` | The full list of privileges available to grant via roles | Admin SDK Directory API |
| `googleworkspace_schema` | A single custom user schema definition | Admin SDK Directory API |
| `googleworkspace_chrome_policy_schema` | Chrome policy schema metadata (what policies exist and their value shapes) | Chrome Policy API |

## APIs used, and how completely

| API | Used for | Coverage |
|---|---|---|
| **Admin SDK Directory API** | Users, groups, group members, org units, domains, domain aliases, roles, role assignments, custom schemas | The bulk of the provider. Notably absent even within this API's own surface: **Chrome OS devices**, **mobile devices**, **calendar resources** (rooms/equipment), and **building/feature** resources - all manageable via this same API but not exposed as Terraform resources here. |
| **Admin SDK Data Transfer API** | Transferring a departing user's data (Drive/Calendar ownership, etc.) to another user | Used *internally only*, as a step inside `googleworkspace_user`'s delete flow (`resource_user.go`) - not exposed as its own resource or data source. You can't currently declare a data-transfer application list or trigger a standalone transfer via this provider. |
| **Groups Settings API** | Group access/moderation config | Fully exposed via `googleworkspace_group_settings`. |
| **Cloud Identity API** | Dynamic group membership queries | Only the dynamic-group-membership piece is used. Cloud Identity also covers things like Security Groups, Group hierarchies/labels beyond what Directory API offers, and device management under some editions - none of that is used here. |
| **Chrome Policy API** | Chrome browser/OS policy assignment | Policy read/write and schema discovery are both covered. Chrome *device* inventory/management (as opposed to policy) is a different API surface (Admin SDK's `chromeosdevices`) and isn't covered. |
| **Gmail API** | Send-as aliases, mailbox delegation | Narrow and intentional - just the two admin-relevant pieces. General mailbox content/settings (filters, labels, vacation responder, forwarding rules, etc.) aren't covered, which is normal for an admin-focused provider - those are end-user Gmail settings, not typically something you'd manage via Terraform. |

## Not covered at all

Google Workspace surface area with no resource or data source in this provider, despite being commonly relevant to an admin's IaC needs:

- **Reports API** - audit logs, usage reports. No way to query these via this provider (not really a Terraform-shaped need anyway - it's read-heavy telemetry, not desired-state config).
- **Alert Center API** - security/admin alerts.
- **Vault API** - retention holds, eDiscovery matters.
- **Chrome OS device management** - device inventory, remote actions (wipe, disable), device-level (not policy-level) settings. Admin SDK Directory API's `chromeosdevices` resource is unused.
- **Mobile device management** - Admin SDK Directory API's `mobiledevices` resource is unused.
- **Calendar resources** - meeting rooms, equipment, buildings, features (Admin SDK Directory API's `resources.calendars`/`resources.buildings`/`resources.features` are unused).
- **Drive/Shared Drives** - no Drive API usage at all (shared drive creation/membership, org policies for Drive, etc.).
- **Sites, Meet, Groups Migration API** - no usage.
- **Data Transfer as a first-class resource** - see note above; it's wired up internally for user deletion only.

## Where to look for the authoritative, per-attribute picture

This table is deliberately coarse (resource-level, not attribute-level). For exactly which fields each resource/data source exposes, the schema in `internal/provider/*.go` (or the rendered docs in `docs/resources/` and `docs/data-sources/`) is the source of truth - this file is a map of *what exists*, not a substitute for reading the schema of what you specifically need.
