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
	_ resource.Resource                = &oidcUserResource{}
	_ resource.ResourceWithConfigure   = &oidcUserResource{}
	_ resource.ResourceWithImportState = &oidcUserResource{}
)

type (
	oidcUserResource struct {
		client *dtrack.Client
		semver *Semver
	}

	oidcUserResourceModel struct {
		ID       types.String `tfsdk:"id"`
		Username types.String `tfsdk:"username"`
	}
)

func NewOIDCUserResource() resource.Resource {
	return &oidcUserResource{}
}

func (*oidcUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_user"
}

func (*oidcUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OIDC User.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Username of the User.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username of the User.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *oidcUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oidcUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userReq := dtrack.OIDCUser{
		Username: plan.Username.ValueString(),
	}

	tflog.Debug(ctx, "Creating OIDC User", map[string]any{
		"username": userReq.Username,
	})

	oidcUserRes, err := r.client.OIDC.CreateUser(ctx, userReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating OIDC user",
			"Error from: "+err.Error(),
		)
		return
	}

	plan = oidcUserResourceModel{
		ID:       types.StringValue(oidcUserRes.Username),
		Username: types.StringValue(oidcUserRes.Username),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created OIDC User", map[string]any{
		"id":       plan.ID.ValueString(),
		"username": plan.Username.ValueString(),
	})
}

func (r *oidcUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state oidcUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.ID.ValueString()

	// Refresh.
	users, err := r.client.OIDC.GetAllUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read OIDC users",
			"Error for user: "+username+", in original error: "+err.Error(),
		)
		return
	}
	user, err := Find(users.Items, func(user dtrack.OIDCUser) bool { return user.Username == username })
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to locate OIDC user",
			"Error for user: "+username+", in original error: "+err.Error(),
		)
		return
	}
	newState := oidcUserResourceModel{
		ID:       types.StringValue(user.Username),
		Username: types.StringValue(user.Username),
	}

	// Update state.
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read OIDC User", map[string]any{
		"id":       state.ID.ValueString(),
		"username": state.Username.ValueString(),
	})
}

func (*oidcUserResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Resource has no update, since all attributes are RequireReplace.
}

func (r *oidcUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load State.
	var state oidcUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	user := dtrack.OIDCUser{
		Username: state.Username.ValueString(),
	}

	// Execute.
	tflog.Debug(ctx, "Deleting OIDC User", map[string]any{
		"id":       user.Username,
		"username": state.Username.ValueString(),
	})
	err := r.client.OIDC.DeleteUser(ctx, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete OIDC user",
			"Error for user: "+user.Username+", from original error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted OIDC User", map[string]any{
		"id":       state.ID.ValueString(),
		"username": state.Username.ValueString(),
	})
}

func (*oidcUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing OIDC User", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported OIDC User", map[string]any{
		"id": req.ID,
	})
}

func (r *oidcUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
