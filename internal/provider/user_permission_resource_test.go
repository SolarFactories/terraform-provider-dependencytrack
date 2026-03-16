package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserPermissionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_user" "test" {
	username = "Test_User"
	fullname = "Test User"
	email = "test_user@example.com"
	password = "Test_Password"
}
resource "dependencytrack_user_permission" "test" {
	username = dependencytrack_user.test.username
	permission = "SYSTEM_CONFIGURATION"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_permission.test", "username",
						"dependencytrack_user.test", "username",
					),
					resource.TestCheckResourceAttr("dependencytrack_user_permission.test", "permission", "SYSTEM_CONFIGURATION"),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_user" "test" {
	username = "Test_User"
	fullname = "Test User"
	email = "test_user@example.com"
	password = "Test_Password"
}
resource "dependencytrack_user_permission" "test" {
	username = dependencytrack_user.test.username
	permission = "SYSTEM_CONFIGURATION"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_permission.test", "username",
						"dependencytrack_user.test", "username",
					),
					resource.TestCheckResourceAttr("dependencytrack_user_permission.test", "permission", "SYSTEM_CONFIGURATION"),
				),
			},
		},
	})
}

func TestAccUserPermissionResourceSSORegression177(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/177
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test" {
	username = "Test_User_Regression177_OIDC"
}
resource "dependencytrack_ldap_user" "test" {
	username = "Test_User_Regression177_LDAP"
}

resource "dependencytrack_user_permission" "oidc" {
	username = dependencytrack_oidc_user.test.username
	permission = "SYSTEM_CONFIGURATION"
}
resource "dependencytrack_user_permission" "ldap" {
	username = dependencytrack_ldap_user.test.username
	permission = "SYSTEM_CONFIGURATION"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_permission.oidc", "username",
						"dependencytrack_oidc_user.test", "username",
					),
					resource.TestCheckResourceAttr(
						"dependencytrack_user_permission.oidc", "permission",
						"SYSTEM_CONFIGURATION",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_permission.ldap", "username",
						"dependencytrack_ldap_user.test", "username",
					),
					resource.TestCheckResourceAttr(
						"dependencytrack_user_permission.ldap", "permission",
						"SYSTEM_CONFIGURATION",
					),
				),
			},
		},
	})
}
