# Terraform Provider Google Workspace
<a href="https://terraform.io">
    <img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" alt="Terraform logo" align="right" height="50" />
</a>

![Status: Internal Fork](https://img.shields.io/badge/status-internal--fork-blue) [![Releases](https://img.shields.io/github/release/BarruleMedicalGroupLimited/terraform-provider-googleworkspace.svg)](https://github.com/BarruleMedicalGroupLimited/terraform-provider-googleworkspace/releases)
[![LICENSE](https://img.shields.io/github/license/BarruleMedicalGroupLimited/terraform-provider-googleworkspace.svg)](https://github.com/BarruleMedicalGroupLimited/terraform-provider-googleworkspace/blob/main/LICENSE)

This Google Workspace provider for Terraform allows you to manage domains, users, and groups in your Google Workspace.

## About This Fork

This repository is an internal fork of [hashicorp/terraform-provider-googleworkspace](https://github.com/hashicorp/terraform-provider-googleworkspace). HashiCorp has since **archived** the upstream project and no longer maintains or supports it.

This fork is maintained internally by **Barrule Medical Group Limited**. It is not affiliated with, published by, or supported by HashiCorp. Please file issues and contributions against this repository rather than the upstream (archived) one. See [Special Recognition](#special-recognition) below for full attribution to the original project.

## Maintainers

This provider is maintained internally by Barrule Medical Group Limited.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.16

## Upgrading the provider

The Google Workspace provider doesn't upgrade automatically once you've started using it. After a new release you can run

```bash
terraform init -upgrade
```

to upgrade to the latest stable version of the Google Workspace provider. See the [Terraform website](https://www.terraform.io/docs/configuration/providers.html#provider-versions)
for more information on provider upgrades, and how to set version constraints on your provider.

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command or `make build`:
```sh
$ make build
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using The provider

This provider is not published on the public Terraform Registry. Build it from source (see [Building The Provider](#building-the-provider) above) and point Terraform at your local build — see [Instructing Terraform to use a local copy of the provider](.github/CONTRIBUTING.md#instructing-terraform-to-use-a-local-copy-of-the-provider) in `CONTRIBUTING.md`.

Provider configuration and resource/data source reference docs live in [`docs/`](./docs/index.md) in this repository.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).
You can use [goenv](https://github.com/syndbg/goenv) to manage your Go version.
To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

For guidance on common development practices such as testing changes, see the [contribution guidelines](.github/CONTRIBUTING.md).
If you have other development questions we don't cover, please file an issue!

## Special Recognition

This project began as [hashicorp/terraform-provider-googleworkspace](https://github.com/hashicorp/terraform-provider-googleworkspace), created and maintained by the Terraform team at HashiCorp until the repository was archived. All credit for the original design and implementation belongs to HashiCorp and its contributors; see the [LICENSE](./LICENSE) for the full copyright notice.

* [Chase](https://github.com/DeviaVir) - for the excellent work creating the `DeviaVir/terraform-provider-gsuite` provider, the inspiration for the original project.
