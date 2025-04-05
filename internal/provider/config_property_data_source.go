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
	_ datasource.DataSource              = &configPropertyDataSource{}
	_ datasource.DataSourceWithConfigure = &configPropertyDataSource{}
)

func NewConfigPropertyDataSource() datasource.DataSource {
	return &configPropertyDataSource{}
}

type configPropertyDataSource struct {
	client *dtrack.Client
	semver *Semver
}

type configPropertyDataSourceModel struct {
	Group       types.String `tfsdk:"group"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (d *configPropertyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_property"
}

func (d *configPropertyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch a config property by group and name.",
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				Description: "Group Name of the config property.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Property Name of the config property.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Property Value of the config property.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Property Type of the config property. See DependencyTrack for valid enum values.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the config property.",
				Computed:    true,
			},
		},
	}
}

func (d *configPropertyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state configPropertyDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := state.Group.ValueString()
	propertyName := state.Name.ValueString()

	tflog.Debug(ctx, "Reading DependencyTrack Config Property", map[string]any{
		"group": groupName,
		"name":  propertyName,
	})
	configProperty, err := d.client.Config.Get(ctx, groupName, propertyName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch config properties.",
			"Unexpected error within: "+err.Error(),
		)
		return
	}
	propertyState := configPropertyDataSourceModel{
		Group:       types.StringValue(configProperty.GroupName),
		Name:        types.StringValue(configProperty.Name),
		Value:       types.StringValue(configProperty.Value),
		Type:        types.StringValue(configProperty.Type),
		Description: types.StringValue(configProperty.Description),
	}
	diags = resp.State.Set(ctx, &propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Read DependencyTrack ConfigProperty")
}

func (d *configPropertyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInfo, ok := req.ProviderData.(clientInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected provider.clientInfo, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = clientInfo.client
	d.semver = clientInfo.semver
}
