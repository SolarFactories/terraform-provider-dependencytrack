package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecretMetadataDataSource(t *testing.T) {
	if apiSemver.Lt(NewSemver(5, 0, 0)) {
		t.SkipNow()
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_secret" "test" {
	name = "Test_Secret_Metadata_Secret"
	value = "Test Value"
}
data "dependencytrack_secret_metadata" "test" {
	name = dependencytrack_secret.test.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_secret.test", "name",
						"data.dependencytrack_secret_metadata.test", "name",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_secret.test", "description",
						"data.dependencytrack_secret_metadata.test", "description",
					),
					resource.TestCheckResourceAttrSet("data.dependencytrack_secret_metadata.test", "created_at"),
					resource.TestCheckNoResourceAttr("data.dependencytrack_secret_metadata.test", "updated_at"),
				),
			},
			{
				Config: providerConfig + `
resource "dependencytrack_secret" "test" {
	name = "Test_Secret_Metadata_Secret"
	value = "Test Value"
	description = "Test_Description"
}
data "dependencytrack_secret_metadata" "test" {
	name = dependencytrack_secret.test.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_secret.test", "description",
						"data.dependencytrack_secret_metadata.test", "description",
					),
					resource.TestCheckResourceAttrSet("data.dependencytrack_secret_metadata.test", "updated_at"),
				),
			},
		},
	})
}
