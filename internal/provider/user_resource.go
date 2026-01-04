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
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

type (
	userResource struct {
		client *dtrack.Client
		semver *Semver
	}

	userResourceModel struct {
		ID                  types.String `tfsdk:"id"`
		Username            types.String `tfsdk:"username"`
		Fullname            types.String `tfsdk:"fullname"`
		Email               types.String `tfsdk:"email"`
		Password            types.String `tfsdk:"password"`
		Suspended           types.Bool   `tfsdk:"suspended"`
		ForcePasswordChange types.Bool   `tfsdk:"force_password_change"`
		PasswordExpires     types.Bool   `tfsdk:"password_expires"`
	}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

func (*userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (*userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Managed User.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User's username.",
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
			"fullname": schema.StringAttribute{
				Description: "Full name of the User.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email Address of the User.",
				Required:    true,
			},
			"suspended": schema.BoolAttribute{
				Description: "Whether the User Account is Suspended.",
				Optional:    true,
				Computed:    true,
			},
			"force_password_change": schema.BoolAttribute{
				Description: "Whether the User Must change their password on next login.",
				Optional:    true,
				Computed:    true,
			},
			"password_expires": schema.BoolAttribute{
				Description: "Whether the User's password expires. Interval set by DependencyTrack.",
				Optional:    true,
				Computed:    true,
			},
			"password": schema.StringAttribute{
				Description: "Updated password to set for the user.",
				Sensitive:   true,
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Password.IsUnknown() || plan.Password.IsNull() {
		resp.Diagnostics.AddError(
			"Missing required password field",
			"Password is required when creating a managed user.",
		)
		return
	}

	userReq := dtrack.ManagedUser{
		Username:            plan.Username.ValueString(),
		Fullname:            plan.Fullname.ValueString(),
		Email:               plan.Email.ValueString(),
		Suspended:           plan.Suspended.ValueBool(),
		ForcePasswordChange: plan.ForcePasswordChange.ValueBool(),
		NonExpiryPassword:   !plan.PasswordExpires.ValueBool(),
		NewPassword:         plan.Password.ValueString(),
		ConfirmPassword:     plan.Password.ValueString(),
	}

	tflog.Debug(ctx, "Creating Managed User", map[string]any{
		"username":              userReq.Username,
		"fullname":              userReq.Fullname,
		"email":                 userReq.Email,
		"suspended":             userReq.Suspended,
		"force_password_change": userReq.ForcePasswordChange,
		"non_expiry_password":   userReq.NonExpiryPassword,
		"password":              userReq.NewPassword,
	})

	userRes, err := r.client.User.CreateManaged(ctx, userReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating managed user",
			"Error from: "+err.Error(),
		)
		return
	}

	plan = userResourceModel{
		ID:                  types.StringValue(userRes.Username),
		Username:            types.StringValue(userRes.Username),
		Fullname:            types.StringValue(userRes.Fullname),
		Email:               types.StringValue(userRes.Email),
		Password:            types.StringValue(userReq.NewPassword),
		Suspended:           types.BoolValue(userRes.Suspended),
		ForcePasswordChange: types.BoolValue(userRes.ForcePasswordChange),
		PasswordExpires:     types.BoolValue(!userRes.NonExpiryPassword),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Managed User", map[string]any{
		"id":                    plan.ID.ValueString(),
		"username":              plan.Username.ValueString(),
		"fullname":              plan.Fullname.ValueString(),
		"email":                 plan.Email.ValueString(),
		"password":              plan.Password.ValueString(),
		"suspended":             plan.Suspended.ValueBool(),
		"force_password_change": plan.ForcePasswordChange.ValueBool(),
		"password_expires":      plan.PasswordExpires.ValueBool(),
	})
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.ID.ValueString()

	// Refresh.
	user, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.ManagedUser], error) {
			return r.client.User.GetAllManaged(ctx, po)
		},
		func(user dtrack.ManagedUser) bool {
			return user.Username == username
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read managed user",
			"Error for user: "+username+", in original error: "+err.Error(),
		)
		return
	}
	if user == nil {
		resp.Diagnostics.AddError(
			"Unable to locate managed user",
			"Could not find managed user with username: "+username,
		)
		return
	}
	newState := userResourceModel{
		ID:                  types.StringValue(user.Username),
		Username:            types.StringValue(user.Username),
		Fullname:            types.StringValue(user.Fullname),
		Email:               types.StringValue(user.Email),
		Password:            state.Password,
		Suspended:           types.BoolValue(user.Suspended),
		ForcePasswordChange: types.BoolValue(user.ForcePasswordChange),
		PasswordExpires:     types.BoolValue(!user.NonExpiryPassword),
	}

	// Update state.
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Managed User", map[string]any{
		"id":                    state.ID.ValueString(),
		"username":              state.ID.ValueString(),
		"fullname":              state.Fullname.ValueString(),
		"email":                 state.Email.ValueString(),
		"password":              state.Password.ValueString(),
		"suspended":             state.Suspended.ValueBool(),
		"force_password_change": state.ForcePasswordChange.ValueBool(),
		"password_expires":      state.PasswordExpires.ValueBool(),
	})
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State.
	var plan userResourceModel
	var state userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	userReq := dtrack.ManagedUser{
		Username:            plan.Username.ValueString(),
		Fullname:            plan.Fullname.ValueString(),
		Email:               plan.Email.ValueString(),
		Suspended:           plan.Suspended.ValueBool(),
		ForcePasswordChange: plan.ForcePasswordChange.ValueBool(),
		NonExpiryPassword:   !plan.PasswordExpires.ValueBool(),
	}

	// Execute.
	tflog.Debug(ctx, "Updating Managed User", map[string]any{
		"username":              userReq.Username,
		"fullname":              userReq.Fullname,
		"email":                 userReq.Email,
		"suspended":             userReq.Suspended,
		"force_password_change": userReq.ForcePasswordChange,
		"non_expiry_password":   userReq.NonExpiryPassword,
	})
	userRes, err := r.client.User.UpdateManaged(ctx, userReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update managed user",
			"Error for user: "+userReq.Username+", from original error: "+err.Error(),
		)
		return
	}

	// Map SDK to TF.
	state = userResourceModel{
		ID:                  types.StringValue(userRes.Username),
		Username:            types.StringValue(userRes.Username),
		Fullname:            types.StringValue(userRes.Fullname),
		Email:               types.StringValue(userRes.Email),
		Password:            state.Password,
		Suspended:           types.BoolValue(userRes.Suspended),
		ForcePasswordChange: types.BoolValue(userRes.ForcePasswordChange),
		PasswordExpires:     types.BoolValue(!userRes.NonExpiryPassword),
	}

	// Update State.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Managed User", map[string]any{
		"id":                    state.ID.ValueString(),
		"username":              state.Username.ValueString(),
		"fullname":              state.Fullname.ValueString(),
		"email":                 state.Email.ValueString(),
		"password":              state.Password.ValueString(),
		"suspended":             state.Suspended.ValueBool(),
		"force_password_change": state.ForcePasswordChange.ValueBool(),
		"password_expires":      state.PasswordExpires.ValueBool(),
	})
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load State.
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	user := dtrack.ManagedUser{
		Username: state.Username.ValueString(),
	}

	// Execute.
	tflog.Debug(ctx, "Deleting Managed User", map[string]any{
		"id": user.Username,
	})
	err := r.client.User.DeleteManaged(ctx, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete managed user",
			"Error for user: "+user.Username+", from original error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Managed User", map[string]any{
		"id": state.ID.ValueString(),
	})
}

func (*userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing Managed User", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Managed User", map[string]any{
		"id": req.ID,
	})
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
