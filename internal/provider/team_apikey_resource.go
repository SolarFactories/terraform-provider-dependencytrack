package provider

import (
	"context"
	"fmt"
	"strings"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &teamAPIKeyResource{}
	_ resource.ResourceWithConfigure   = &teamAPIKeyResource{}
	_ resource.ResourceWithImportState = &teamAPIKeyResource{}
)

type (
	teamAPIKeyResource struct {
		client *dtrack.Client
		semver *Semver
	}

	teamAPIKeyResourceModel struct {
		ID       types.String `tfsdk:"id"`
		TeamID   types.String `tfsdk:"team"`
		Key      types.String `tfsdk:"key"`
		Comment  types.String `tfsdk:"comment"`
		Masked   types.String `tfsdk:"masked"`
		PublicID types.String `tfsdk:"public_id"`
		Legacy   types.Bool   `tfsdk:"legacy"`
	}
)

func NewTeamAPIKeyResource() resource.Resource {
	return &teamAPIKeyResource{}
}

func (*teamAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_apikey"
}

func (*teamAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				// Defaults to "".
				Computed: true,
			},
			"masked": schema.StringAttribute{
				Description: "The masked API Key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_id": schema.StringAttribute{
				Description: "The public identifier for API Keys in DependencyTrack 4.13+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"legacy": schema.BoolAttribute{
				Description: "Whether the API Key is generated by DependencyTrack pre-4,13.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *teamAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamAPIKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	teamID, diag := TryParseUUID(plan.TeamID, LifecycleCreate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	comment := plan.Comment.ValueString()

	tflog.Debug(ctx, "Creating API Key", map[string]any{
		"team": teamID.String(),
	})
	key, err := r.client.Team.GenerateAPIKey(ctx, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team API Key",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	if comment != "" {
		tflog.Debug(ctx, "Setting Comment for API Key", map[string]any{
			"team":    teamID.String(),
			"masked":  key.MaskedKey,
			"comment": comment,
		})
		comment, err = r.client.Team.UpdateAPIKeyComment(ctx, key.Key, comment)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting API Key comment",
				"Unexpected error: "+err.Error(),
			)
			return
		}
	}

	plan = teamAPIKeyResourceModel{
		ID:       types.StringNull(), // Set below.
		TeamID:   types.StringValue(teamID.String()),
		Key:      types.StringValue(key.Key),
		Comment:  types.StringValue(comment),
		Masked:   types.StringValue(key.MaskedKey),
		PublicID: types.StringValue(key.PublicId),
		Legacy:   types.BoolValue(key.Legacy),
	}
	if r.isLegacy(key) {
		plan.ID = types.StringValue(fmt.Sprintf("%s/%s", teamID.String(), key.Key))
	} else {
		plan.ID = types.StringValue(fmt.Sprintf("%s/%s", teamID.String(), key.PublicId))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created API Key", map[string]any{
		"team":    plan.TeamID.ValueString(),
		"masked":  plan.Masked.ValueString(),
		"comment": plan.Comment.ValueString(),
	})
}

func (r *teamAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state teamAPIKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := state.Key.ValueString()
	publicID := state.PublicID.ValueString() // DT API v4.13+.
	team, diag := TryParseUUID(state.TeamID, LifecycleRead, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "Reading API Key", map[string]any{
		"team":    team.String(),
		"masked":  state.Masked.ValueString(),
		"comment": state.Comment.ValueString(),
	})

	keys, err := r.client.Team.GetAPIKeys(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read API Keys",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	apiKey, err := Find(keys, func(apiKey dtrack.APIKey) bool {
		if r.isLegacy(apiKey) {
			return apiKey.Key == key
		} else {
			return apiKey.PublicId == publicID
		}
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to find API Key",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	state = teamAPIKeyResourceModel{
		// Due to construction of ID varying for legacy, copy across.
		ID:     types.StringValue(state.ID.ValueString()),
		TeamID: types.StringValue(team.String()),
		// Key value not returned in API 4.13+, since it changed to ENCRYPTEDSTRING, so retain current state value.
		Key:      types.StringValue(state.Key.ValueString()),
		Comment:  types.StringValue(apiKey.Comment),
		Masked:   types.StringValue(apiKey.MaskedKey),
		PublicID: types.StringValue(apiKey.PublicId),
		Legacy:   types.BoolValue(apiKey.Legacy),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read API Key", map[string]any{
		"team":    state.TeamID.ValueString(),
		"masked":  state.Masked.ValueString(),
		"comment": state.Comment.ValueString(),
	})
}

