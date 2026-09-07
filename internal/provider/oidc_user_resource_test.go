package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccOIDCUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test" {
	username = "Test_Username"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "id", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "username", "Test_Username"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_oidc_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test" {
	username = "Test_Username"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "id", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "username", "Test_Username"),
				),
			},
		},
	})
}

func TestAccOIDCUserResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial OIDC User.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test" {
	username = "Test_OIDCUser_236"
}
`,
			},
			// Duplicate reference to OIDC User.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test" {
	username = "Test_OIDCUser_236"
}
resource "dependencytrack_oidc_user" "test2" {
	username = "Test_OIDCUser_236"
}
import {
	to = dependencytrack_oidc_user.test2
	id = dependencytrack_oidc_user.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_user.test", "id",
						"dependencytrack_oidc_user.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test2" {
	username = "Test_OIDCUser_236"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_oidc_user.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
