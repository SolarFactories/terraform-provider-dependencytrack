package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	"github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *dtrack.Client
}

type projectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Active      types.Bool   `tfsdk:"active"`
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectReq := dtrack.Project{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Active:      plan.Active.ValueBool(),
	}

	tflog.Debug(ctx, "Creating a new project, with name: "+projectReq.Name)
	projectRes, err := r.client.Project.Create(ctx, projectReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project, unexpected error within Client: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(projectRes.UUID.String())
	plan.Name = types.StringValue(projectRes.Name)
	plan.Description = types.StringValue(projectRes.Description)
	plan.Active = types.BoolValue(projectRes.Active)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a new project, with id: "+projectRes.UUID.String())
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh
	tflog.Debug(ctx, "Refreshing project with id: "+state.ID.ValueString())
	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Read, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	project, err := r.client.Project.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated project",
			"Error with reading project: "+id.String()+", in original error: "+err.Error(),
		)
		return
	}
	state.ID = types.StringValue(project.UUID.String())
	state.Name = types.StringValue(project.Name)
	state.Description = types.StringValue(project.Description)
	state.Active = types.BoolValue(project.Active)

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed project with id: "+state.ID.ValueString())
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK
	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Update, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	projectReq := dtrack.Project{
		UUID:        id,
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Active:      plan.Active.ValueBool(),
	}

	// Execute
	tflog.Debug(ctx, "Updating project with id: "+id.String())
	_, err = r.client.Project.Update(ctx, projectReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project",
			"Error in: "+id.String()+", from: "+err.Error(),
		)
		return
	}
	projectRes, err := r.client.Project.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project",
			"Error in: "+id.String()+", from: "+err.Error(),
		)
		return
	}

	// Map SDK to TF
	plan.ID = types.StringValue(projectRes.UUID.String())
	plan.Name = types.StringValue(projectRes.Name)
	plan.Description = types.StringValue(projectRes.Description)
	plan.Active = types.BoolValue(projectRes.Active)

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated project with id: "+id.String())
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK
	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Delete, unable to parse UUID",
			"Error parsing UUID from: "+state.ID.ValueString()+", error: "+err.Error(),
		)
		return
	}

	// Execute
	tflog.Debug(ctx, "Deleting project with id: "+id.String())
	err = r.client.Project.Delete(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete project",
			"Unexpected error when trying to delete project: "+id.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted project with id: "+id.String())
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = client
}