func (r *teamAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State.
	var plan teamAPIKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicIDOrKey := r.getPublicIDOrKey(plan)
	comment := plan.Comment.ValueString()
	tflog.Debug(ctx, "Updating API Key Comment", map[string]any{
		"team":    plan.TeamID.ValueString(),
		"masked":  plan.Masked.ValueString(),
		"comment": comment,
	})

	commentOut, err := r.client.Team.UpdateAPIKeyComment(ctx, publicIDOrKey, comment)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update the API Key comment.",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	plan.Comment = types.StringValue(commentOut)

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated API Key", map[string]any{
		"team":    plan.TeamID.ValueString(),
		"masked":  plan.Masked.ValueString(),
		"comment": plan.Comment.ValueString(),
	})
}

func (r *teamAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state teamAPIKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	publicIDOrKey := r.getPublicIDOrKey(state)
	team, diag := TryParseUUID(state.TeamID, LifecycleDelete, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	// Execute.
	tflog.Debug(ctx, "Deleting API Key", map[string]any{
		"team":    team.String(),
		"masked":  state.Masked.ValueString(),
		"comment": state.Comment.ValueString(),
	})
	err := r.client.Team.DeleteAPIKey(ctx, publicIDOrKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Team API Key",
			"Unexpected error when trying to delete Team API Key: "+team.String()+", from error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted API Key", map[string]any{
		"team":    state.TeamID.ValueString(),
		"masked":  state.Masked.ValueString(),
		"comment": state.Comment.ValueString(),
	})
}

func (r *teamAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected import id",
			"Expected id in format <UUID>/<PublicIdOrKey>. Received "+req.ID,
		)
		return
	}
	teamID, diag := TryParseUUID(types.StringValue(idParts[0]), LifecycleImport, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	publicIDOrKey := idParts[1]
	tflog.Debug(ctx, "Importing API Key", map[string]any{
		"team": teamID.String(),
	})
	keys, err := r.client.Team.GetAPIKeys(ctx, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to retrieve Team API Keys.",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	apiKey, err := Find(keys, func(apiKey dtrack.APIKey) bool {
		if r.isLegacy(apiKey) {
			return apiKey.Key == publicIDOrKey
		} else {
			return apiKey.PublicId == publicIDOrKey
		}
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to identify API Key.",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	keyState := teamAPIKeyResourceModel{
		ID:       types.StringValue(fmt.Sprintf("%s/%s", teamID.String(), publicIDOrKey)),
		TeamID:   types.StringValue(teamID.String()),
		Key:      types.StringValue(apiKey.Key),
		Comment:  types.StringValue(apiKey.Comment),
		Masked:   types.StringValue(apiKey.MaskedKey),
		PublicID: types.StringValue(apiKey.PublicId),
		Legacy:   types.BoolValue(apiKey.Legacy),
	}

	diags := resp.State.Set(ctx, keyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported API Key", map[string]any{
		"team":    keyState.TeamID.ValueString(),
		"masked":  keyState.Masked.ValueString(),
		"comment": keyState.Comment.ValueString(),
	})
}

func (r *teamAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = clientInfoData.client
	r.semver = clientInfoData.semver
}

func (r *teamAPIKeyResource) isLegacy(key dtrack.APIKey) bool {
	if r.semver.Major == 4 && r.semver.Minor < 13 {
		return true
	}
	return key.Legacy
}

func (r *teamAPIKeyResource) getPublicIDOrKey(model teamAPIKeyResourceModel) string {
	key := dtrack.APIKey{Legacy: model.Legacy.ValueBool()}
	if r.isLegacy(key) {
		return model.Key.ValueString()
	} else {
		return model.PublicID.ValueString()
	}
}
