package provider

import (
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	providerConfig = func() string {
		option := os.Getenv("DEPENDENCYTRACK_TEST_PROVIDER")
		locals := getLocals(option)

		return locals + `
		provider "dependencytrack" {
			host = local.provider_host
			key = local.provider_key
			auth = local.provider_auth
			root_ca = local.provider_root_ca
			mtls = local.provider_mtls
		}`
	}()

	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dependencytrack": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func getLocals(option string) string {
	locals := "locals {\n"
	switch option {
	case "rootCA":
		{
			locals += "\tprovider_host = \"https://localhost:8082\"\n"
			locals += "\tprovider_key = \"OS_ENV\"\n"
			locals += "\tprovider_auth = null\n"
		}
	case "mtls":
		{
			locals += "\tprovider_host = \"http://localhost:8083\"\n"
			locals += "\tprovider_auth = { type = \"KEY\" key = \"OS_ENV\" }\n"
			locals += "\tprovider_key = null\n"
		}
	case "rootCA+mtls":
		{
			locals += "\tprovider_host = \"https://localhost:8084\"\n"
			locals += "\tprovider_auth = { type = \"KEY\" key = \"OS_ENV\" }\n"
			locals += "\tprovider_key = null\n"
		}
	default:
		{
			locals += "\tprovider_host = \"http://localhost:8081\"\n"
			locals += "\tprovider_key = \"OS_ENV\"\n"
			locals += "\tprovider_auth = null\n"
		}
	}
	if strings.Contains(option, "rootCA") {
		rootCa, err := os.ReadFile("/opt/server_cert.pem")
		if err != nil {
			panic("Root CA file is unable to be read: " + err.Error())
		}
		locals += "\tprovider_root_ca = \"" + strings.ReplaceAll(string(rootCa), "\n", "\\n") + "\"\n"
	} else {
		locals += "\tprovider_root_ca = null\n"
	}
	if strings.Contains(option, "mtls") {
		locals += "\tprovider_mtls = { key_path = \"/opt/client_key.pem\" cert_path = \"/opt/client_cert.pem\" }\n"
	} else {
		locals += "\tprovider_mtls = null\n"
	}

	locals += "}"
	return locals
}
