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
	_ resource.Resource              = &secretResource{}
	_ resource.ResourceWithConfigure = &secretResource{}
)

type (
	secretResource struct {
		client *dtrack.Client
		semver *Semver
	}

	secretResourceModel struct {
		Name        types.String `tfsdk:"name"`
		Value       types.String `tfsdk:"value"`
		Description types.String `tfsdk:"description"`
		URL         types.String `tfsdk:"url"`
	}
)

func NewSecretResource() resource.Resource {
	return &secretResource{}
}

func (*secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (*secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Secret. Requires API version >= 5.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the Secret.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value of the Secret.",
				Required:    true,
				Sensitive:   true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Secret.",
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "Location of the Secret.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretReq := dtrack.CreateSecretRequest{
		Name:        plan.Name.ValueString(),
		Value:       plan.Value.ValueString(),
		Description: plan.Description.ValueString(),
	}

	secretURL, err := r.client.Secrets.Create(ctx, secretReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to create Secret with name: "+secretReq.Name,
			"Error from: "+err.Error(),
		)
		return
	}
	state := secretResourceModel{
		Name:        plan.Name,
		Value:       plan.Value,
		Description: plan.Description,
		URL:         types.StringValue(secretURL),
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Secret", map[string]any{
		"name":        state.Name,
		"value.#":     len(state.Value.ValueString()),
		"description": state.Description.ValueString(),
		"url":         state.URL,
	})
}

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretName := state.Name.ValueString()
	tflog.Debug(ctx, "Reading Secret", map[string]any{
		"name":        secretName,
		"description": state.Description,
		"url":         state.URL,
	})

	metadata, err := r.client.Secrets.GetMetadata(ctx, secretName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to read Metadata for Secret: "+secretName,
			"Error from: "+err.Error(),
		)
		return
	}

	newState := secretResourceModel{
		Name:        state.Name,
		Value:       state.Value,
		Description: types.StringUnknown(), // Set Below.
		URL:         state.URL,
	}
	if len(metadata.Description) == 0 && state.Description.IsNull() {
		newState.Description = types.StringNull()
	} else {
		newState.Description = types.StringValue(metadata.Description)
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Secret", map[string]any{
		"name":        newState.Name,
		"description": newState.Description,
		"url":         newState.URL,
	})
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state secretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretName := plan.Name.ValueString()

	// At least one of these should be set to non-nil, since TF called `Update`.
	secretReq := dtrack.UpdateSecretRequest{
		Description: nil,
		Value:       nil,
	}
	if !plan.Description.Equal(state.Description) {
		secretReq.Description = plan.Description.ValueStringPointer()
	}
	if !plan.Value.Equal(state.Value) {
		secretReq.Value = plan.Value.ValueStringPointer()
	}

	tflog.Debug(ctx, "Updating Secret", map[string]any{
		"name":              secretName,
		"plan.description":  plan.Description,
		"plan.value.#":      len(plan.Value.ValueString()),
		"state.description": state.Description,
		"state.value.#":     len(state.Value.ValueString()),
	})

	err := r.client.Secrets.Update(ctx, secretName, secretReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update secret with name: "+secretName,
			"Error from: "+err.Error(),
		)
		return
	}

	newState := secretResourceModel{
		Name:        plan.Name,
		Value:       plan.Value,
		Description: plan.Description,
		URL:         state.URL,
	}
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updating Secret", map[string]any{
		"name":        secretName,
		"description": newState.Description,
		"value.#":     len(newState.Value.ValueString()),
	})
}

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretName := state.Name.ValueString()
	tflog.Debug(ctx, "Deleting Secret", map[string]any{
		"name":        state.Name,
		"description": state.Description,
		"value.#":     len(state.Value.ValueString()),
	})
	err := r.client.Secrets.Delete(ctx, secretName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete secret with name: "+secretName,
			"Unexpected error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Secret", map[string]any{
		"name":        state.Name,
		"description": state.Description,
		"value.#":     len(state.Value.ValueString()),
	})
}

func (r *secretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
