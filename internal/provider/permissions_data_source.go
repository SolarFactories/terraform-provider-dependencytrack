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
	_ datasource.DataSource              = &permissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &permissionsDataSource{}
)

type (
	permissionsDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	permissionsDataSourceModel struct {
		Permissions []permissionsDataSourceModelPermission `tfsdk:"permissions"`
	}

	permissionsDataSourceModelPermission struct {
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
	}
)

func NewPermissionsDataSource() datasource.DataSource {
	return &permissionsDataSource{}
}

func (*permissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

func (*permissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch available permissions.",
		Attributes: map[string]schema.Attribute{
			"permissions": schema.ListNestedAttribute{
				Description: "List of available permissions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the permission.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the permission.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *permissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state permissionsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Permissions")
	permissions, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Permission], error) {
		return d.client.Permission.GetAll(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch Permissions.",
			"Unexpected error within: "+err.Error(),
		)
		return
	}
	newState := permissionsDataSourceModel{
		Permissions: Map(permissions, func(permission dtrack.Permission) permissionsDataSourceModelPermission {
			return permissionsDataSourceModelPermission{
				Name:        types.StringValue(permission.Name),
				Description: types.StringValue(permission.Description),
			}
		}),
	}
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Permissions", map[string]any{
		"permissions": newState.Permissions,
	})
}

func (d *permissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
