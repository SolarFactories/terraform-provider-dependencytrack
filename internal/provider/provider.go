package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Host types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// TODO: Replace with introduction of an SDK
type DependencyTrackClient struct {
	Host string
	Token string
}

func (p *dependencyTrackProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dependencytrack"
	resp.Version = p.version
}

func (p *dependencyTrackProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema {
		Attributes: map[string]schema.Attribute {
			"host": schema.StringAttribute {
				Required: true,
				Optional: false,
			},
			"token": schema.StringAttribute {
				Required: true,
				Optional: false,
				Sensitive: true,
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

	// Check for unspecified values
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown DependencyTrack Host",
			"Unable to create DependencyTrack Client, due to missing host configuration.",
		)
	}
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown DependencyTrack Token",
			"Unable to create DependencyTrack Client, due to missing API token configuration.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch values, and perform value validation
	host := config.Host.ValueString()
	token := config.Host.ValueString()

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing DependencyTrack Host",
			"Host for DependencyTrack was provided, but it was empty.",
		)
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing DependencyTrack Token",
			"Token for DependencyTrack was provided, but it was empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.DataSourceData = DependencyTrackClient { Host: host, Token: token }
	resp.ResourceData = DependencyTrackClient { Host: host, Token: token }
}

func (p *dependencyTrackProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *dependencyTrackProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource {
		NewProjectDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dependencyTrackProvider{
			version: version,
		}
	}
}
