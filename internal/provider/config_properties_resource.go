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
	_ resource.Resource              = &configPropertiesResource{}
	_ resource.ResourceWithConfigure = &configPropertiesResource{}
)

func NewConfigPropertiesResource() resource.Resource {
	return &configPropertiesResource{}
}

type (
	configPropertiesResource struct {
		client *dtrack.Client
		semver *Semver
	}

	configPropertiesResourceModel struct {
		Properties []configPropertiesElementResourceModel `tfsdk:"properties"`
	}

	configPropertiesElementResourceModel struct {
		Group       types.String `tfsdk:"group"`
		Name        types.String `tfsdk:"name"`
		Value       types.String `tfsdk:"value"`
		Type        types.String `tfsdk:"type"`
		Description types.String `tfsdk:"description"`
	}
)

func (*configPropertiesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_properties"
}

func (*configPropertiesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages multiple Config Properties.",
		Attributes: map[string]schema.Attribute{
			"properties": schema.ListNestedAttribute{
				Description: "Config properties, to be bulk managed.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							Description: "Group name of the Config Property.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"name": schema.StringAttribute{
							Description: "Property name of the Config Property.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"value": schema.StringAttribute{
							Description: "Value of the Config Property.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of the Config Property. See DependencyTrack for valid enum values.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"description": schema.StringAttribute{
							Description: "Description of the Config Property.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *configPropertiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configPropertiesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	type Identity struct {
		group string
		name  string
	}

	configPropertiesReq := make([]dtrack.ConfigProperty, 0, len(plan.Properties))
	encryptedStringRetention := map[Identity]string{}

	for _, propertyReq := range plan.Properties {
		configProperty := dtrack.ConfigProperty{
			GroupName: propertyReq.Group.ValueString(),
			Name:      propertyReq.Name.ValueString(),
			Value:     propertyReq.Value.ValueString(),
			Type:      propertyReq.Type.ValueString(),
		}
		if configProperty.Type == PropertyTypeEncryptedString {
			encryptedStringRetention[Identity{
				group: configProperty.GroupName,
				name:  configProperty.Name,
			}] = configProperty.Value
		}

		configPropertiesReq = append(configPropertiesReq, configProperty)
		tflog.Debug(ctx, "Creating bulk config property", map[string]any{
			"group": configProperty.GroupName,
			"name":  configProperty.Name,
			"value": configProperty.Value,
			"type":  configProperty.Type,
		})
	}

	tflog.Debug(ctx, "Creating bulk config properties", map[string]any{
		"count": len(configPropertiesReq),
	})
	configPropertiesRes, err := r.client.Config.UpdateAll(ctx, configPropertiesReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring config properties.",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	configPropertiesState := configPropertiesResourceModel{
		Properties: []configPropertiesElementResourceModel{},
	}

	for _, propertyRes := range configPropertiesRes {
		model := configPropertiesElementResourceModel{
			Group:       types.StringValue(propertyRes.GroupName),
			Name:        types.StringValue(propertyRes.Name),
			Value:       types.StringValue(propertyRes.Value),
			Type:        types.StringValue(propertyRes.Type),
			Description: types.StringValue(propertyRes.Description),
		}
		if propertyRes.Type == PropertyTypeEncryptedString {
			model.Value = types.StringValue(encryptedStringRetention[Identity{
				group: propertyRes.GroupName,
				name:  propertyRes.Name,
			}])
		}
		configPropertiesState.Properties = append(configPropertiesState.Properties, model)
		tflog.Debug(ctx, "Created bulk config property", map[string]any{
			"group": propertyRes.GroupName,
			"name":  propertyRes.Name,
			"value": propertyRes.Value,
			"type":  propertyRes.Type,
		})
	}

	diags = resp.State.Set(ctx, &configPropertiesState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created bulk config properties", map[string]any{
		"count": len(configPropertiesRes),
	})
}

func (r *configPropertiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configPropertiesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading all config properties")
	configPropertiesAll, err := r.client.Config.GetAll(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("properties"),
			"Within Read, unable to fetch config properties.",
			"Error from: "+err.Error(),
		)
		return
	}

	for idx, configPropertyModel := range state.Properties {
		groupName := configPropertyModel.Group.ValueString()
		propertyName := configPropertyModel.Name.ValueString()
		propertyType := configPropertyModel.Type.ValueString()
		value := configPropertyModel.Value
		configProperty, err := Find(configPropertiesAll, func(configProperty dtrack.ConfigProperty) bool {
			if configProperty.GroupName != groupName {
				return false
			}
			if configProperty.Name != propertyName {
				return false
			}
			return true
		})
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("properties"),
				"Within Read, unable to match config properties: "+groupName+" "+propertyName,
				"Error from: "+err.Error(),
			)
			continue
		}
		if configProperty.Type != propertyType {
			resp.Diagnostics.AddAttributeError(
				path.Root("properties"),
				"Within Read, unable to match config property type",
				"Group: "+groupName+", Name: "+propertyName+", Type: "+configProperty.Type,
			)
			continue
		}
		state.Properties[idx] = configPropertiesElementResourceModel{
			Group:       types.StringValue(configProperty.GroupName),
			Name:        types.StringValue(configProperty.Name),
			Value:       types.StringValue(configProperty.Value),
			Type:        types.StringValue(configProperty.Type),
			Description: types.StringValue(configProperty.Description),
		}
		if configProperty.Type == PropertyTypeEncryptedString {
			state.Properties[idx].Value = value
		}
		tflog.Debug(ctx, "Read bulk config property", map[string]any{
			"group":       state.Properties[idx].Group.ValueString(),
			"name":        state.Properties[idx].Name.ValueString(),
			"value":       state.Properties[idx].Value.ValueString(),
			"type":        state.Properties[idx].Type.ValueString(),
			"description": state.Properties[idx].Description.ValueString(),
		})
	}
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read bulk config properties", map[string]any{
		"count": len(state.Properties),
	})
}

