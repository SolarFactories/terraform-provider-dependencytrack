package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &secretMetadataDataSource{}
	_ datasource.DataSourceWithConfigure = &secretMetadataDataSource{}
)

type (
	secretMetadataDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	secretMetadataDataSourceModel struct {
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
		CreatedAt   types.Int64  `tfsdk:"created_at"`
		UpdatedAt   types.Int64  `tfsdk:"updated_at"`
	}
)

func NewSecretMetadataDataSource() datasource.DataSource {
	return &secretMetadataDataSource{}
}

func (*secretMetadataDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_metadata"
}

func (*secretMetadataDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch metadata for a Secret. Requires API >= 5.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the Secret.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Secret.",
				Computed:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "Millisecond timestamp of when the Secret was created.",
				Computed:    true,
			},
			"updated_at": schema.Int64Attribute{
				Description: "Millisecond timestamp of when the Secret was most-recently updated, if at all.",
				Computed:    true,
			},
		},
	}
}

func (d *secretMetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config secretMetadataDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretName := config.Name.ValueString()

	tflog.Debug(ctx, "Reading SecretMetadata", map[string]any{
		"name": secretName,
	})

	metadata, err := d.client.Secrets.GetMetadata(ctx, secretName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Secret Metadata for secret: "+secretName,
			"Error from: "+err.Error(),
		)
		return
	}
	state := secretMetadataDataSourceModel{
		Name:        types.StringValue(metadata.Name),
		Description: types.StringUnknown(), // Set Below.
		CreatedAt:   types.Int64Unknown(),  // Set Below.
		UpdatedAt:   types.Int64Unknown(),  // Set Below.
	}
	if len(metadata.Description) == 0 {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(metadata.Description)
	}
	if metadata.CreatedAt != nil {
		state.CreatedAt = types.Int64Value(int64(*metadata.CreatedAt))
	} else {
		state.CreatedAt = types.Int64Null()
	}
	if metadata.UpdatedAt != nil {
		state.UpdatedAt = types.Int64Value(int64(*metadata.UpdatedAt))
	} else {
		state.UpdatedAt = types.Int64Null()
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Secret Metadata", map[string]any{
		"name":        state.Name.ValueString(),
		"description": state.Description,
		"created_at":  state.CreatedAt,
		"updated_at":  state.UpdatedAt,
	})
}

func (d *secretMetadataDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = clientInfoData.client
	d.semver = clientInfoData.semver
}
