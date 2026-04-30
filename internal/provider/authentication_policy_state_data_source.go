package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nokia/eda/apps/terraform-provider-aaa/internal/datasource_authentication_policy_state"
	"github.com/nokia/eda/apps/terraform-provider-aaa/internal/eda/apiclient"
	"github.com/nokia/eda/apps/terraform-provider-aaa/internal/tfutils"
)

const read_ds_authenticationPolicyState = "/apps/aaa.eda.nokia.com/v1/namespaces/{namespace}/authenticationpolicystates/{name}"

var (
	_ datasource.DataSource              = (*authenticationPolicyStateDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*authenticationPolicyStateDataSource)(nil)
)

func NewAuthenticationPolicyStateDataSource() datasource.DataSource {
	return &authenticationPolicyStateDataSource{}
}

type authenticationPolicyStateDataSource struct {
	client *apiclient.EdaApiClient
}

func (d *authenticationPolicyStateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authentication_policy_state"
}

func (d *authenticationPolicyStateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_authentication_policy_state.AuthenticationPolicyStateDataSourceSchema(ctx)
}

func (d *authenticationPolicyStateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_authentication_policy_state.AuthenticationPolicyStateModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract query params from Terraform model
	queryParams, err := tfutils.ModelToStringMap(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error extracting query params", err.Error())
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Read()::API request", map[string]any{
		"path":  read_ds_authenticationPolicyState,
		"data":  spew.Sdump(data),
		"query": queryParams,
	})

	t0 := time.Now()
	result := map[string]any{}
	err = d.client.GetByQuery(ctx, read_ds_authenticationPolicyState, map[string]string{
		"namespace": tfutils.StringValue(data.Namespace),
		"name":      tfutils.StringValue(data.Name),
	}, queryParams, &result)

	tflog.Info(ctx, "Read()::API returned", map[string]any{
		"path":      read_ds_authenticationPolicyState,
		"result":    spew.Sdump(result),
		"timeTaken": time.Since(t0).String(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Error reading resource", err.Error())
		return
	}

	// Convert API response to Terraform model
	err = tfutils.AnyMapToModel(ctx, result, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build response from API result", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (r *authenticationPolicyStateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.EdaApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.EdaApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}
