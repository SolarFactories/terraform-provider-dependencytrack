package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecretResource(t *testing.T) {
	if apiSemver.Lt(NewSemver(5, 0, 0)) {
		t.SkipNow()
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_secret" "test" {
	name = "Test_Secret_Name"
	value = "Test_Secret_Value"
}
resource "dependencytrack_secret" "test2" {
	name = "Test_Secret_Name_2"
	value = "Test_Secret_Value_2"
	description = "Test_Secret_Description_2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_secret.test", "name", "Test_Secret_Name"),
					resource.TestCheckResourceAttr("dependencytrack_secret.test", "value", "Test_Secret_Value"),
					resource.TestCheckResourceAttrSet("dependencytrack_secret.test", "url"),
					resource.TestCheckNoResourceAttr("dependencytrack_secret.test", "description"),
					//
					resource.TestCheckResourceAttr("dependencytrack_secret.test2", "name", "Test_Secret_Name_2"),
					resource.TestCheckResourceAttr("dependencytrack_secret.test2", "value", "Test_Secret_Value_2"),
					resource.TestCheckResourceAttrSet("dependencytrack_secret.test2", "url"),
					resource.TestCheckResourceAttr("dependencytrack_secret.test2", "description", "Test_Secret_Description_2"),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_secret" "test" {
	name = "Test_Secret_Name"
	value = "Test_Secret_Value_With_Change"
}
resource "dependencytrack_secret" "test2" {
	name = "Test_Secret_Name_2"
	value = "Test_Secret_Value_2"
	description = "Test_Secret_Description_2_With_Change"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_secret.test", "name", "Test_Secret_Name"),
					resource.TestCheckResourceAttr("dependencytrack_secret.test", "value", "Test_Secret_Value_With_Change"),
					resource.TestCheckResourceAttrSet("dependencytrack_secret.test", "url"),
					resource.TestCheckNoResourceAttr("dependencytrack_secret.test", "description"),
					//
					resource.TestCheckResourceAttr("dependencytrack_secret.test2", "name", "Test_Secret_Name_2"),
					resource.TestCheckResourceAttr("dependencytrack_secret.test2", "value", "Test_Secret_Value_2"),
					resource.TestCheckResourceAttrSet("dependencytrack_secret.test2", "url"),
					resource.TestCheckResourceAttr("dependencytrack_secret.test2", "description", "Test_Secret_Description_2_With_Change"),
				),
			},
		},
	})
}
