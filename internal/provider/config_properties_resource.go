package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

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

	configPropertyIdentity struct {
		group string
		name  string
	}
)

func NewConfigPropertiesResource() resource.Resource {
	return &configPropertiesResource{}
}

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
							Description: "Type of the Config Property. Required in API v4. Unused in API v5. See DependencyTrack for valid enum values.",
							Optional:    true,
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

	savedValues := map[configPropertyIdentity]string{}
	savedTypes := map[configPropertyIdentity]string{}
	createdValues := map[configPropertyIdentity]string{}
	createdDescriptions := map[configPropertyIdentity]string{}

	// Create SDK Request, and save values.
	configPropertiesReq := Map(plan.Properties, func(model configPropertiesElementResourceModel) dtrack.ConfigProperty {
		configProperty := dtrack.ConfigProperty{
			GroupName: model.Group.ValueString(),
			Name:      model.Name.ValueString(),
			Value:     model.Value.ValueString(),
		}
		identity := configPropertyIdentity{
			group: configProperty.GroupName,
			name:  configProperty.GroupName,
		}
		savedValues[identity] = configProperty.Value
		if !model.Type.IsNull() && !model.Type.IsUnknown() {
			savedTypes[identity] = model.Type.ValueString()
		}
		if r.semver.Major < 5 {
			// Type is not used in request in v5.
			configProperty.Type = model.Type.ValueString()
			if len(configProperty.Type) == 0 {
				resp.Diagnostics.AddError(
					"Missing Type for config property",
					"Config property with group: "+configProperty.GroupName+", and name: "+configProperty.Name+", is missing its defined Type.",
				)
			}
		}
		tflog.Debug(ctx, "Creating bulk config property", map[string]any{
			"group": configProperty.GroupName,
			"name":  configProperty.Name,
			"value": configProperty.Value,
			"type":  configProperty.Type,
		})
		return configProperty
	})
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Creating bulk config properties", map[string]any{
		"count": len(configPropertiesReq),
	})

	// Make API Request
	configPropertiesRes, err := r.client.Config.UpdateAll(ctx, configPropertiesReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring config properties.",
			"Unexpected error: "+err.Error(),
		)
		return
	}

	// Convert `[]ConfigProperty` into `map[configPropertyIdentity]string{}`'s, removing requirement for API response elements to be in the same order as request.
	for _, configPropertyResItem := range configPropertiesRes {
		var configProperty dtrack.ConfigProperty
		if r.semver.Major < 5 || configPropertyResItem.Response != nil {
			configProperty = *configPropertyResItem.Response
		} else if configPropertyResItem.Error != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("properties"),
				"Error when setting a config property",
				"API error response: "+*configPropertyResItem.Error,
			)
			continue
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("properties"),
				"Missing ConfigProperty / error within response",
				"This is a bug in interaction between Provider and SDK, and should be reported to the Provider developers.",
			)
			continue
		}
		identity := configPropertyIdentity{
			group: configProperty.GroupName,
			name:  configProperty.Name,
		}

		createdValues[identity] = configProperty.Value
		createdDescriptions[identity] = configProperty.Description
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from saved / created into Model, iterating through items in request, to ensure model order is retained.
	modelProperties := Map(configPropertiesReq, func(sdkReq dtrack.ConfigProperty) configPropertiesElementResourceModel {
		identity := configPropertyIdentity{
			group: sdkReq.GroupName,
			name:  sdkReq.Name,
		}

		model := configPropertiesElementResourceModel{
			Group:       types.StringValue(identity.group),
			Name:        types.StringValue(identity.name),
			Type:        types.StringUnknown(), // Set Below.
			Description: types.StringValue(createdDescriptions[identity]),
			Value:       types.StringUnknown(), // Set Below.
		}
		{
			savedType, ok := savedTypes[identity]
			if ok {
				model.Type = types.StringValue(savedType)
			} else {
				model.Type = types.StringNull()
			}
		}

		// API v5 replaced `EncryptedString` with a Secrets Management system.
		if r.semver.Major < 5 && sdkReq.Type == PropertyTypeEncryptedString {
			model.Value = types.StringValue(savedValues[identity])
		} else if r.semver.Major < 5 && identity.group == "vuln-source" && identity.name == "google.osv.enabled" {
			// API v5 moved this from a config property to an extension point.
			retained := savedValues[identity]
			created := createdValues[identity]

			retainedSorted := slices.SortedStableFunc(strings.SplitSeq(retained, ";"), strings.Compare)
			responseSorted := slices.SortedStableFunc(strings.SplitSeq(created, ";"), strings.Compare)
			if slices.EqualFunc(retainedSorted, responseSorted, strings.EqualFold) {
				model.Value = types.StringValue(retained)
			} else {
				model.Value = types.StringValue(created)
			}
		} else {
			created, ok := createdValues[identity]
			if ok {
				model.Value = types.StringValue(created)
			} else {
				model.Value = types.StringValue(savedValues[identity])
			}
		}
		tflog.Debug(ctx, "Created bulk config property", map[string]any{
			"group": model.Group,
			"name":  model.Name,
			"value": model.Value,
			"type":  model.Type,
		})

		return model
	})

	newState := configPropertiesResourceModel{
		Properties: modelProperties,
	}

	diags = resp.State.Set(ctx, &newState)
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

	readProperties := map[configPropertyIdentity]dtrack.ConfigProperty{}

	for _, configProperty := range configPropertiesAll {
		identity := configPropertyIdentity{
			group: configProperty.GroupName,
			name:  configProperty.Name,
		}
		readProperties[identity] = configProperty
	}

	newProperties := Map(state.Properties, func(old configPropertiesElementResourceModel) configPropertiesElementResourceModel {
		identity := configPropertyIdentity{
			group: old.Group.ValueString(),
			name:  old.Name.ValueString(),
		}
		model := configPropertiesElementResourceModel{
			Group:       types.StringValue(identity.group),
			Name:        types.StringValue(identity.name),
			Type:        types.StringUnknown(), // Set Below.
			Description: types.StringUnknown(), // Set Below.
			Value:       types.StringUnknown(), // Set Below.
		}
		{
			read, ok := readProperties[identity]
			if !ok {
				resp.Diagnostics.AddAttributeError(
					path.Root("properties"),
					"Within Read, cannot find property within list of all properties",
					"Non-existent property in group: "+identity.group+", with name: "+identity.name,
				)
				model.Type = old.Type
				model.Description = old.Description
				model.Value = old.Value
				return model
			}
			if read.Type != old.Type.ValueString() {
				resp.Diagnostics.AddAttributeError(
					path.Root("properties"),
					"Within Read, unable to match config property type",
					"Group: "+identity.group+", Name: "+identity.name+", Type: "+read.Type,
				)
				return model
			}
			model.Type = types.StringValue(read.Type)
			model.Description = types.StringValue(read.Description)
			if r.semver.Major < 5 && read.Type == PropertyTypeEncryptedString {
				model.Value = old.Value
			} else if r.semver.Major < 5 && identity.group == "vuln-source" && identity.name == "google.osv.enabled" {
				retained := old.Value.ValueString()
				retainedSorted := slices.SortedStableFunc(strings.SplitSeq(retained, ";"), strings.Compare)
				responseSorted := slices.SortedStableFunc(strings.SplitSeq(read.Value, ";"), strings.Compare)
				if slices.EqualFunc(retainedSorted, responseSorted, strings.EqualFold) {
					model.Value = old.Value
				} else {
					model.Value = types.StringValue(read.Value)
				}
			} else {
				model.Value = types.StringValue(read.Value)
			}
		}

		return model
	})
	if resp.Diagnostics.HasError() {
		return
	}

	newState := configPropertiesResourceModel{
		Properties: newProperties,
	}

	diags = resp.State.Set(ctx, &newState)
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

	configPropertiesReq := make([]dtrack.ConfigProperty, 0, len(plan.Properties))
	valueStringRetention := map[configPropertyIdentity]string{}

	for _, propertyReq := range plan.Properties {
		configProperty := dtrack.ConfigProperty{
			GroupName: propertyReq.Group.ValueString(),
			Name:      propertyReq.Name.ValueString(),
			Value:     propertyReq.Value.ValueString(),
			Type:      propertyReq.Type.ValueString(),
		}
		identity := configPropertyIdentity{
			group: configProperty.GroupName,
			name:  configProperty.Name,
		}
		if r.semver.Major < 5 && configProperty.Type == PropertyTypeEncryptedString {
			valueStringRetention[identity] = configProperty.Value
		} else if r.semver.Major < 5 && configProperty.GroupName == "vuln-source" && configProperty.Name == "google.osv.enabled" {
			valueStringRetention[identity] = configProperty.Value
		}

		configPropertiesReq = append(configPropertiesReq, configProperty)
		tflog.Debug(ctx, "Updating bulk config property", map[string]any{
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

	for _, propertyResItem := range configPropertiesRes {
		var propertyRes dtrack.ConfigProperty
		if r.semver.Major < 5 {
			propertyRes = *propertyResItem.Response
		} else if propertyResItem.Response != nil {
			propertyRes = *propertyResItem.Response
		} else if propertyResItem.Error != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("properties"),
				"Error within updating config property",
				*propertyResItem.Error,
			)
			continue
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("properties"),
				"Missing ConfigProperty / error within response",
				"This is a bug in interaction between Provider and SDK, and should be reported to the Provider developers.",
			)
			continue
		}

		model := configPropertiesElementResourceModel{
			Group:       types.StringValue(propertyRes.GroupName),
			Name:        types.StringValue(propertyRes.Name),
			Value:       types.StringValue(propertyRes.Value),
			Type:        types.StringValue(propertyRes.Type),
			Description: types.StringValue(propertyRes.Description),
		}

		identity := configPropertyIdentity{
			group: propertyRes.GroupName,
			name:  propertyRes.Name,
		}
		if r.semver.Major < 5 && propertyRes.Type == PropertyTypeEncryptedString {
			model.Value = types.StringValue(valueStringRetention[identity])
		}
		if r.semver.Major < 5 && propertyRes.GroupName == "vuln-source" && propertyRes.Name == "google.osv.enabled" {
			requested := valueStringRetention[identity]
			requestedSorted := slices.SortedStableFunc(strings.SplitSeq(requested, ";"), strings.Compare)
			responseSorted := slices.SortedStableFunc(strings.SplitSeq(propertyRes.Value, ";"), strings.Compare)
			if slices.EqualFunc(requestedSorted, responseSorted, strings.EqualFold) {
				model.Value = types.StringValue(requested)
			}
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
