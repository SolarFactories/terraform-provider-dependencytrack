package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &userPermissionResource{}
	_ resource.ResourceWithConfigure = &userPermissionResource{}
)

type (
	userPermissionResource struct {
		client *dtrack.Client
		semver *Semver
	}

	userPermissionResourceModel struct {
		Username   types.String `tfsdk:"username"`
		Permission types.String `tfsdk:"permission"`
	}
)

func NewUserPermissionResource() resource.Resource {
	return &userPermissionResource{}
}

func (*userPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_permission"
}

func (*userPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the attachment of a Permission to a User.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Username for the User for which to manage the permission.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission": schema.StringAttribute{
				Description: "Permission name to attach to the User.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userPermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	username := plan.Username.ValueString()
	permission := dtrack.Permission{
		Name: plan.Permission.ValueString(),
	}
	tflog.Debug(ctx, "Creating User Permission", map[string]any{
		"username":   username,
		"permission": permission.Name,
	})
	user, err := r.client.Permission.AddPermissionToUser(ctx, permission, username)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user permission",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	state := userPermissionResourceModel{
		Username:   types.StringValue(user.Username),
		Permission: types.StringValue(permission.Name),
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created User Permission", map[string]any{
		"username":   state.Username.ValueString(),
		"permission": state.Permission.ValueString(),
	})
}

func (r *userPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state userPermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh.
	username := state.Username.ValueString()
	tflog.Debug(ctx, "Reading User Permission", map[string]any{
		"username":   username,
		"permission": state.Permission.ValueString(),
	})
	user, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.ManagedUser], error) {
			return r.client.User.GetAllManaged(ctx, po)
		},
		func(usr dtrack.ManagedUser) bool {
			return usr.Username == username
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated user",
			"Error with reading user: "+username+", in original error: "+err.Error(),
		)
		return
	}
	permission, err := Find(user.Permissions, func(permission dtrack.Permission) bool {
		return permission.Name == state.Permission.ValueString()
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to identify user permission",
			"Unexpected Error from: "+err.Error(),
		)
		return
	}
	state = userPermissionResourceModel{
		Username:   types.StringValue(user.Username),
		Permission: types.StringValue(permission.Name),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read User Permission", map[string]any{
		"username":   state.Username.ValueString(),
		"permission": state.Permission.ValueString(),
	})
}

func (*userPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions.
	// Get State.
	var plan userPermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated User Permission", map[string]any{
		"username":   plan.Username.ValueString(),
		"permission": plan.Permission.ValueString(),
	})
}

func (r *userPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state userPermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	username := state.Username.ValueString()
	permission := dtrack.Permission{
		Name: state.Permission.ValueString(),
	}

	// Execute.
	tflog.Debug(ctx, "Deleting User Permission", map[string]any{
		"username":   username,
		"permission": permission.Name,
	})
	_, err := r.client.Permission.RemovePermissionFromUser(ctx, permission, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete user permission",
			"Unexpected error when trying to delete user permission: "+username+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted User Permission", map[string]any{
		"username":   state.Username.ValueString(),
		"permission": state.Permission.ValueString(),
	})
}

func (r *userPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
