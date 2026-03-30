package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pleny-labs/terraform-provider-clickstack/internal/client"
)

var (
	_ resource.Resource                = &DashboardResource{}
	_ resource.ResourceWithImportState = &DashboardResource{}
)

type DashboardResource struct {
	client *client.Client
}

type DashboardResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Tiles              types.List   `tfsdk:"tiles"`
	Tags               types.List   `tfsdk:"tags"`
	Filters            types.List   `tfsdk:"filters"`
	SavedQuery         types.String `tfsdk:"saved_query"`
	SavedQueryLanguage types.String `tfsdk:"saved_query_language"`
	SavedFilterValues  types.List   `tfsdk:"saved_filter_values"`
}

type TileModel struct {
	X      types.Float64 `tfsdk:"x"`
	Y      types.Float64 `tfsdk:"y"`
	W      types.Float64 `tfsdk:"w"`
	H      types.Float64 `tfsdk:"h"`
	Config types.List     `tfsdk:"config"`
}

type TileConfigModel struct {
	Name          types.String `tfsdk:"name"`
	DisplayType   types.String `tfsdk:"display_type"`
	Source        types.String `tfsdk:"source"`
	GroupBy       types.String `tfsdk:"group_by"`
	Where         types.String `tfsdk:"where"`
	WhereLanguage types.String `tfsdk:"where_language"`
	Granularity   types.String `tfsdk:"granularity"`
	Content       types.String `tfsdk:"content"`
	SortOrder     types.String `tfsdk:"sort_order"`
	Fields        types.List   `tfsdk:"fields"`
	Select        types.List   `tfsdk:"select"`
}

type SelectItemModel struct {
	AggFn                types.String  `tfsdk:"agg_fn"`
	ValueExpression      types.String  `tfsdk:"value_expression"`
	AggCondition         types.String  `tfsdk:"agg_condition"`
	AggConditionLanguage types.String  `tfsdk:"agg_condition_language"`
	Alias                types.String  `tfsdk:"alias"`
	Level                types.Float64 `tfsdk:"level"`
}

type FilterModel struct {
	Type       types.String `tfsdk:"type"`
	Name       types.String `tfsdk:"name"`
	Expression types.String `tfsdk:"expression"`
	SourceID   types.String `tfsdk:"source_id"`
}

type SavedFilterValueModel struct {
	Type      types.String `tfsdk:"type"`
	Condition types.String `tfsdk:"condition"`
}

func NewDashboardResource() resource.Resource {
	return &DashboardResource{}
}

