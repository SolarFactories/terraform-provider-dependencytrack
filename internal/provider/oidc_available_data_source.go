package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface impl check.
var (
	_ datasource.DataSource              = &oidcAvailableDataSource{}
	_ datasource.DataSourceWithConfigure = &oidcAvailableDataSource{}
)

type (
	oidcAvailableDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	oidcAvailableDataSourceModel struct {
		Available types.Bool `tfsdk:"available"`
	}
)

func NewOidcAvailableDataSource() datasource.DataSource {
	return &oidcAvailableDataSource{}
}

func (*oidcAvailableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_available"
}

func (*oidcAvailableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch whether OIDC is available.",
		Attributes: map[string]schema.Attribute{
			"available": schema.BoolAttribute{
				Description: "Whether OIDC is available.",
				Computed:    true,
			},
		},
	}
}

func (d *oidcAvailableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oidcAvailableDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading OIDC Availability")
	available, err := d.client.OIDC.Available(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch OIDC availability.",
			"Unexpected error within: "+err.Error(),
		)
		return
	}
	newState := oidcAvailableDataSourceModel{
		Available: types.BoolValue(available),
	}
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read OIDC Availability", map[string]any{
		"available": newState.Available.ValueBool(),
	})
}

func (d *oidcAvailableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInfoData, ok := req.ProviderData.(clientInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected provider.clientInfo, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = clientInfoData.client
	d.semver = clientInfoData.semver
}
