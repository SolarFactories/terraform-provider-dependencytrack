package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &oidcGroupMappingResource{}
	_ resource.ResourceWithConfigure   = &oidcGroupMappingResource{}
	_ resource.ResourceWithImportState = &oidcGroupMappingResource{}
)

type (
	oidcGroupMappingResource struct {
		client *dtrack.Client
		semver *Semver
	}

	oidcGroupMappingResourceModel struct {
		ID    types.String `tfsdk:"id"`
		Team  types.String `tfsdk:"team"`
		Group types.String `tfsdk:"group"`
	}
)

func NewOidcGroupMappingResource() resource.Resource {
	return &oidcGroupMappingResource{}
}

func (*oidcGroupMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_group_mapping"
}

func (*oidcGroupMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a mapping from OIDC Group to Team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID for the Mapping, as generated by DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team": schema.StringAttribute{
				Description: "UUID for the Team to map from the OIDC Group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group": schema.StringAttribute{
				Description: "UUID for the OIDC Group to map to the Team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *oidcGroupMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oidcGroupMappingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, diag := TryParseUUID(plan.Team, LifecycleCreate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	group, diag := TryParseUUID(plan.Group, LifecycleCreate, path.Root("group"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	mappingReq := dtrack.OIDCMappingRequest{
		Group: group,
		Team:  team,
	}
	tflog.Debug(ctx, "Creating OIDC Group Mapping", map[string]any{
		"group": group.String(),
		"team":  team.String(),
	})
	mappingRes, err := r.client.OIDC.AddTeamMapping(ctx, mappingReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group mapping",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	plan = oidcGroupMappingResourceModel{
		ID:    types.StringValue(mappingRes.UUID.String()),
		Group: types.StringValue(mappingRes.Group.UUID.String()),
		// Response does not include Team UUID.
		Team: types.StringValue(team.String()),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created OIDC Group Mapping", map[string]any{
		"id":    plan.ID.ValueString(),
		"group": plan.Group.ValueString(),
		"team":  plan.Team.ValueString(),
	})
}

func (r *oidcGroupMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state oidcGroupMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading OIDC Group Mapping", map[string]any{
		"id":    state.ID.ValueString(),
		"team":  state.Team.ValueString(),
		"group": state.Group.ValueString(),
	})

	// Refresh.
	id, diag := TryParseUUID(state.ID, LifecycleRead, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	mappingInfo, err := FindPagedOidcMapping(id, func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
		return r.client.Team.GetAll(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get group team mapping within Read",
			fmt.Sprintf(
				"Error with reading OIDC Group Mapping with id: %s, for team: %s, and group: %s, in original errr: %s",
				id.String(), state.Team.ValueString(), state.Group.ValueString(), err.Error(),
			),
		)
		return
	}
	state = oidcGroupMappingResourceModel{
		ID:    types.StringValue(id.String()),
		Team:  types.StringValue(mappingInfo.Team.String()),
		Group: types.StringValue(mappingInfo.Group.String()),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read OIDC Group Mapping", map[string]any{
		"id":    state.ID.ValueString(),
		"team":  state.Team.ValueString(),
		"group": state.Group.ValueString(),
	})
}

func (r *oidcGroupMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State.
	var plan oidcGroupMappingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating OIDC Group Mapping", map[string]any{
		"id":    plan.ID.ValueString(),
		"team":  plan.Team.ValueString(),
		"group": plan.Group.ValueString(),
	})

	id, diag := TryParseUUID(plan.ID, LifecycleUpdate, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	mappingInfo, err := FindPagedOidcMapping(id, func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
		return r.client.Team.GetAll(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get group team mapping within Update",
			fmt.Sprintf(
				"Error with reading OIDC Group Mapping with id: %s, for team: %s, and group: %s, in original errr: %s",
				id.String(), mappingInfo.Team.String(), mappingInfo.Group.String(), err.Error(),
			),
		)
		return
	}

	plan = oidcGroupMappingResourceModel{
		ID:    types.StringValue(id.String()),
		Team:  types.StringValue(mappingInfo.Team.String()),
		Group: types.StringValue(mappingInfo.Group.String()),
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated OIDC Group Mapping", map[string]any{
		"id":    plan.ID.ValueString(),
		"team":  plan.Team.ValueString(),
		"group": plan.Group.ValueString(),
	})
}

func (r *oidcGroupMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state oidcGroupMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	id, diag := TryParseUUID(state.ID, LifecycleDelete, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	// Execute.
	tflog.Debug(ctx, "Deleting OIDC Group Mapping", map[string]any{
		"id":    id.String(),
		"team":  state.Team.ValueString(),
		"group": state.Group.ValueString(),
	})
	err := r.client.OIDC.RemoveTeamMapping(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete group mapping",
			"Unexpected error when trying to delete oidc group mapping with id: "+id.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted OIDC Group Mapping", map[string]any{
		"id":    state.ID.ValueString(),
		"team":  state.Team.ValueString(),
		"group": state.Group.ValueString(),
	})
}

func (*oidcGroupMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing OIDC Group Mapping", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported OIDC Group Mapping", map[string]any{
		"id": req.ID,
	})
}

func (r *oidcGroupMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
