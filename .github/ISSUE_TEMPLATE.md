Hi there,

Thank you for opening an issue. This repository is Barrule Medical Group Limited's internal fork of the (now archived) `hashicorp/terraform-provider-googleworkspace`. Please file issues here rather than against the upstream project.

### Terraform Version
Run `terraform -v` to show the version.

### Affected Resource(s)
Please list the resources as a list, for example:
- googleworkspace_user
- googleworkspace_group

If this issue appears to affect multiple resources, it may be an issue with Terraform's core, so please mention this.

### Terraform Configuration Files
```hcl
# Copy-paste your Terraform configurations here. Remove or redact anything
# sensitive first - service account credentials, access tokens, real user
# emails, customer/org IDs, etc.
```

### Debug Output
Please provide a link to a Gist or paste containing the complete debug output: https://www.terraform.io/docs/internals/debugging.html. `TF_LOG=DEBUG` scrubs known sensitive fields (e.g. `accessToken`, `Authorization` headers) from provider request/response logs, but that's not a substitute for reviewing it yourself first - redact anything else sensitive too: real user emails, customer/org IDs, domain names, or any other request/response data you don't want public.

### Panic Output
If Terraform produced a panic, please include the output of the `crash.log`.

### Expected Behavior
What should have happened?

### Actual Behavior
What actually happened?

### Steps to Reproduce
Please list the steps required to reproduce the issue, for example:
1. `terraform apply`

### Important Factoids
Is there anything atypical about your setup we should know? For example: custom `oauth_scopes`, domain-wide delegation vs. direct admin roles, impersonation, non-default org unit structure, etc.

### References
Are there any other issues (open or closed) or pull requests that should be linked here?
