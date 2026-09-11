package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccRepositoryResource(t *testing.T) {
	if apiSemver.Major > 4 {
		t.Skip("TODO: password field has changed from value, to being a name of a managed secret. Will need Secrets Management resource, and splitting into two tests.")
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test" {
	type = "GITHUB"
	identifier = "Test_Repository"
	url = "https://localhost"
	precedence = 2
	enabled = true
	internal = false
	username = "Test_Username"
	password = "Test_Password"
}
`,

				Check: resource.ComposeAggregateTestCheckFunc(
					// Authenticated.
					resource.TestCheckResourceAttrSet("dependencytrack_repository.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "type", "GITHUB"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "identifier", "Test_Repository"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "url", "https://localhost"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "precedence", "2"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "internal", "false"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "password", "Test_Password"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test" {
	type = "GITHUB"
	identifier = "Test_Repository_With_Change"
	url = "https://localhost"
	precedence = 2
	enabled = true
	internal = false
	username = "Test_Username"
	password = "Test_Password_With_Change"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_repository.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "type", "GITHUB"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "identifier", "Test_Repository_With_Change"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "url", "https://localhost"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "precedence", "2"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "internal", "false"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "password", "Test_Password_With_Change"),
				),
			},
		},
	})
}

func TestAccRepositoryResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Repository.
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test" {
	type = "GITHUB"
	identifier = "Test_Repository_236"
	url = "https://localhost"
	internal = false
	enabled = true
	username = ""
	password = ""
}
`,
			},
			// Duplicate reference to the Repository.
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test" {
	type = "GITHUB"
	identifier = "Test_Repository_236"
	url = "https://localhost"
	internal = false
	enabled = true
	username = ""
	password = ""
}
resource "dependencytrack_repository" "test2" {
	type = "GITHUB"
	identifier = "Test_Repository_236"
	url = "https://localhost"
	internal = false
	enabled = true
	username = ""
	password = ""
}
import {
	to = dependencytrack_repository.test2
	id = dependencytrack_repository.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_repository.test", "id",
						"dependencytrack_repository.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test2" {
	type = "GITHUB"
	identifier = "Test_Repository_236"
	url = "https://localhost"
	internal = false
	enabled = true
	username = ""
	password = ""
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_repository.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
