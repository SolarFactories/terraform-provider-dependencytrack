package provider

import (
	"context"
	"os"

	"github.com/DependencyTrack/client-go"
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
	Host    types.String `tfsdk:"host"`
	Key     types.String `tfsdk:"key"`
	Headers types.List   `tfsdk:"headers"`
}

type dependencyTrackProviderHeadersModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (p *dependencyTrackProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dependencytrack"
	resp.Version = p.version
}

func (p *dependencyTrackProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
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
			"Unable to create DependencyTrack Client, due to missing API host configuration.",
		)
	}
	if config.Key.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("key"),
			"Unknown DependencyTrack API Key",
			"Unable to create DependencyTrack Client, due to missing API key configuration.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch values, and perform value validation
	host := config.Host.ValueString()
	key := config.Key.ValueString()

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
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DependencyTrack client")
	client, err := dtrack.NewClient(host, dtrack.WithAPIKey(key))
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

func (p *dependencyTrackProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewProjectPropertyResource,
	}
}

func (p *dependencyTrackProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewProjectPropertyDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dependencyTrackProvider{
			version: version,
		}
	}
}
