package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "dependencytrack" {
	host = "http://localhost:8081"
	token = "odt_dcqVqQWFy84PAxWfpEQBTItkEAMWeeoG"
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dependencytrack": providerserver.NewProtocol6WithError(New("test")()),
	}
)
