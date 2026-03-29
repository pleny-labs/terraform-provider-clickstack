package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pleny-labs/terraform-provider-clickstack/internal/client"
)

var _ datasource.DataSource = &SourcesDataSource{}

type SourcesDataSource struct {
	client *client.Client
}

type SourcesDataSourceModel struct {
	Sources types.List `tfsdk:"sources"`
}

var sourceFromAttrTypes = map[string]attr.Type{
	"database_name": types.StringType,
	"table_name":    types.StringType,
}

var sourceAttrTypes = map[string]attr.Type{
	"id":                    types.StringType,
	"name":                  types.StringType,
	"kind":                  types.StringType,
	"connection":            types.StringType,
	"from":                  types.ObjectType{AttrTypes: sourceFromAttrTypes},
	"timestamp_value_expression":          types.StringType,
	"service_name_expression":             types.StringType,
	"severity_text_expression":            types.StringType,
	"body_expression":                     types.StringType,
	"event_attributes_expression":         types.StringType,
	"resource_attributes_expression":      types.StringType,
	"displayed_timestamp_value_expression": types.StringType,
	"metric_source_id":                    types.StringType,
	"trace_source_id":                     types.StringType,
	"trace_id_expression":                 types.StringType,
	"span_id_expression":                  types.StringType,
	"implicit_column_expression":          types.StringType,
}

func NewSourcesDataSource() datasource.DataSource {
	return &SourcesDataSource{}
}

func (d *SourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sources"
}

func (d *SourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all ClickStack data sources.",
		Attributes: map[string]schema.Attribute{
			"sources": schema.ListNestedAttribute{
				Description: "List of data sources.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Source identifier.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Source name.",
							Computed:    true,
						},
						"kind": schema.StringAttribute{
							Description: "Source kind (e.g., log, trace, metric).",
							Computed:    true,
						},
						"connection": schema.StringAttribute{
							Description: "Connection string.",
							Computed:    true,
						},
						"from": schema.SingleNestedAttribute{
							Description: "Source table reference.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"database_name": schema.StringAttribute{
									Description: "Database name.",
									Computed:    true,
								},
								"table_name": schema.StringAttribute{
									Description: "Table name.",
									Computed:    true,
								},
							},
						},
						"timestamp_value_expression": schema.StringAttribute{
							Description: "Expression for timestamp values.",
							Computed:    true,
						},
						"service_name_expression": schema.StringAttribute{
							Description: "Expression for service name.",
							Computed:    true,
						},
						"severity_text_expression": schema.StringAttribute{
							Description: "Expression for severity text.",
							Computed:    true,
						},
						"body_expression": schema.StringAttribute{
							Description: "Expression for body content.",
							Computed:    true,
						},
						"event_attributes_expression": schema.StringAttribute{
							Description: "Expression for event attributes.",
							Computed:    true,
						},
						"resource_attributes_expression": schema.StringAttribute{
							Description: "Expression for resource attributes.",
							Computed:    true,
						},
						"displayed_timestamp_value_expression": schema.StringAttribute{
							Description: "Expression for displayed timestamp.",
							Computed:    true,
						},
						"metric_source_id": schema.StringAttribute{
							Description: "Related metric source ID.",
							Computed:    true,
						},
						"trace_source_id": schema.StringAttribute{
							Description: "Related trace source ID.",
							Computed:    true,
						},
						"trace_id_expression": schema.StringAttribute{
							Description: "Expression for trace ID.",
							Computed:    true,
						},
						"span_id_expression": schema.StringAttribute{
							Description: "Expression for span ID.",
							Computed:    true,
						},
						"implicit_column_expression": schema.StringAttribute{
							Description: "Expression for implicit column.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *SourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *SourcesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	sources, err := d.client.ListSources(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading sources", err.Error())
		return
	}

	sourceValues := make([]attr.Value, len(sources))
	for i, s := range sources {
		// Build from object
		var fromObj attr.Value
		if s.From != nil {
			obj, diags := types.ObjectValue(sourceFromAttrTypes, map[string]attr.Value{
				"database_name": types.StringValue(s.From.DatabaseName),
				"table_name":    types.StringValue(s.From.TableName),
			})
			resp.Diagnostics.Append(diags...)
			fromObj = obj
		} else {
			fromObj = types.ObjectNull(sourceFromAttrTypes)
		}

		obj, diags := types.ObjectValue(sourceAttrTypes, map[string]attr.Value{
			"id":                                   types.StringValue(s.ID),
			"name":                                 types.StringValue(s.Name),
			"kind":                                 types.StringValue(s.Kind),
			"connection":                           types.StringValue(s.Connection),
			"from":                                 fromObj,
			"timestamp_value_expression":           stringOrNull(s.TimestampValueExpression),
			"service_name_expression":              stringOrNull(s.ServiceNameExpression),
			"severity_text_expression":             stringOrNull(s.SeverityTextExpression),
			"body_expression":                      stringOrNull(s.BodyExpression),
			"event_attributes_expression":          stringOrNull(s.EventAttributesExpression),
			"resource_attributes_expression":       stringOrNull(s.ResourceAttributesExpression),
			"displayed_timestamp_value_expression": stringOrNull(s.DisplayedTimestampValueExpression),
			"metric_source_id":                     stringOrNull(s.MetricSourceID),
			"trace_source_id":                      stringOrNull(s.TraceSourceID),
			"trace_id_expression":                  stringOrNull(s.TraceIDExpression),
			"span_id_expression":                   stringOrNull(s.SpanIDExpression),
			"implicit_column_expression":           stringOrNull(s.ImplicitColumnExpression),
		})
		resp.Diagnostics.Append(diags...)
		sourceValues[i] = obj
	}

	sourcesList, diags := types.ListValue(types.ObjectType{AttrTypes: sourceAttrTypes}, sourceValues)
	resp.Diagnostics.Append(diags...)

	state := SourcesDataSourceModel{Sources: sourcesList}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
