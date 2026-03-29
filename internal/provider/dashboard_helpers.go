package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pleny-labs/terraform-provider-clickstack/internal/client"
)

func expandDashboard(ctx context.Context, plan DashboardResourceModel, diags *diag.Diagnostics) client.Dashboard {
	d := client.Dashboard{
		Name: plan.Name.ValueString(),
	}

	// Tiles
	if !plan.Tiles.IsNull() {
		var tileModels []TileModel
		diags.Append(plan.Tiles.ElementsAs(ctx, &tileModels, false)...)
		for _, tm := range tileModels {
			tile := client.Tile{
				Name: tm.Name.ValueString(),
				X:    tm.X.ValueFloat64(),
				Y:    tm.Y.ValueFloat64(),
				W:    tm.W.ValueFloat64(),
				H:    tm.H.ValueFloat64(),
			}

			if !tm.Config.IsNull() {
				var configModels []TileConfigModel
				diags.Append(tm.Config.ElementsAs(ctx, &configModels, false)...)
				if len(configModels) > 0 {
					cm := configModels[0]
					tile.Config.DisplayType = cm.DisplayType.ValueString()

					if !cm.SourceID.IsNull() {
						v := cm.SourceID.ValueString()
						tile.Config.SourceID = &v
					}
					if !cm.Content.IsNull() {
						v := cm.Content.ValueString()
						tile.Config.Content = &v
					}
					if !cm.SortOrder.IsNull() {
						v := cm.SortOrder.ValueString()
						tile.Config.SortOrder = &v
					}
					if !cm.GroupBy.IsNull() {
						var groupBy []string
						diags.Append(cm.GroupBy.ElementsAs(ctx, &groupBy, false)...)
						tile.Config.GroupBy = groupBy
					}
					if !cm.Fields.IsNull() {
						var fields []string
						diags.Append(cm.Fields.ElementsAs(ctx, &fields, false)...)
						tile.Config.Fields = fields
					}
					if !cm.Select.IsNull() {
						var selectModels []SelectItemModel
						diags.Append(cm.Select.ElementsAs(ctx, &selectModels, false)...)
						for _, sm := range selectModels {
							si := client.SelectItem{
								AggFn: sm.AggFn.ValueString(),
							}
							if !sm.ValueExpression.IsNull() {
								si.ValueExpression = sm.ValueExpression.ValueString()
							}
							if !sm.Where.IsNull() {
								si.Where = sm.Where.ValueString()
							}
							if !sm.WhereLanguage.IsNull() {
								si.WhereLanguage = sm.WhereLanguage.ValueString()
							}
							if !sm.Alias.IsNull() {
								si.Alias = sm.Alias.ValueString()
							}
							if !sm.Level.IsNull() {
								v := sm.Level.ValueFloat64()
								si.Level = &v
							}
							tile.Config.Select = append(tile.Config.Select, si)
						}
					}
				}
			}

			d.Tiles = append(d.Tiles, tile)
		}
	}

	// Tags
	if !plan.Tags.IsNull() {
		var tags []string
		diags.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		d.Tags = tags
	}

	// Saved query
	if !plan.SavedQuery.IsNull() {
		v := plan.SavedQuery.ValueString()
		d.SavedQuery = &v
	}
	if !plan.SavedQueryLanguage.IsNull() {
		v := plan.SavedQueryLanguage.ValueString()
		d.SavedQueryLanguage = &v
	}

	// Filters
	if !plan.Filters.IsNull() {
		var filterModels []FilterModel
		diags.Append(plan.Filters.ElementsAs(ctx, &filterModels, false)...)
		for _, fm := range filterModels {
			d.Filters = append(d.Filters, client.Filter{
				Type:       fm.Type.ValueString(),
				Name:       fm.Name.ValueString(),
				Expression: fm.Expression.ValueString(),
				SourceID:   fm.SourceID.ValueString(),
			})
		}
	}

	// Saved filter values
	if !plan.SavedFilterValues.IsNull() {
		var sfvModels []SavedFilterValueModel
		diags.Append(plan.SavedFilterValues.ElementsAs(ctx, &sfvModels, false)...)
		for _, sfv := range sfvModels {
			sv := client.SavedFilterValue{
				Condition: sfv.Condition.ValueString(),
			}
			if !sfv.Type.IsNull() {
				sv.Type = sfv.Type.ValueString()
			}
			d.SavedFilterValues = append(d.SavedFilterValues, sv)
		}
	}

	return d
}

