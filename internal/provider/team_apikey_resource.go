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
	_ resource.Resource                = &teamApiKeyResource{}
	_ resource.ResourceWithConfigure   = &teamApiKeyResource{}
	_ resource.ResourceWithImportState = &teamApiKeyResource{}
)

func NewTeamApiKeyResource() resource.Resource {
	return &teamApiKeyResource{}
}

type teamApiKeyResource struct {
	client *dtrack.Client
}

type teamApiKeyResourceModel struct {
	ID      types.String `tfsdk:"id"`
	TeamID  types.String `tfsdk:"team"`
	Key     types.String `tfsdk:"key"`
	Comment types.String `tfsdk:"comment"`
}

func (r *teamApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_apikey"
}

func (r *teamApiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an API Key for a Team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID used by provider. Has no meaning to DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team": schema.StringAttribute{
				Description: "UUID for the Team for which to manage the permission.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The generated API Key for the Team.",
				Sensitive:   true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				Description: "The comment to assign to the API Key.",
				Optional:    true,
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
	comment := plan.Comment.ValueString()

	tflog.Debug(ctx, "Creating API Key for team "+team.String())
	key, err := r.client.Team.GenerateAPIKey(ctx, team)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team API Key",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	if comment != "" {
		tflog.Debug(ctx, "Setting Comment for API Key for team "+team.String())
		_, err = r.client.Team.UpdateAPIKeyComment(ctx, key, comment)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting API Key comment",
				"Unexpected error: "+err.Error(),
			)
			return
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", team.String(), key))
	plan.TeamID = types.StringValue(team.String())
	plan.Key = types.StringValue(key)
	plan.Comment = types.StringValue(comment)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created API Key for team, with id: "+team.String())
}

func (r *teamApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state teamApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := state.Key.ValueString()
	team, err := uuid.Parse(state.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Read, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	keys, err := r.client.Team.GetAPIKeys(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read API Keys",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	apiKey, err := Find(keys, func(apiKey dtrack.APIKey) bool { return apiKey.Key == key })
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to find API Key",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", team.String(), apiKey.Key))
	state.TeamID = types.StringValue(team.String())
	state.Key = types.StringValue(apiKey.Key)
	state.Comment = types.StringValue(apiKey.Comment)

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed API Key for team with id: "+state.TeamID.ValueString())
}

func (r *teamApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State
	var plan teamApiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := plan.Key.ValueString()
	comment := plan.Comment.ValueString()

	_, err := r.client.Team.UpdateAPIKeyComment(ctx, key, comment)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update the API Key comment.",
			"Unexpected error: "+err.Error(),
		)
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
	key := state.Key.ValueString()
	team, err := uuid.Parse(state.TeamID.ValueString())
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
	err = r.client.Team.DeleteAPIKey(ctx, key)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Team API Key",
			"Unexpected error when trying to delete Team API Key: "+team.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted API Key from team with id: "+team.String())
}

func (r *teamApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected import id",
			fmt.Sprintf("Expected id in format <UUID>/<Key>. Received %s", req.ID),
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
	key := idParts[1]
	keys, err := r.client.Team.GetAPIKeys(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to retrieve Team API Keys.",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	apiKey, err := Find(keys, func(apiKey dtrack.APIKey) bool { return apiKey.Key == key })
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to identify API Key.",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	keyState := teamApiKeyResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", uuid.String(), apiKey.Key)),
		TeamID:  types.StringValue(uuid.String()),
		Key:     types.StringValue(apiKey.Key),
		Comment: types.StringValue(apiKey.Comment),
	}

	diags := resp.State.Set(ctx, keyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported a project property.")
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
