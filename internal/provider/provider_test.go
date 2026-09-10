package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	providerConfig = func() string {
		option := os.Getenv("DEPENDENCYTRACK_TEST_PROVIDER")
		if option == "rootCA" {
			rootCa, err := os.ReadFile("/opt/server_cert.pem")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			return `provider "dependencytrack" {
				host = "https://localhost:8082"
				key = "OS_ENV"
				root_ca = "` + strings.ReplaceAll(string(rootCa), "\n", "\\n") + `"
			}`
		}
		if option == "mtls" {
			return `provider "dependencytrack" {
				host = "http://localhost:8083"
				auth = {
					type = "KEY"
					key = "OS_ENV"
				}
				mtls = {
					key_path = "/opt/client_key.pem",
					cert_path = "/opt/client_cert.pem",
				}
			}`
		}
		if option == "rootCA+mtls" {
			rootCa, err := os.ReadFile("/opt/server_cert.pem")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			return `provider "dependencytrack" {
				host = "https://localhost:8084"
				auth = {
					type = "KEY"
					key = "OS_ENV"
				}
				root_ca = "` + strings.ReplaceAll(string(rootCa), "\n", "\\n") + `"
				mtls = {
					key_path = "/opt/client_key.pem",
					cert_path = "/opt/client_cert.pem",
				}
			}`
		}
		if option == "v5" {
			return `provider "dependencytrack" {
				host = "http://localhost:9081"
				key = "OS_ENV"
			}`
		}
		return `provider "dependencytrack" {
			host = "http://localhost:8081"
			key = "OS_ENV"
		}`
	}()

	getClientConfig = func(option string) dependencyTrackProviderModel {
		if option == "rootCA" {
			rootCa, err := os.ReadFile("/opt/server_cert.pem")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			rootCaStr := strings.ReplaceAll(string(rootCa), "\n", "\\n")
			return dependencyTrackProviderModel{
				Host:   types.StringValue("https://localhost:8082"),
				Key:    types.StringValue("OS_ENV"),
				RootCA: types.StringValue(rootCaStr),
			}
		}
		if option == "mtls" {
			return dependencyTrackProviderModel{
				Host: types.StringValue("http://localhost:8083"),
				Auth: &providerAuthModel{
					Type: types.StringValue("KEY"),
					Key:  types.StringValue("OS_ENV"),
				},
				MTLS: &dependencyTrackProviderMtlsModel{
					KeyPath:  types.StringValue("/opt/client_key.pem"),
					CertPath: types.StringValue("/opt/client_cert.pem"),
				},
			}
		}
		if option == "rootCA+mtls" {
			rootCa, err := os.ReadFile("/opt/server_cert.pem")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			rootCaStr := strings.ReplaceAll(string(rootCa), "\n", "\\n")
			return dependencyTrackProviderModel{
				Host: types.StringValue("https://localhost:8084"),
				Auth: &providerAuthModel{
					Type: types.StringValue("KEY"),
					Key:  types.StringValue("OS_ENV"),
				},
				RootCA: types.StringValue(rootCaStr),
				MTLS: &dependencyTrackProviderMtlsModel{
					KeyPath:  types.StringValue("/opt/client_key.pem"),
					CertPath: types.StringValue("/opt/client_cert.pem"),
				},
			}
		}
		if option == "v5" {
			return dependencyTrackProviderModel{
				Host: types.StringValue("http://localhost:9081"),
				Key:  types.StringValue("OS_ENV"),
			}
		}
		return dependencyTrackProviderModel{
			Host: types.StringValue("http://localhost:8081"),
			Key:  types.StringValue("OS_ENV"),
		}
	}

	getAPISemver = func(option string) Semver {
		config := getClientConfig(option)
		prov := dependencyTrackProvider{version: "test"}
		diags := diag.Diagnostics{}
		clientInfo := prov.configureImpl(context.Background(), config, &diags)
		if diags.HasError() {
			panic("Test Provider has diagnostic errors, when creating.")
		}
		if clientInfo == nil {
			panic("Test Provider creation returned nil clientInfo.")
		}
		return *clientInfo.semver
	}

	apiSemver = getAPISemver(os.Getenv("DEPENDENCYTRACK_TEST_PROVIDER"))

	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dependencytrack": providerserver.NewProtocol6WithError(New("test")()),
	}
)
