package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"

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
	_ resource.Resource                = &repositoryResource{}
	_ resource.ResourceWithConfigure   = &repositoryResource{}
	_ resource.ResourceWithImportState = &repositoryResource{}
)

func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

type repositoryResource struct {
	client *dtrack.Client
}

type repositoryResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	Identifier types.String `tfsdk:"identifier"`
	URL        types.String `tfsdk:"url"`
	Precedence types.Int32  `tfsdk:"precedence"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Internal   types.Bool   `tfsdk:"internal"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
}

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID for the Repository as generated by DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of the Repository. See DependencyTrack for valid enum values.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identifier": schema.StringAttribute{
				Description: "Identifier of the Repository.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Description: "URL of the Repository.",
				Required:    true,
			},
			"precedence": schema.Int32Attribute{
				Description: "Precedence / Resolution Order of the Repository.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the Repository Enabled.",
				Required:    true,
			},
			"internal": schema.BoolAttribute{
				Description: "Whether the Repository is Internal.",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username to use for Authentication to Repository.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password to use for Authentication to Repository.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryReq := dtrack.Repository{
		Type:            dtrack.RepositoryType(plan.Type.ValueString()),
		Identifier:      plan.Identifier.ValueString(),
		Url:             plan.URL.ValueString(),
		ResolutionOrder: int(plan.Precedence.ValueInt32()),
		Enabled:         plan.Enabled.ValueBool(),
		Internal:        plan.Internal.ValueBool(),
		Username:        plan.Username.ValueString(),
		Password:        plan.Password.ValueString(),
	}

	tflog.Debug(ctx, "Creating a new repository, with type: "+string(repositoryReq.Type)+", and identifier: "+repositoryReq.Identifier)
	// NOTE: Has a patch applied in `http_client.go`
	repositoryRes, err := r.client.Repository.Create(ctx, repositoryReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository",
			"Could not create repository, unexpected error within Client: "+err.Error(),
		)
		return
	}

	plan = repositoryResourceModel{
		ID:         types.StringValue(repositoryRes.UUID.String()),
		Type:       types.StringValue(string(repositoryRes.Type)),
		Identifier: types.StringValue(repositoryRes.Identifier),
		URL:        types.StringValue(repositoryRes.Url),
		Precedence: types.Int32Value(int32(repositoryRes.ResolutionOrder)),
		Enabled:    types.BoolValue(repositoryRes.Enabled),
		Internal:   types.BoolValue(repositoryRes.Internal),
		Username:   types.StringValue(repositoryRes.Username),
		// API Response does not include Password
		Password: plan.Password,
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a new repository, with id: "+repositoryRes.UUID.String())
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state repositoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh
	repoType := state.Type.ValueString()
	uuidString := state.ID.ValueString()

	tflog.Debug(ctx, "Refreshing repository with type: "+repoType+", id: "+uuidString)
	id, err := uuid.Parse(uuidString)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Read, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	repository, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.Repository], error) {
			return r.client.Repository.GetByType(ctx, dtrack.RepositoryType(repoType), po)
		},
		func(repo dtrack.Repository) bool {
			if repo.UUID != id {
				return false
			}
			return true
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated repository",
			"Error with reading repository: "+id.String()+", in original error: "+err.Error(),
		)
		return
	}

	state = repositoryResourceModel{
		ID:         types.StringValue(repository.UUID.String()),
		Type:       types.StringValue(string(repository.Type)),
		Identifier: types.StringValue(repository.Identifier),
		URL:        types.StringValue(repository.Url),
		Precedence: types.Int32Value(int32(repository.ResolutionOrder)),
		Enabled:    types.BoolValue(repository.Enabled),
		Internal:   types.BoolValue(repository.Internal),
		Username:   types.StringValue(repository.Username),
		// API Response does not include Password
		Password: state.Password,
	}

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed repository with id: "+state.ID.ValueString())
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State
	var plan repositoryResourceModel
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
	repositoryReq := dtrack.Repository{
		UUID:            id,
		Type:            dtrack.RepositoryType(plan.Type.ValueString()),
		Identifier:      plan.Identifier.ValueString(),
		Url:             plan.URL.ValueString(),
		ResolutionOrder: int(plan.Precedence.ValueInt32()),
		Enabled:         plan.Enabled.ValueBool(),
		Internal:        plan.Internal.ValueBool(),
		Username:        plan.Username.ValueString(),
		Password:        plan.Password.ValueString(),
	}

	// Execute
	tflog.Debug(ctx, "Updating repository with id: "+id.String())
	repositoryRes, err := r.client.Repository.Update(ctx, repositoryReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update repository",
			"Error in: "+id.String()+", from: "+err.Error(),
		)
		return
	}

	// Map SDK to TF
	plan = repositoryResourceModel{
		ID:         types.StringValue(repositoryRes.UUID.String()),
		Type:       types.StringValue(string(repositoryRes.Type)),
		Identifier: types.StringValue(repositoryRes.Identifier),
		URL:        types.StringValue(repositoryRes.Url),
		Precedence: types.Int32Value(int32(repositoryRes.ResolutionOrder)),
		Enabled:    types.BoolValue(repositoryRes.Enabled),
		Internal:   types.BoolValue(repositoryRes.Internal),
		Username:   types.StringValue(repositoryRes.Username),
		// API Response does not include Password
		Password: plan.Password,
	}

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated repository with id: "+id.String())
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state repositoryResourceModel
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
	tflog.Debug(ctx, "Deleting repository with id: "+id.String())
	err = r.client.Repository.Delete(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete repository",
			"Unexpected error when trying to delete repository: "+id.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted repository with id: "+id.String())
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *repositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
