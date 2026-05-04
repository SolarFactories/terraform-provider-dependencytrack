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
	_ datasource.DataSource              = &licenseGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &licenseGroupDataSource{}
)

type (
	licenseGroupDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	licenseGroupDataSourceModel struct {
		ID         types.String                         `tfsdk:"id"`
		Name       types.String                         `tfsdk:"name"`
		Licenses   []licenseGroupDataSourceModelLicense `tfsdk:"licenses"`
		RiskWeight types.Int32                          `tfsdk:"risk_weight"`
	}

	licenseGroupDataSourceModelLicense struct {
		UUID   types.String `tfsdk:"uuid"`
		Name   types.String `tfsdk:"name"`
		SpdxID types.String `tfsdk:"spdx_id"`
	}
)

func NewLicenseGroupDataSource() datasource.DataSource {
	return &licenseGroupDataSource{}
}

func (*licenseGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license_group"
}

func (*licenseGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch a license group by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the license group.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the license group.",
				Required:    true,
			},
			"risk_weight": schema.Int32Attribute{
				Description: "Risk Weight of the license group.",
				Computed:    true,
			},
			"licenses": schema.ListNestedAttribute{
				Description: "List of licenses in license group.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Description: "UUID for the license.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Full name of the license.",
							Computed:    true,
						},
						"spdx_id": schema.StringAttribute{
							Description: "SPDX ID of the license, as per https://spdx.org/licenses",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *licenseGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state licenseGroupDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	tflog.Debug(ctx, "Reading license group", map[string]any{
		"name": name,
	})

	group, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.LicenseGroup], error) {
			return d.client.LicenseGroup.GetAll(ctx, po, dtrack.SortOptions{})
		},
		func(group dtrack.LicenseGroup) bool {
			return group.Name == name
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch License Group.",
			"Error for License Group: "+name+",in original error: "+err.Error(),
		)
		return
	}

	newState := licenseGroupDataSourceModel{
		ID:         types.StringValue(group.UUID.String()),
		Name:       types.StringValue(group.Name),
		RiskWeight: types.Int32Value(group.RiskWeight),
		Licenses: Map(group.Licenses, func(license dtrack.License) licenseGroupDataSourceModelLicense {
			return licenseGroupDataSourceModelLicense{
				UUID:   types.StringValue(license.UUID.String()),
				Name:   types.StringValue(license.Name),
				SpdxID: types.StringValue(license.LicenseID),
			}
		}),
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read License Group", map[string]any{
		"id":          newState.ID,
		"name":        newState.Name,
		"risk_weight": newState.RiskWeight,
		"licenses.#":  len(newState.Licenses),
	})
}

func (d *licenseGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
