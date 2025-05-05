## 1.12.2

#### FEATURES
- Add explicit support for DependencyTrack `4.13.1`, with testing.

#### MISC
- Increase level of standardised logging across all resources and data sources, to use standardised log structure.
- Address disabled linting rules, by actioning, correcting and enabling.

#### DEPENDENCIES
- `golangci/golangci-lint-action` `7.0.0` -> `8.0.0`

## 1.12.1

#### FIXES
- Within Update `comment` on `dependencytrack_team_apikey` was not being written back to state using return value from DependencyTrack.
- Within Create, Update, the permission list was assigned from a `dtrack.Team`, which may have been `nil`.
- Within Import, `dependencytrack_acl_mapping` would still write to state if unable regardless of whether UUIDs parsed successfully.

#### MISC
- Standardised log error reporting when failing to parse a `UUID`.

## 1.12.0

#### FEATURES
- Add `dependencytrack_ldap_team_mapping` resource to manage dynamic membership of Teams, from LDAP Servers.

#### FIXES
- Fix example for `dependencytrack_oidc_group_mapping` incorrectly using `dependencytrack_oidc_group` value for `team`.

## 1.11.0

#### FEATURES
- Add `dependencytrack_acl_mapping` resource to manage Portfolio Access Control for Projects.

#### ISSUES
- [Fixed in `1.12.1`] Import of `dependencytrack_acl_mapping` would import regardless of whether UUIDs parsed successfully.

#### DEPENDENCIES
- `github.com/DependencyTrack/client-go` `0.16.0` -> `main`

## 1.10.2

#### FIXES
- `comment` on `dependencytrack_team_apikey` resource was improperly set upon creation to an empty string.
	- Thanks to `@acidghost` for contributing a fix.
	- Added regression test within `team_apikey_resource_test.go`.

#### DEPENDENCIES
- `actions/download-artifact` `4.2.1` -> `4.3.0`

## 1.10.1

#### MISC
- Add support for DependencyTrack `4.13.x`, updating README to reflect.
- Add `public_id`, `masked`, `legacy` fields to `dependencytrack_team_apikey`, with `masked` being available in earlier versions of API.
- Remove issue comment triage workflow, as it is unused, and so causes unnecessary action runs.
- Update `docker_compose.yml` file to use an external `postgres` database, as recommended.
	- GitHub actions are lagging, due to inability to manage dependencies between job services.

#### ISSUES
- [Fixed in `1.10.2`] Comments on API keys are set to an empty string in state upon creation. Thanks to `@acidghost` for reporting.

## 1.10.0

#### FEATURES
- Add `dependencytrack_policy` resource to manage a policy to apply to projects.
- Add `dependencytrack_policy_condition` resource to manage the contents of policies.
- Add `dependencytrack_policy_project` resource to select which projects should have a policy applied to them.
- Add `dependencytrack_policy_tag` resource to select which tags should have a policy applied to them.

#### MISC
- Add missing example for `dependencytrack_team_permissions` resource.
- Removed HTTP Patch for `authenticationRequired` within `Repository` requests, as SDK has been updated to include the missing field.

#### FIXES
- Fix references within examples to resources without identifiers.
- Fix example for `dependencytrack_oidc_group_mapping` incorrectly using `dependencytrack_config_property`.

#### DEPENDENCIES
- `actions/setup-node` `4.3.0` -> `4.4.0`
- `github.com/DependencyTrack/client-go` `0.15.0` -> `0.16.0`
- `golang.org/x/net` `0.37.0` -> `0.38.0`
- `golang.org/x/net` `0.36.0` -> `0.38.0` in `/tools`

## 1.9.0

#### FEATURES
- Add `dependencytrack_team_permissions` resource to canonically manage the permissions assigned to a Team.

#### ISSUES
- [Fixed in `1.12.1`] Permission list is assigned from a potentially `nil` `dtrack.Team`.

#### MISC
- Remove deprecated field from within `golangci` config file for `goconst`.

## 1.8.2

#### MISC
- Add support for DependencyTrack API `4.11.x`, updating README to reflect.
- Add testing of multiple DependencyTrack API versions in a matrix, as opposed to just `latest`
- Updated golang lint config to be strict, adding explicit exceptions.
- Add golangci config validation to `make lint` command.
- Add exclusion to `^TestAcc` when running `make test` as these tests are not run without `TF_ACC="1"`.
- Updated `ProviderData` from `*dtrack.Client`, to a struct containing `*dtrack.Client` and Semver information, for resources an datasources.
- Added request within Provider Configuration, to retrieve API version, to be used for compatibility within Provider.
- Increased verbosity of `Debug` logs within `dependencytrack_project`, to cover non-sensitive attributes.
- Added testing utilities, as well as validation of Semver value returned by API.

