package provider

import (
	"context"
	"fmt"

	"github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface impl check.
var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client *dtrack.Client
}

type projectDataSourceModel struct {
	Name       types.String             `tfsdk:"name"`
	Version    types.String             `tfsdk:"version"`
	ID         types.String             `tfsdk:"id"`
	Properties []projectPropertiesModel `tfsdk:"properties"`
}

type projectPropertiesModel struct {
	Group       types.String `tfsdk:"group"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"version": schema.StringAttribute{
				Optional: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"properties": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"value": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading DependencyTrack project", map[string]any{"name": state.Name.ValueString(), "version": state.Version.ValueString()})
	project, err := d.client.Project.Lookup(ctx, state.Name.ValueString(), state.Version.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read project from DependencyTrack",
			"Eror located within SDK Client: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Found project with UUID: "+project.UUID.String())
	// Transform data into model
	projectState := projectDataSourceModel{
		Name:       types.StringValue(project.Name),
		Version:    types.StringValue(project.Version),
		ID:         types.StringValue(project.UUID.String()),
		Properties: make([]projectPropertiesModel, 0),
	}
	for _, property := range project.Properties {
		tflog.Debug(ctx, "Found property with group"+property.Group)
		projectState.Properties = append(projectState.Properties, projectPropertiesModel{
			Group:       types.StringValue(property.Group),
			Name:        types.StringValue(property.Name),
			Value:       types.StringValue(property.Value),
			Type:        types.StringValue(property.Type),
			Description: types.StringValue(property.Description),
		})
	}
	// Update state
	diags2 := resp.State.Set(ctx, &projectState)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Read DependencyTrack project", map[string]any{"uuid": project.UUID.String()})
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*dtrack.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *dtrack.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}
