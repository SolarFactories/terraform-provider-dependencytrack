package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccManagedUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_managed_user" "test" {
	username = "Test_Username"
	fullname = "Test_Fullname"
	email = "Test_Email@example.com"
	password = "Test_Password"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_managed_user.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "fullname", "Test_Fullname"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "email", "Test_Email@example.com"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "password", "Test_Password"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "suspended", "false"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "force_password_change", "false"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "password_expires", "false"),
				),
			},
			// ImportState testing.
			{
				ResourceName:            "dependencytrack_managed_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_managed_user" "test" {
	username = "Test_Username"
	fullname = "Test_Fullname_With_Change"
	email = "Test_Email@example.com"
	password_expires = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_managed_user.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "fullname", "Test_Fullname_With_Change"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "email", "Test_Email@example.com"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "password", "Test_Password"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "suspended", "false"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "force_password_change", "false"),
					resource.TestCheckResourceAttr("dependencytrack_managed_user.test", "password_expires", "true"),
				),
			},
		},
	})
}
