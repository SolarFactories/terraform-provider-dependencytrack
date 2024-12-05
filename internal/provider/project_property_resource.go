package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"strings"

	"github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &projectPropertyResource{}
	_ resource.ResourceWithConfigure   = &projectPropertyResource{}
	_ resource.ResourceWithImportState = &projectPropertyResource{}
)

func NewProjectPropertyResource() resource.Resource {
	return &projectPropertyResource{}
}

type projectPropertyResource struct {
	client *dtrack.Client
}

type projectPropertyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Project     types.String `tfsdk:"project"`
	Group       types.String `tfsdk:"group"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (r *projectPropertyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_property"
}

func (r *projectPropertyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Project Property.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID used by provider. Has no meaning to DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project": schema.StringAttribute{
				Description: "UUID for the Project in which to create the Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group": schema.StringAttribute{
				Description: "Group name of the Project Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Property name of the Project Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value of the Project Property.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the Project Property. See DependencyTrack for valid enum values.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Project Property.",
				Optional:    true,
			},
		},
	}
}

func (r *projectPropertyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectPropertyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := uuid.Parse(plan.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Create, Unable to parse Project UUID from value.",
			"From Error: "+err.Error(),
		)
		return
	}
	propertyReq := dtrack.ProjectProperty{
		Group:       plan.Group.ValueString(),
		Name:        plan.Name.ValueString(),
		Value:       plan.Value.ValueString(),
		Type:        plan.Type.ValueString(),
		Description: plan.Description.ValueString(),
	}

	tflog.Debug(ctx, "Creating a new project property", map[string]any{
		"group": propertyReq.Group,
		"name":  propertyReq.Name,
	})
	propertyRes, err := r.client.ProjectProperty.Create(ctx, project, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project property.",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", project.String(), propertyRes.Group, propertyRes.Name))
	plan.Project = types.StringValue(project.String())
	plan.Group = types.StringValue(propertyRes.Group)
	plan.Name = types.StringValue(propertyRes.Name)
	plan.Value = types.StringValue(propertyRes.Value)
	plan.Type = types.StringValue(propertyRes.Type)
	plan.Description = types.StringValue(propertyRes.Description)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a new project property")
}

func (r *projectPropertyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectPropertyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Refreshing project property")
	project, err := uuid.Parse(state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Read, unable to parse project into UUID.",
			"Error from: "+err.Error(),
		)
		return
	}
	property, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.ProjectProperty], error) {
			return r.client.ProjectProperty.GetAll(ctx, project, po)
		},
		func(property dtrack.ProjectProperty) bool {
			if property.Group != state.Group.ValueString() {
				return false
			}
			if property.Name != state.Name.ValueString() {
				return false
			}
			return true
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to locate project property.",
			"Unexpected error from: "+err.Error(),
		)
	}
	propertyState := projectPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s/%s", project.String(), property.Group, property.Name)),
		Project:     types.StringValue(project.String()),
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
	tflog.Debug(ctx, "Refreshed project property.")
}

func (r *projectPropertyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectPropertyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := uuid.Parse(plan.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Update, unable to parse project into UUID.",
			"Error from: "+err.Error(),
		)
		return
	}
	propertyReq := dtrack.ProjectProperty{
		Group:       plan.Group.ValueString(),
		Name:        plan.Name.ValueString(),
		Value:       plan.Value.ValueString(),
		Type:        plan.Type.ValueString(),
		Description: plan.Description.ValueString(),
	}

	tflog.Debug(ctx, "Updating project property")
	propertyRes, err := r.client.ProjectProperty.Update(ctx, project, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project property.",
			"Error from: "+err.Error(),
		)
		return
	}
	state := projectPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s/%s", project.String(), propertyRes.Group, propertyRes.Name)),
		Project:     types.StringValue(project.String()),
		Group:       types.StringValue(propertyRes.Group),
		Name:        types.StringValue(propertyRes.Name),
		Value:       types.StringValue(propertyRes.Value),
		Type:        types.StringValue(propertyRes.Type),
		Description: types.StringValue(propertyRes.Description),
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated project property")
}

func (r *projectPropertyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectPropertyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := uuid.Parse(state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Delete, unable to parse project into UUID.",
			"Error from: "+err.Error(),
		)
		return
	}
	groupName := state.Group.ValueString()
	propertyName := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting project property", map[string]any{
		"project": project.String(),
		"group":   groupName,
		"name":    propertyName,
	})
	/*err = r.client.ProjectProperty.Delete(ctx, project, groupName, propertyName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete project property.",
			"Error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted project property.")*/
	resp.Diagnostics.AddWarning(
		"Project property has not been deleted.",
		"Due to an error when using the SDK, this provider is unable to delete project properties.",
	)
}

func (r *projectPropertyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected import id",
			fmt.Sprintf("Expected id in format <UUID>/<Group>/<Name>. Received %s", req.ID),
		)
		return
	}
	uuid, err := uuid.Parse(idParts[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected import id",
			"Unable to parse UUID: "+err.Error(),
		)
		return
	}
	groupName := idParts[1]
	propertyName := idParts[2]

	property, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.ProjectProperty], error) {
			return r.client.ProjectProperty.GetAll(ctx, uuid, po)
		},
		func(property dtrack.ProjectProperty) bool {
			if property.Group != groupName {
				return false
			}
			if property.Name != propertyName {
				return false
			}
			return true
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to locate project property.",
			"Unexpected error from: "+err.Error(),
		)
	}
	propertyState := projectPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s/%s", uuid.String(), property.Group, property.Name)),
		Project:     types.StringValue(uuid.String()),
		Group:       types.StringValue(property.Group),
		Name:        types.StringValue(property.Name),
		Value:       types.StringValue(property.Value),
		Type:        types.StringValue(property.Type),
		Description: types.StringValue(property.Description),
	}
	diags := resp.State.Set(ctx, propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported a project property.")
}

func (r *projectPropertyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = client
}
