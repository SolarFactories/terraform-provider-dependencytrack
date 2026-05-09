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

// Interface impl check.
var (
	_ datasource.DataSource              = &licenseDataSource{}
	_ datasource.DataSourceWithConfigure = &licenseDataSource{}
)

type (
	licenseDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	licenseDataSourceModel struct {
		ID                    types.String   `tfsdk:"id"`
		UUID                  types.String   `tfsdk:"uuid"`
		Name                  types.String   `tfsdk:"name"`
		Text                  types.String   `tfsdk:"text"`
		Template              types.String   `tfsdk:"template"`
		Header                types.String   `tfsdk:"header"`
		Comment               types.String   `tfsdk:"comment"`
		SeeAlso               []types.String `tfsdk:"see_also"`
		OSIApproved           types.Bool     `tfsdk:"osi_approved"`
		FSFLibre              types.Bool     `tfsdk:"fsf_libre"`
		IsDeprecatedLicenseID types.Bool     `tfsdk:"is_deprecated_license_id"`
	}
)

func NewLicenseDataSource() datasource.DataSource {
	return &licenseDataSource{}
}

func (*licenseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (*licenseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch a license by SPDX ID",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "SPDX ID of the license.",
				Required:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "UUID of the license.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the license.",
				Computed:    true,
			},
			"text": schema.StringAttribute{
				Description: "Text of the license.",
				Computed:    true,
			},
			"template": schema.StringAttribute{
				Description: "Template of the license.",
				Computed:    true,
			},
			"header": schema.StringAttribute{
				Description: "Header of the license.",
				Computed:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Comment on the license.",
				Computed:    true,
			},
			"osi_approved": schema.BoolAttribute{
				Description: "Whether the license is approved by Open Source Initiative.",
				Computed:    true,
			},
			"fsf_libre": schema.BoolAttribute{
				Description: "Whether the license is considered libre by Free Software Foundation.",
				Computed:    true,
			},
			"is_deprecated_license_id": schema.BoolAttribute{
				Description: "Whether the License ID used is deprecated.",
				Computed:    true,
			},
			"see_also": schema.ListAttribute{
				Description: "Links to external information about license.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *licenseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state licenseDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	spdx := state.ID.ValueString()

	tflog.Debug(ctx, "Reading license", map[string]any{
		"id": spdx,
	})

	license, err := d.client.License.Get(ctx, spdx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch License.",
			"Error for License: "+spdx+", in original error: "+err.Error(),
		)
		return
	}

	newState := licenseDataSourceModel{
		ID:                    types.StringValue(license.LicenseID),
		UUID:                  types.StringValue(license.UUID.String()),
		Name:                  types.StringValue(license.Name),
		Text:                  types.StringValue(license.Text),
		Template:              types.StringValue(license.Template),
		Header:                types.StringValue(license.Header),
		Comment:               types.StringValue(license.Comment),
		OSIApproved:           types.BoolValue(license.OSIApproved),
		FSFLibre:              types.BoolValue(license.FSFLibre),
		IsDeprecatedLicenseID: types.BoolValue(license.DeprecatedLicenseID),
		SeeAlso:               Map(license.SeeAlso, types.StringValue),
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read License", map[string]any{
		"id":   newState.ID,
		"uuid": newState.UUID,
		"name": newState.Name,
	})
}

func (d *licenseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
