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
	_ resource.Resource                = &ldapUserResource{}
	_ resource.ResourceWithConfigure   = &ldapUserResource{}
	_ resource.ResourceWithImportState = &ldapUserResource{}
)

type (
	ldapUserResource struct {
		client *dtrack.Client
		semver *Semver
	}

	ldapUserResourceModel struct {
		ID       types.String `tfsdk:"id"`
		Username types.String `tfsdk:"username"`
	}
)

func NewLDAPUserResource() resource.Resource {
	return &ldapUserResource{}
}

func (*ldapUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_user"
}

func (*ldapUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an LDAP User.",
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

func (r *ldapUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ldapUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userReq := dtrack.LdapUser{
		Username: plan.Username.ValueString(),
	}

	tflog.Debug(ctx, "Creating LDAP User", map[string]any{
		"username": userReq.Username,
	})

	ldapUserRes, err := r.client.LDAP.CreateUser(ctx, userReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating LDAP user",
			"Error from: "+err.Error(),
		)
		return
	}

	plan = ldapUserResourceModel{
		ID:       types.StringValue(ldapUserRes.Username),
		Username: types.StringValue(ldapUserRes.Username),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created LDAP User", map[string]any{
		"id":       plan.ID.ValueString(),
		"username": plan.Username.ValueString(),
	})
}

func (r *ldapUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ldapUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.ID.ValueString()

	// Refresh.
	user, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.LdapUser], error) {
			return r.client.LDAP.GetUsers(ctx, po)
		},
		func(user dtrack.LdapUser) bool {
			return user.Username == username
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read LDAP user",
			"Error for user: "+username+", in original error: "+err.Error(),
		)
		return
	}

	newState := ldapUserResourceModel{
		ID:       types.StringValue(user.Username),
		Username: types.StringValue(user.Username),
	}

	// Update state.
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read LDAP User", map[string]any{
		"id":       state.ID.ValueString(),
		"username": state.Username.ValueString(),
	})
}

func (*ldapUserResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Resource has no update, since all attributes are RequireReplace.
}

func (r *ldapUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load State.
	var state ldapUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	user := dtrack.LdapUser{
		Username: state.Username.ValueString(),
	}

	// Execute.
	tflog.Debug(ctx, "Deleting LDAP User", map[string]any{
		"id":       user.Username,
		"username": state.Username.ValueString(),
	})
	err := r.client.LDAP.DeleteUser(ctx, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete LDAP user",
			"Error for user: "+user.Username+", from original error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted LDAP User", map[string]any{
		"id":       state.ID.ValueString(),
		"username": state.Username.ValueString(),
	})
}

func (*ldapUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing LDAP User", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported LDAP User", map[string]any{
		"id": req.ID,
	})
}

func (r *ldapUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
