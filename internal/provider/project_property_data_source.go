package provider

import (
	"context"
	"fmt"

	"github.com/DependencyTrack/client-go"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface impl check.
var (
	_ datasource.DataSource              = &projectPropertyDataSource{}
	_ datasource.DataSourceWithConfigure = &projectPropertyDataSource{}
)

func NewProjectPropertyDataSource() datasource.DataSource {
	return &projectPropertyDataSource{}
}

type projectPropertyDataSource struct {
	client *dtrack.Client
}

type projectPropertyDataSourceModel struct {
	Project     types.String `tfsdk:"project"`
	Group       types.String `tfsdk:"group"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (d *projectPropertyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_property"
}

func (d *projectPropertyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch a project property by group and name.",
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Description: "UUID of the Project which contains the property.",
				Required:    true,
			},
			"group": schema.StringAttribute{
				Description: "Group Name of the property on the project.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Property Name of the property on the project.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Property Value of the property on the project.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Property Type of the property on the project. See DependencyTrack for valid enum values.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the property on the project.",
				Computed:    true,
			},
		},
	}
}

func (d *projectPropertyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectPropertyDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading DependencyTrack Project Property", map[string]any{
		"project": state.Project.ValueString(),
		"group":   state.Group.ValueString(),
		"name":    state.Name.ValueString(),
	})
	uuid, err := uuid.Parse(state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Read, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	properties, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.ProjectProperty], error) {
		return d.client.ProjectProperty.GetAll(ctx, uuid, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch all project properties",
			"Error from: "+err.Error(),
		)
	}

	filtered := []dtrack.ProjectProperty{}
	for _, property := range properties {
		if property.Group != state.Group.ValueString() {
			continue
		}
		if property.Name != state.Name.ValueString() {
			continue
		}
		filtered = append(filtered, property)
	}
	if len(filtered) == 0 {
		resp.Diagnostics.AddError(
			"Within Read, unable to locate property.",
			"No such property on project",
		)
		return
	} else if len(filtered) > 1 {
		resp.Diagnostics.AddError(
			"Within Read, found multiple matching properties.",
			"This is supposed to be an impossible situation.",
		)
		return
	}
	property := filtered[0]
	propertyState := projectPropertyDataSourceModel{
		Project:     types.StringValue(uuid.String()),
		Group:       types.StringValue(property.Group),
		Name:        types.StringValue(property.Name),
		Value:       types.StringValue(property.Value),
		Type:        types.StringValue(property.Type),
		Description: types.StringValue(property.Description),
	}
	diags = resp.State.Set(ctx, &propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Read DependencyTrack ProjectProperty")
}

func (d *projectPropertyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*dtrack.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected *dtrack.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}