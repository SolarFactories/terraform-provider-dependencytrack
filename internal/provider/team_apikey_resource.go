package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	"github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &teamApiKeyResource{}
	_ resource.ResourceWithConfigure = &teamApiKeyResource{}
)

func NewTeamApiKeyResource() resource.Resource {
	return &teamApiKeyResource{}
}

type teamApiKeyResource struct {
	client *dtrack.Client
}

type teamApiKeyResourceModel struct {
	TeamID types.String `tfsdk:"team"`
	Key    types.String `tfsdk:"key"`
}

func (r *teamApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_apikey"
}

func (r *teamApiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an API Key for a Team..",
		Attributes: map[string]schema.Attribute{
			"team": schema.StringAttribute{
				Description: "UUID for the Team for which to manage the permission.",
				Required:    true,
			},
			"key": schema.StringAttribute{
				Description: "The generated API Key for the Team.",
				Sensitive:   true,
				Computed:    true,
			},
		},
	}
}

func (r *teamApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamApiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, err := uuid.Parse(plan.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Create, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Creating API Key for team "+team.String())
	key, err := r.client.Team.GenerateAPIKey(ctx, team)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team API Key",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	plan.TeamID = types.StringValue(team.String())
	plan.Key = types.StringValue(key)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created API Key for team, with id: "+team.String())
}

func (r *teamApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No way to re-read and identify the API Key, without already knowing it. This resource only has Create, Delete actions
	// Fetch state
	var state teamApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed API Key for team with id: "+state.TeamID.ValueString())
}

func (r *teamApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions
	// Get State
	var plan teamApiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated API Key for team with id: "+plan.TeamID.ValueString())
}

func (r *teamApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state teamApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK
	/*team, err := uuid.Parse(state.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Delete, unable to parse UUID",
			"Error parsing UUID from: "+state.TeamID.ValueString()+", error: "+err.Error(),
		)
		return
	}
	// Execute
	tflog.Debug(ctx, "Deleting API Key from team with id: "+team.String())
	_, err = r.client.Team.DeleteAPIKey(ctx, team, state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Team API Key",
			"Unexpected error when trying to delete Team API Key: "+team.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted API Key from team with id: "+team.String())*/
	resp.Diagnostics.AddWarning(
		"Team API Key has not been deleted.",
		"Due to functionality not existing within the SDK, this provider is unable to delete Team API Keys.",
	)

}

func (r *teamApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
