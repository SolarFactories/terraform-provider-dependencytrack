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