#### FIXES
- Set minimum TLS version used by TLS Client to Version 1.3, reducing security exposure from weak ciphers.
- Created separate `dependencytrack_project` within `dependencytrack_project_property` tests, due to intermittent timing issue when deleting multiple project properties in quick succession.
	- This is still an issue caused by DependencyTrack API, but is no longer affecting pipeline.

#### DEPENDENCIES
- `crazy-max/ghaction-import-gpg` `6.2.0` -> `6.3.0`
- `goreleaser/goreleaser-action` `6.2.1` -> `6.3.0`
- `golangci/golangci-lint-action` `6.5.2` -> `7.0.0`

## 1.8.1

#### MISC
- Document in `README.md`, supported versions of Terraform, and DependencyTrack.

#### FIXES
- Using `dependencytrack_config_properties`, `dependencytrack_config_property`, or `dependencytrack_project_property`, with a `type` of `"ENCRYPTEDSTRING"`, would result in the value being replaced by the placeholder value from DependencyTrack.
	- Now the current value is persisted in the statefile, across operations.
- Marked `description` in `dependencytrack_project_property` as `Computed` to account for it changing from `null` to `""`, when it is not provided.

#### ISSUES
- Deleting multiple `dependencytrack_project_property` on the same `dependencytrack_project` in quick succession can cause intermittent errors. This is caused by a delay within deleting within the DependencyTrack API.

## 1.8.0

#### FEATURES
- Added ability to manage several attributes of `dependencytrack_project` Resource - `version`, `parent`, `classifier`, `cpe`, `group`, `purl`, `swid`.
- Added several attributes to `dependencytrack_project` DataSource - `parent`, `classifier`, `cpe`, `group`, `purl`, `swid`

#### MISC
- Increase quality of testing for where two id's are expected to match, rather than just both being set.

#### FIXES
- Fixed an update to `dependencytrack_project` Resource from overriding existing settings of unmanaged properties, i.e. previosuly `parent` when change `name`.
	- Now retrieves the current settings, before updating - as unable to use a partial `PATCH` - due to inability to unset optional fields, e.g. `parent`.

## 1.7.1

#### DEPENDENCIES
- `github.com/hashicorp/teraform-plugin-testing` `1.11.0` -> `1.12.0`
- `actions/setup-go` `5.3.0` -> `5.4.0`
- `actions/download-artifact` `4.1.9` -> `4.2.0`
- `golangci/golangci-lint-action` `6.5.1` -> `6.5.2`
- `actions/setup-node` `4.2.0` -> `4.3.0` for `CDKTF`
- `github.com/golang-jwt/jwt/v4` `4.5.1` -> `4.5.2` in `/tools`

## 1.7.0

#### FEATURES
- Added `root_ca` option to `dependencytrack` Provider, to allow for setting a custom certificate for API TLS verification, defaulting to system certificates.
- Added `mtls` option to `dependencytrack` Provider, to allow for configuring client side TLS, which when `host` is using `https` results in `mTLS`.

#### MISC
- Added `nginx` instance to pipeline tests to test the different combinations of `root_ca` and `mtls` options on Provider.
- Added bash flags to git hooks and scripts, to increase error checking.
- Increased `go` version in `go.mod` from `1.22.7` -> `1.23.0`
- Introduced `toolchain` requirement in `go.mod` of `1.24.1`

#### DEPENDENCIES
- `golangci/golangci-lint-action` `6.5.0` -> `6.5.1`
- `golang.org/x/net` `0.33.0` -> `0.36.0` in `/tools`
- `golang.org/x/net` `0.34.0` -> `0.36.0`

## 1.6.0

#### FEATURES
- `dependencytrack_oidc_group` Resource, to manage an OIDC Group.
- `dependencytrack_oidc_group_mapping` Resource, to manage a mapping from an OIDC Group to a Team.

#### MISC
- Added examples for `dependencytrack_repository` due to being absent within `1.5.0` release.

#### DEPENDENCIES
- `golangci/golangci-lint-action` `6.4.0` -> `6.5.0`

## 1.5.0

#### FEATURES
- `dependencytrack_repository` Resource, to manage an external source repository.
- Added HTTP interception patching for select API requests, for which the SDK does not provide a working function.

