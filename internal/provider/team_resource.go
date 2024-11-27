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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	client *dtrack.Client
}

type teamResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID for the Team as generated by DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Team.",
				Required:    true,
			},
		},
	}
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamReq := dtrack.Team{
		Name: plan.Name.ValueString(),
	}
	tflog.Debug(ctx, "Creating a new team, with name: "+teamReq.Name)
	teamRes, err := r.client.Team.Create(ctx, teamReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(teamRes.UUID.String())
	plan.Name = types.StringValue(teamRes.Name)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a new team, with id: "+teamRes.UUID.String())
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state teamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh
	tflog.Debug(ctx, "Refreshing team with id: "+state.ID.ValueString())
	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Read, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	team, err := r.client.Team.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated team",
			"Error with reading team: "+id.String()+", in original error: "+err.Error(),
		)
		return
	}
	state.ID = types.StringValue(team.UUID.String())
	state.Name = types.StringValue(team.Name)

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed team with id: "+state.ID.ValueString())
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State
	var plan teamResourceModel
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
	teamReq := dtrack.Team{
		UUID: id,
		Name: plan.Name.ValueString(),
	}

	// Execute
	tflog.Debug(ctx, "Updating team with id: "+id.String())
	teamRes, err := r.client.Team.Update(ctx, teamReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update team",
			"Error in: "+id.String()+", from: "+err.Error(),
		)
		return
	}

	// Map SDK to TF
	plan.ID = types.StringValue(teamRes.UUID.String())
	plan.Name = types.StringValue(teamRes.Name)

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated team with id: "+id.String())
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state teamResourceModel
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
	team := dtrack.Team{
		UUID: id,
	}

	// Execute
	tflog.Debug(ctx, "Deleting team with id: "+id.String())
	err = r.client.Team.Delete(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete team",
			"Unexpected error when trying to delete team: "+id.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted team with id: "+id.String())
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
