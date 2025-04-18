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
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client *dtrack.Client
	semver *Semver
}

type projectDataSourceModel struct {
	Name       types.String             `tfsdk:"name"`
	Version    types.String             `tfsdk:"version"`
	ID         types.String             `tfsdk:"id"`
	Properties []projectPropertiesModel `tfsdk:"properties"`
	Classifier types.String             `tfsdk:"classifier"`
	CPE        types.String             `tfsdk:"cpe"`
	Group      types.String             `tfsdk:"group"`
	Parent     types.String             `tfsdk:"parent"`
	PURL       types.String             `tfsdk:"purl"`
	SWID       types.String             `tfsdk:"swid"`
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
		Description: "Fetch an existing Project by name and version. Requires the project to have a version defined on DependencyTrack.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the project to find.",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the project to find.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "UUID of the project located.",
				Computed:    true,
			},
			"properties": schema.ListNestedAttribute{
				Description: "Existing properties within the Project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							Description: "Group Name for the project Property.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Property Name for the project Property.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Property Value for the project Property.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Property Type for the project Property as a string enum.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description for the project Property.",
							Computed:    true,
						},
					},
				},
			},
			"classifier": schema.StringAttribute{
				Description: "Classifier of the Project. See DependencyTrack for possible enum values.",
				Computed:    true,
			},
			"cpe": schema.StringAttribute{
				Description: "Common Platform Enumeration for the Project. Standardised format v2.2 / v2.3 from MITRE / NIST",
				Computed:    true,
			},
			"group": schema.StringAttribute{
				Description: "Namespace / group / vendor of the Project.",
				Computed:    true,
			},
			"parent": schema.StringAttribute{
				Description: "UUID of a parent project, if nested.",
				Computed:    true,
				Optional:    true,
			},
			"purl": schema.StringAttribute{
				Description: "Package URL of the Project. Follows standardised format.",
				Computed:    true,
			},
			"swid": schema.StringAttribute{
				Description: "SWID Tag ID. ISO/IEC 19770-2:2015",
				Computed:    true,
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
		Classifier: types.StringValue(project.Classifier),
		CPE:        types.StringValue(project.CPE),
		Group:      types.StringValue(project.Group),
		PURL:       types.StringValue(project.PURL),
		SWID:       types.StringValue(project.SWIDTagID),
		Parent:     types.StringNull(),
	}
	if project.ParentRef != nil {
		projectState.Parent = types.StringValue(project.ParentRef.UUID.String())
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
	diags = resp.State.Set(ctx, &projectState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read DependencyTrack project", map[string]any{"uuid": project.UUID.String()})
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
