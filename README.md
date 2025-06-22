## Terraform Provider for DependencyTrack

Uses [Terraform Plugin Framework]("https://github.com/hashicorp/terraform-plugin-framework) as a template from which this is developed.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22
- [DependencyTrack](https://dependencytrack.org)
  - A [Docker Compose](https://docs.docker.com/compose) file is provided to start a local instance with UI at `http://localhost:8080` and API at `http://localhost:8081`

## Contents
- [Source code](./internal/provider)
- [Examples](./examples)
- [Generated Documentation](./docs)

## Developing

1. Clone the repository
1. Enter the repository directory
1. Run `git config core.hooksPath .hooks` to setup the use of hooks, which are located in [.hooks](./.hooks) directory.
1. Run `go install` within a shell.
1. Configure [provider overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) so that Terraform uses the local installation
	```
	provider_installation {
		dev_overrides {
			"registry.terraform.io/solarfactories/dependencytrack" = "$HOME/go/bin"
		}
		direct {}
	}
	```
1. Create required resources for Data Resource Tests
	1. Create a `Project` with name `Project_Data_Test` and version `1`
	1. Create a `ProjectProperty` on `Project_Data_Test` with `Group1`, `Name1`, `Value1`, `STRING`, `Description1`
	1. Create a `ProjectProperty` on `Project_Data_Test` with `Group2`, `Name2`, `2`, `INTEGER`, `Description2`
	1. Create a `Tag` on `Project_Data_Test` with `project_data_test_tag`

## Contributing
Contributions are welcome, either as a PR, or raising an issue to request functionality.

When contributing to resources, or data sources:
1. Implementation, with appropriate input validation.
1. Descriptions within the Schemas, listing options if only valid options are permitted.
1. Examples, to show how to use the new item.

## Supported versions
Various resources have minimum DependencyTrack API versions, which are documented within their descriptions.
The following versions are tested and supported with any combination from options.
Other API versions may work, with a subset of functionality, but are not guaranteed.
The latest patch version within a minor release is supported, even if it might not be tested - PR's to update would always be welcome.
The list of API Versions will grow as functionality adapts to allow tests to pass, which at present is only a small subset.
The latest 2 patches within the latest minor version will be tested, and supported to allow for continued support while migrating.
- Terraform: `1.0` -> `1.12`
- DependencyTrack: `4.11.7`, `4.12.7`, `4.13.0`, `4.13.1`, `4.13.2`
