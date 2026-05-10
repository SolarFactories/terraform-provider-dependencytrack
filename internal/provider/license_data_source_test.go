package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLicenseDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_license" "test" {
	id = "MIT"
}
data "dependencytrack_license" "afl" {
	id = "AFL-1.1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dependencytrack_license.test", "id", "MIT"),
					resource.TestCheckResourceAttr("data.dependencytrack_license.test", "name", "MIT License"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license.test", "uuid"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license.test", "text"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license.test", "template"),
					resource.TestCheckResourceAttr("data.dependencytrack_license.test", "osi_approved", "true"),
					resource.TestCheckResourceAttr("data.dependencytrack_license.test", "fsf_libre", "true"),
					resource.TestCheckResourceAttr("data.dependencytrack_license.test", "is_deprecated_license_id", "false"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license.test", "see_also.#"),
					// Element `see_also.1` is added for `MIT` in API 4.13.5.
					resource.TestCheckResourceAttr("data.dependencytrack_license.test", "see_also.0", "https://opensource.org/license/mit/"),
					// Comment, Header are not present on MIT, so using another license to verify the retrieval of values.
					resource.TestCheckResourceAttr("data.dependencytrack_license.afl", "comment", "This license has been superseded by later versions."),
					resource.TestCheckResourceAttr("data.dependencytrack_license.afl", "header", "\"Licensed under the Academic Free License version 1.1.\""),
				),
			},
		},
	})
}
