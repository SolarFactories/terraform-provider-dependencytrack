package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccLDAPUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_ldap_user" "test" {
	username = "Test_Username"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_ldap_user.test", "id", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_ldap_user.test", "username", "Test_Username"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_ldap_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_ldap_user" "test" {
	username = "Test_Username"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_ldap_user.test", "id", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_ldap_user.test", "username", "Test_Username"),
				),
			},
		},
	})
}

func TestAccLDAPUserResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial LDAP User.
			{
				Config: providerConfig + `
resource "dependencytrack_ldap_user" "test" {
	username = "Test_LDAP_User_236"
}
`,
			},
			// Duplicate reference to LDAP User.
			{
				Config: providerConfig + `
resource "dependencytrack_ldap_user" "test" {
	username = "Test_LDAP_User_236"
}
resource "dependencytrack_ldap_user" "test2" {
	username = "Test_LDAP_User_236"
}
import {
	to = dependencytrack_ldap_user.test2
	id = dependencytrack_ldap_user.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_ldap_user.test", "id",
						"dependencytrack_ldap_user.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_ldap_user" "test2" {
	username = "Test_LDAP_User_236"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_ldap_user.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