func flattenDashboard(ctx context.Context, d *client.Dashboard, state *DashboardResourceModel, diags *diag.Diagnostics) {
	state.Name = types.StringValue(d.Name)

	// Tiles
	tileValues := make([]attr.Value, len(d.Tiles))
	for i, t := range d.Tiles {
		// Build select items
		selectValues := make([]attr.Value, len(t.Config.Select))
		for j, si := range t.Config.Select {
			level := types.Float64Null()
			if si.Level != nil {
				level = types.Float64Value(*si.Level)
			}
			obj, d := types.ObjectValue(selectItemAttrTypes, map[string]attr.Value{
				"agg_fn":           types.StringValue(si.AggFn),
				"value_expression": stringOrNull(si.ValueExpression),
				"where":            stringOrNull(si.Where),
				"where_language":   stringOrNull(si.WhereLanguage),
				"alias":            stringOrNull(si.Alias),
				"level":            level,
			})
			diags.Append(d...)
			selectValues[j] = obj
		}
		selectList, d := types.ListValue(types.ObjectType{AttrTypes: selectItemAttrTypes}, selectValues)
		diags.Append(d...)

		// Build group_by
		groupByValues := make([]attr.Value, len(t.Config.GroupBy))
		for j, g := range t.Config.GroupBy {
			groupByValues[j] = types.StringValue(g)
		}
		groupByList, d := types.ListValue(types.StringType, groupByValues)
		diags.Append(d...)

		// Build fields
		fieldValues := make([]attr.Value, len(t.Config.Fields))
		for j, f := range t.Config.Fields {
			fieldValues[j] = types.StringValue(f)
		}
		fieldsList, d := types.ListValue(types.StringType, fieldValues)
		diags.Append(d...)

		// Build config
		configObj, d := types.ObjectValue(tileConfigAttrTypes, map[string]attr.Value{
			"display_type": types.StringValue(t.Config.DisplayType),
			"source_id":    stringPtrOrNull(t.Config.SourceID),
			"content":      stringPtrOrNull(t.Config.Content),
			"sort_order":   stringPtrOrNull(t.Config.SortOrder),
			"group_by":     groupByList,
			"fields":       fieldsList,
			"select":       selectList,
		})
		diags.Append(d...)

		configList, d := types.ListValue(types.ObjectType{AttrTypes: tileConfigAttrTypes}, []attr.Value{configObj})
		diags.Append(d...)

		// Build tile
		tileObj, d := types.ObjectValue(tileAttrTypes, map[string]attr.Value{
			"name":   types.StringValue(t.Name),
			"x":      types.Float64Value(t.X),
			"y":      types.Float64Value(t.Y),
			"w":      types.Float64Value(t.W),
			"h":      types.Float64Value(t.H),
			"config": configList,
		})
		diags.Append(d...)
		tileValues[i] = tileObj
	}
	tilesList, tileDiags := types.ListValue(types.ObjectType{AttrTypes: tileAttrTypes}, tileValues)
	diags.Append(tileDiags...)
	state.Tiles = tilesList

	// Tags
	if len(d.Tags) > 0 {
		tagValues := make([]attr.Value, len(d.Tags))
		for i, tag := range d.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		tagsList, d := types.ListValue(types.StringType, tagValues)
		diags.Append(d...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	// Saved query
	if d.SavedQuery != nil {
		state.SavedQuery = types.StringValue(*d.SavedQuery)
	} else {
		state.SavedQuery = types.StringNull()
	}
	if d.SavedQueryLanguage != nil {
		state.SavedQueryLanguage = types.StringValue(*d.SavedQueryLanguage)
	} else {
		state.SavedQueryLanguage = types.StringNull()
	}

	// Filters
	if len(d.Filters) > 0 {
		filterValues := make([]attr.Value, len(d.Filters))
		for i, f := range d.Filters {
			obj, d := types.ObjectValue(filterAttrTypes, map[string]attr.Value{
				"type":       types.StringValue(f.Type),
				"name":       types.StringValue(f.Name),
				"expression": types.StringValue(f.Expression),
				"source_id":  types.StringValue(f.SourceID),
			})
			diags.Append(d...)
			filterValues[i] = obj
		}
		filtersList, d := types.ListValue(types.ObjectType{AttrTypes: filterAttrTypes}, filterValues)
		diags.Append(d...)
		state.Filters = filtersList
	} else {
		state.Filters = types.ListNull(types.ObjectType{AttrTypes: filterAttrTypes})
	}

	// Saved filter values
	if len(d.SavedFilterValues) > 0 {
		sfvValues := make([]attr.Value, len(d.SavedFilterValues))
		for i, sfv := range d.SavedFilterValues {
			obj, d := types.ObjectValue(savedFilterValueAttrTypes, map[string]attr.Value{
				"type":      stringOrNull(sfv.Type),
				"condition": types.StringValue(sfv.Condition),
			})
			diags.Append(d...)
			sfvValues[i] = obj
		}
		sfvList, d := types.ListValue(types.ObjectType{AttrTypes: savedFilterValueAttrTypes}, sfvValues)
		diags.Append(d...)
		state.SavedFilterValues = sfvList
	} else {
		state.SavedFilterValues = types.ListNull(types.ObjectType{AttrTypes: savedFilterValueAttrTypes})
	}
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func stringPtrOrNull(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
