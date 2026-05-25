package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserSelfDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_user" "user" {
	username = "Test_User_Self"
	fullname = "Test User"
	email = "test@example.com"
	force_password_change = false
	password = "PASSWORD"
}

provider "dependencytrack" {
	host = "http://localhost:8081"
	auth = {
		type = "NONE"
	}
	alias = "unauthenticated"
}

data "dependencytrack_user_login" "test" {
	username = dependencytrack_user.user.username
	password = dependencytrack_user.user.password
	provider = dependencytrack.unauthenticated
}

provider "dependencytrack" {
	host = "http://localhost:8081"
	auth = {
		type = "BEARER"
		bearer = data.dependencytrack_user_login.test.token
	}
	alias = "managed"
}

data "dependencytrack_user_self" "test" {
	provider = dependencytrack.managed
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_user_login.test", "token"),
					resource.TestCheckResourceAttr("data.dependencytrack_user_login.test", "username", "admin"),
					resource.TestCheckResourceAttr("data.dependencytrack_user_login.test", "password", ""),
				),
			},
		},
	})
}