func (r *DashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

var selectItemAttrTypes = map[string]attr.Type{
	"agg_fn":                 types.StringType,
	"value_expression":       types.StringType,
	"agg_condition":          types.StringType,
	"agg_condition_language": types.StringType,
	"alias":                  types.StringType,
	"level":                  types.Float64Type,
}

var tileConfigAttrTypes = map[string]attr.Type{
	"name":           types.StringType,
	"display_type":   types.StringType,
	"source":         types.StringType,
	"group_by":       types.StringType,
	"where":          types.StringType,
	"where_language": types.StringType,
	"granularity":    types.StringType,
	"content":        types.StringType,
	"sort_order":     types.StringType,
	"fields":         types.ListType{ElemType: types.StringType},
	"select":         types.ListType{ElemType: types.ObjectType{AttrTypes: selectItemAttrTypes}},
}

var tileAttrTypes = map[string]attr.Type{
	"x":      types.Float64Type,
	"y":      types.Float64Type,
	"w":      types.Float64Type,
	"h":      types.Float64Type,
	"config": types.ListType{ElemType: types.ObjectType{AttrTypes: tileConfigAttrTypes}},
}

var filterAttrTypes = map[string]attr.Type{
	"type":       types.StringType,
	"name":       types.StringType,
	"expression": types.StringType,
	"source_id":  types.StringType,
}

var savedFilterValueAttrTypes = map[string]attr.Type{
	"type":      types.StringType,
	"condition": types.StringType,
}

func (r *DashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ClickStack dashboard.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the dashboard.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the dashboard.",
				Required:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags for the dashboard.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"saved_query": schema.StringAttribute{
				Description: "Default saved query for the dashboard.",
				Optional:    true,
			},
			"saved_query_language": schema.StringAttribute{
				Description: "Language for the saved query: sql or lucene.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("sql", "lucene"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"tiles": schema.ListNestedBlock{
				Description: "Dashboard tiles/charts.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"x": schema.Float64Attribute{
							Description: "Horizontal position.",
							Required:    true,
						},
						"y": schema.Float64Attribute{
							Description: "Vertical position.",
							Required:    true,
						},
						"w": schema.Float64Attribute{
							Description: "Width in grid units.",
							Required:    true,
						},
						"h": schema.Float64Attribute{
							Description: "Height in grid units.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"config": schema.ListNestedBlock{
							Description: "Tile configuration (exactly one block).",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Tile display name.",
										Required:    true,
									},
									"display_type": schema.StringAttribute{
										Description: "Chart type: line, table, number, search, or markdown.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("line", "table", "number", "search", "markdown"),
										},
									},
									"source": schema.StringAttribute{
										Description: "Data source identifier.",
										Optional:    true,
									},
									"group_by": schema.StringAttribute{
										Description: "Field to group results by.",
										Optional:    true,
									},
									"where": schema.StringAttribute{
										Description: "Filter condition for the tile.",
										Optional:    true,
									},
									"where_language": schema.StringAttribute{
										Description: "Filter language: sql or lucene.",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("sql", "lucene"),
										},
									},
									"granularity": schema.StringAttribute{
										Description: "Time granularity (e.g. 5 minute, 1 hour).",
										Optional:    true,
									},
									"content": schema.StringAttribute{
										Description: "Content for markdown tiles.",
										Optional:    true,
									},
									"sort_order": schema.StringAttribute{
										Description: "Sort order: asc or desc.",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("asc", "desc"),
										},
									},
									"fields": schema.ListAttribute{
										Description: "Fields for search-type displays.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"select": schema.ListNestedBlock{
										Description: "Aggregation specifications.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"agg_fn": schema.StringAttribute{
													Description: "Aggregation function: avg, count, max, min, sum, count_distinct, last_value, quantile.",
													Required:    true,
													Validators: []validator.String{
														stringvalidator.OneOf("avg", "count", "max", "min", "sum", "count_distinct", "last_value", "quantile"),
													},
												},
												"value_expression": schema.StringAttribute{
													Description: "Column or expression to aggregate.",
													Optional:    true,
												},
												"agg_condition": schema.StringAttribute{
													Description: "Aggregation filter condition (e.g. SeverityText:ERROR).",
													Optional:    true,
												},
												"agg_condition_language": schema.StringAttribute{
													Description: "Language for agg_condition: sql or lucene.",
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.OneOf("sql", "lucene"),
													},
												},
												"alias": schema.StringAttribute{
													Description: "Display name for the aggregation.",
													Optional:    true,
												},
												"level": schema.Float64Attribute{
													Description: "Percentile level for quantile aggregation (0.5-0.99).",
													Optional:    true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"filters": schema.ListNestedBlock{
				Description: "Dashboard-level filters.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Filter type (QUERY_EXPRESSION).",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Human-readable filter label.",
							Required:    true,
						},
						"expression": schema.StringAttribute{
							Description: "Column/field for filtering.",
							Required:    true,
						},
						"source_id": schema.StringAttribute{
							Description: "Associated data source.",
							Required:    true,
						},
					},
				},
			},
			"saved_filter_values": schema.ListNestedBlock{
				Description: "Persistent filter defaults.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Filter type (defaults to sql).",
							Optional:    true,
						},
						"condition": schema.StringAttribute{
							Description: "SQL filter expression.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *DashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiDashboard := expandDashboard(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateDashboard(ctx, apiDashboard)
	if err != nil {
		resp.Diagnostics.AddError("Error creating dashboard", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboard, err := r.client.GetDashboard(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading dashboard", err.Error())
		return
	}

	flattenDashboard(ctx, dashboard, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiDashboard := expandDashboard(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateDashboard(ctx, state.ID.ValueString(), apiDashboard)
	if err != nil {
		resp.Diagnostics.AddError("Error updating dashboard", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDashboard(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting dashboard", err.Error())
		return
	}
}

func (r *DashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	dashboard, err := r.client.GetDashboard(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing dashboard", err.Error())
		return
	}

	var state DashboardResourceModel
	state.ID = types.StringValue(dashboard.ID)
	flattenDashboard(ctx, dashboard, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