func (r *configPropertiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configPropertiesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	type Identity struct {
		group string
		name  string
	}

	configPropertiesReq := make([]dtrack.ConfigProperty, 0, len(plan.Properties))
	encryptedStringRetention := map[Identity]string{}

	for _, propertyReq := range plan.Properties {
		configProperty := dtrack.ConfigProperty{
			GroupName: propertyReq.Group.ValueString(),
			Name:      propertyReq.Name.ValueString(),
			Value:     propertyReq.Value.ValueString(),
			Type:      propertyReq.Type.ValueString(),
		}
		if configProperty.Type == PropertyTypeEncryptedString {
			encryptedStringRetention[Identity{
				group: configProperty.GroupName,
				name:  configProperty.Name,
			}] = configProperty.Value
		}
		configPropertiesReq = append(configPropertiesReq, configProperty)
		tflog.Debug(ctx, "Updating bulk config properties", map[string]any{
			"group": configProperty.GroupName,
			"name":  configProperty.Name,
			"value": configProperty.Value,
			"type":  configProperty.Type,
		})
	}

	tflog.Debug(ctx, "Updating bulk config properties", map[string]any{
		"count": len(configPropertiesReq),
	})
	configPropertiesRes, err := r.client.Config.UpdateAll(ctx, configPropertiesReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring config properties.",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	configPropertiesState := configPropertiesResourceModel{
		Properties: []configPropertiesElementResourceModel{},
	}

	for _, propertyRes := range configPropertiesRes {
		model := configPropertiesElementResourceModel{
			Group:       types.StringValue(propertyRes.GroupName),
			Name:        types.StringValue(propertyRes.Name),
			Value:       types.StringValue(propertyRes.Value),
			Type:        types.StringValue(propertyRes.Type),
			Description: types.StringValue(propertyRes.Description),
		}
		if propertyRes.Type == PropertyTypeEncryptedString {
			model.Value = types.StringValue(encryptedStringRetention[Identity{
				group: propertyRes.GroupName,
				name:  propertyRes.Name,
			}])
		}
		configPropertiesState.Properties = append(configPropertiesState.Properties, model)
		tflog.Debug(ctx, "Updated bulk config property", map[string]any{
			"group":       model.Group.ValueString(),
			"name":        model.Name.ValueString(),
			"value":       model.Type.ValueString(),
			"type":        model.Type.ValueString(),
			"description": model.Description.ValueString(),
		})
	}
	diags = resp.State.Set(ctx, configPropertiesState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated bulk config properties", map[string]any{
		"count": len(configPropertiesState.Properties),
	})
}

func (r *configPropertiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state configPropertiesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configPropertiesReq := make([]dtrack.ConfigProperty, 0, len(state.Properties))

	for _, propertyReq := range state.Properties {
		configProperty := dtrack.ConfigProperty{
			GroupName: propertyReq.Group.ValueString(),
			Name:      propertyReq.Name.ValueString(),
			Value:     "",
			Type:      propertyReq.Type.ValueString(),
		}
		configPropertiesReq = append(configPropertiesReq, configProperty)
		tflog.Debug(ctx, "Deleting bulk config property", map[string]any{
			"group": configProperty.GroupName,
			"name":  configProperty.Name,
			"value": configProperty.Value,
			"type":  configProperty.Type,
		})
	}

	tflog.Debug(ctx, "Deleting bulk config properties", map[string]any{
		"count": len(configPropertiesReq),
	})
	configPropertiesRes, err := r.client.Config.UpdateAll(ctx, configPropertiesReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bulk config properties.",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted bulk config properties", map[string]any{
		"count": len(configPropertiesRes),
	})
}

func (r *configPropertiesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
