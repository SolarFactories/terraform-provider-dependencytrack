package provider

import (
	"context"
	"fmt"
	"os"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure satisfies various provider interfaces.
var (
	_ provider.Provider = &dependencyTrackProvider{}
)

type dependencyTrackProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type dependencyTrackProviderModel struct {
	Host    types.String                          `tfsdk:"host"`
	Key     types.String                          `tfsdk:"key"`
	Headers []dependencyTrackProviderHeadersModel `tfsdk:"headers"`
	RootCA  types.String                          `tfsdk:"root_ca"`
	MTLS    *dependencyTrackProviderMtlsModel     `tfsdk:"mtls"`
}

type dependencyTrackProviderHeadersModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type dependencyTrackProviderMtlsModel struct {
	KeyPath  types.String `tfsdk:"key_path"`
	CertPath types.String `tfsdk:"cert_path"`
}

func (p *dependencyTrackProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dependencytrack"
	resp.Version = p.version
}

func (p *dependencyTrackProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with DependencyTrack.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for DependencyTrack API.",
				Required:    true,
			},
			"key": schema.StringAttribute{
				Description: "API Key for authentication to DependencyTrack. Must have permissions for all attempted actions. Set to 'OS_ENV' to read from DEPENDENCYTRACK_API_KEY environment variable.",
				Required:    true,
				Sensitive:   true,
			},
			"headers": schema.ListNestedAttribute{
				Description: "Add additional headers to client API requests. Useful for proxy authentication.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the header to specify.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the header to specify.",
							Required:    true,
						},
					},
				},
			},
			"root_ca": schema.StringAttribute{
				Description: "Root CA Certificate(s) used for TLS connection to DependencyTrack API in PEM format.",
				Optional:    true,
			},
			"mtls": schema.SingleNestedAttribute{
				Description: "Client Key and Certificate paths to use for mTLS connection to DependencyTrack API.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"key_path": schema.StringAttribute{
						Description: "Path to the file containing the client key.",
						Required:    true,
					},
					"cert_path": schema.StringAttribute{
						Description: "Path to the file containing the client certificate.",
						Required:    true,
					},
				},
			},
		},
	}
}

func (p *dependencyTrackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Get provider data from config
	var config dependencyTrackProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	host := config.Host.ValueString()
	key := config.Key.ValueString()
	clientCertFile := ""
	clientKeyFile := ""
	headers := make([]Header, 0, len(config.Headers))
	rootCAs := config.RootCA.ValueString()

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing DependencyTrack Host",
			"Host for DependencyTrack was provided, but it was empty.",
		)
	}
	// If key is the magic value 'OS_ENV', load from environment variable
	if key == "OS_ENV" {
		key = os.Getenv("DEPENDENCYTRACK_API_KEY")
	}
	if key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("key"),
			"Missing DependencyTrack API Key",
			"API Key for DependencyTrack was provided, but it was empty.",
		)
	}
	// Headers
	for _, header := range config.Headers {
		name := header.Name.ValueString()
		value := header.Value.ValueString()
		if name == "" || value == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("headers"),
				"Missing header attributes",
				fmt.Sprintf("Found Header Name: %s, and Value: %s", name, value),
			)
			continue
		}
		headers = append(headers, Header{name, value})
	}
	// mTLS
	if config.MTLS != nil {
		clientCertFile = config.MTLS.CertPath.ValueString()
		clientKeyFile = config.MTLS.KeyPath.ValueString()
	}
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating DependencyTrack client")
	httpClient, err := NewHTTPClient(headers, []byte(rootCAs), clientCertFile, clientKeyFile)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create HTTP Client",
			"An unexpected error occurred when creating the HTTP Client in error: "+err.Error(),
		)
		return
	}
	client, err := dtrack.NewClient(host, dtrack.WithHttpClient(httpClient), dtrack.WithAPIKey(key))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create DependencyTrack API Client",
			"An Unexpected error occurred when creating the DependencyTrack API Client. "+err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "Configured DependencyTrack client", map[string]any{"success": true})
}

func (p *dependencyTrackProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewProjectPropertyResource,
		NewTeamResource,
		NewTeamPermissionResource,
		NewTeamAPIKeyResource,
		NewConfigPropertyResource,
		NewConfigPropertiesResource,
		NewRepositoryResource,
		NewOidcGroupResource,
		NewOidcGroupMappingResource,
	}
}

func (p *dependencyTrackProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewProjectPropertyDataSource,
		NewTeamDataSource,
		NewConfigPropertyDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dependencyTrackProvider{
			version: version,
		}
	}
}