#### MISC
- Added automated testing against Terraform `1.11.x`.
- Swapped `golangci-lint` rules from `enable` specific set, to `enable-all`, with specific `disable` to increase range of linters used.
- Reviewed linting rules, actioning, or identifying where not actioning.
- Added named import for `github.com/DependencyTrack/client-go` to resolve typecheck errors due to updated golang version.
- Removed secondary `Get` request when updating `dependencytrack_project`, instead using the return type of `Update` function.
- Added property tests for configuring properties within `dependencytrack_project`.

#### FIXES
- Fixed inability to delete a `dependencytrack_project_property`, as raised in `1.1.0`.
- Marked `type` as requiring replace when updated within `dependencytrack_project_property`.
- Fixed `active` not defaulting to `true` within `dependencytrack_project`.

#### DEPENDENCIES
- `actions/download-artifact` `4.1.8` -> `4.1.9`
- `github.com/hashicorp/terraform-plugin-framework` `v1.13.0` -> `v1.14.1`

## 1.4.0

#### FEATURES
- `dependencytrack_config_property` Resource, to manage a config property.
- `dependencytrack_config_property` DataSource, to retrieve a config property.
- `dependencytrack_config_properties` Resource, to manage multiple config properties more efficiently.

#### ISSUES
- [Fixed in `1.8.1`] Resource `dependencytrack_config_property` does not retain `value` when `type` is `"ENCRYPTEDSTRING"`.
- [Fixed in `1.8.1`] `properties` on `dependencytrack_config_properties` Resource does not retain `value` when `type` is `"ENCRYPTEDSTRING"`.

#### MISC
- Added automated testing against Terraform `1.10.x`.
- Disabled CDKTF binding generation, while it is not fully featured.
- Removed workflow to mark inactive issues as resolved.

#### DEPENDENCIES
- `golangci/golangci-lint-action` `6.3.3` -> `6.4.0`

## 1.3.3

#### DEPENDENCIES
- `github.com/DependencyTrack/client-go` `0.14.0` -> `0.15.0`
- `golang/golangci-lint-action` `6.3.0` -> `6.3.3`
- `goreleaser/goreleaser-action` `6.1.0` -> `6.2.1`

## 1.3.2

#### DEPENDENCIES
- `actions/setup-go` `5.2.0` -> `5.3.0`
- `github.com/hashicorp/terraform-plugin-go` `0.25.0` -> `0.26.0`
- `actions/setup-node` `4.1.0` -> `4.2.0`
- `golangci/golangci-lint-action` `6.2.0` -> `6.3.0`

## 1.3.1

#### DEPENDENCIES
- `golang.org/x/net` `0.28.0` -> `0.33.0`
- `golangci/golangci-lint-action` `6.1.1` -> `6.2.0`

##### /tools
- `golang.org.net` `0.23.0` -> `0.33.0`

## 1.3.0

#### FEATURES
- `dependencytrack_team` Resource, to manage a team.
- `dependencytrack_team` DataSource, to retrieve a team.
- `dependencytrack_team_apikey` Resource, to manage an API Key for a team.
- `dependencytrack_team_permission` Resource, to manage the permissions of a team.

#### DEPENDENCIES
- `DependencyTrack/client-go` `v0.13.0` -> `v0.14.0`
- `hashicorp/terraform-plugin-testing` `v1.10.0` -> `v1.11.0`

## 1.2.0

#### FEATURES
- `dependencytrack` Provider - Added options for setting additional custom headers.

## 1.1.0

#### FEATURES
- `dependencytrack_project_property` Resource, to manage a project property.
- `dependencytrack_project_property` DataSource, to retrieve a singular property.

#### ISSUES
- [Fixed in `1.5.0`] Unable to delete project property within DependencyTrack, when using `dependencytrack_project_property` resource.
- [Fixed in `1.5.0`] Updating `type` on `dependencytrack_project_property` does not recreate the resource, which is required to change the `type`.
- [Fixed in `1.8.1`] Resource `dependencytrack_project_property` does not retain `value` when `type` is `"ENCRYPTEDSTRING"`.

## 1.0.0

#### FEATURES
- Provider authentication via API Key, optionally reading from environment variable.
- `dependencytrack_project` Resource, for Projects, able to set minimal functionality.
- `dependencytrack_project` DataSource, to identify from a Project name and version, able to access properties.

#### ISSUES
- [Fixed in `1.5.0`] `dependencytrack_project.active` does not default to `true`.
- [Fixed in `1.8.0`] `dependencytrack_project` overrides non-managed properties on resources, when updating
