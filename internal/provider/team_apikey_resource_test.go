package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTeamApiKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_apikey" "test" {
	team = dependencytrack_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_apikey.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "key"),
					resource.TestCheckResourceAttr("dependencytrack_team_apikey.test", "comment", ""),
					resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "masked"),
					// TODO: Re-enable once able to conditionally check these for API Versions
					//resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "public_id"),
					//resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "legacy"),.
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_team_apikey.test",
				ImportState:       true,
				ImportStateVerify: true,
				// TODO: Ignore in only 4.13+ API versions.
				ImportStateVerifyIgnore: []string{"key"},
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_apikey" "test" {
	team = dependencytrack_team.test.id
	comment = "Sample comment"
}

`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_apikey.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "key"),
					resource.TestCheckResourceAttr("dependencytrack_team_apikey.test", "comment", "Sample comment"),
					resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "masked"),
					// TODO: Re-enable once able to conditionally check these for API Versions
					//resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "public_id"),
					//resource.TestCheckResourceAttrSet("dependencytrack_team_apikey.test", "legacy"),.
				),
			},
		},
	})
}

func TestAccAPIKeyResourceCommentRegression72(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/72
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_apikey" "test" {
	team = dependencytrack_team.test.id
	comment = "Sample Creation Comment"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_team_apikey.test", "comment", "Sample Creation Comment"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_team_apikey.test",
				ImportState:       true,
				ImportStateVerify: true,
				// TODO: Ignore in only 4.13+ API versions.
				ImportStateVerifyIgnore: []string{"key"},
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_apikey" "test" {
	team = dependencytrack_team.test.id
	comment = "Sample Update Comment"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_team_apikey.test", "comment", "Sample Update Comment"),
				),
			},
		},
	})
}

func TestAccAPIKeyResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial API Key.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_API_Key_236"
}
resource "dependencytrack_team_apikey" "test" {
	team = dependencytrack_team.test.id
}
`,
			},
			// Duplicate reference to API Key.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_API_Key_236"
}
resource "dependencytrack_team_apikey" "test" {
	team = dependencytrack_team.test.id
}
resource "dependencytrack_team_apikey" "test2" {
	team = dependencytrack_team.test.id
}
import {
	to = dependencytrack_team_apikey.test2
	id = dependencytrack_team_apikey.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_apikey.test", "id",
						"dependencytrack_team_apikey.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_API_Key_236"
}
resource "dependencytrack_team_apikey" "test2" {
	team = dependencytrack_team.test.id
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_team_apikey.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
