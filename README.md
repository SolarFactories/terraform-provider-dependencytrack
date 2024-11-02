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

## Contributing
Contributions are welcome.

When contributing to resources, or data sources:
1. Implementation, with appropriate input validation.
1. Descriptions within the Schemas, listing options if only valid options are permitted.
1. Examples, to show how to use the new item.
