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
				X: tm.X.ValueFloat64(),
				Y: tm.Y.ValueFloat64(),
				W: tm.W.ValueFloat64(),
				H: tm.H.ValueFloat64(),
			}

			if !tm.Config.IsNull() {
				var configModels []TileConfigModel
				diags.Append(tm.Config.ElementsAs(ctx, &configModels, false)...)
				if len(configModels) > 0 {
					cm := configModels[0]
					tile.Config.Name = cm.Name.ValueString()
					tile.Config.DisplayType = cm.DisplayType.ValueString()

					if !cm.Source.IsNull() {
						tile.Config.Source = cm.Source.ValueString()
					}
					if !cm.GroupBy.IsNull() {
						tile.Config.GroupBy = cm.GroupBy.ValueString()
					}
					if !cm.Where.IsNull() {
						tile.Config.Where = cm.Where.ValueString()
					}
					if !cm.WhereLanguage.IsNull() {
						tile.Config.WhereLanguage = cm.WhereLanguage.ValueString()
					}
					if !cm.Granularity.IsNull() {
						tile.Config.Granularity = cm.Granularity.ValueString()
					}
					if !cm.Content.IsNull() {
						v := cm.Content.ValueString()
						tile.Config.Content = &v
					}
					if !cm.SortOrder.IsNull() {
						v := cm.SortOrder.ValueString()
						tile.Config.SortOrder = &v
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
							if !sm.AggCondition.IsNull() {
								si.AggCondition = sm.AggCondition.ValueString()
							}
							if !sm.AggConditionLanguage.IsNull() {
								si.AggConditionLanguage = sm.AggConditionLanguage.ValueString()
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
			obj, objDiags := types.ObjectValue(selectItemAttrTypes, map[string]attr.Value{
				"agg_fn":                 types.StringValue(si.AggFn),
				"value_expression":       stringOrNull(si.ValueExpression),
				"agg_condition":          stringOrNull(si.AggCondition),
				"agg_condition_language": stringOrNull(si.AggConditionLanguage),
				"alias":                  stringOrNull(si.Alias),
				"level":                  level,
			})
			diags.Append(objDiags...)
			selectValues[j] = obj
		}
		selectList, slDiags := types.ListValue(types.ObjectType{AttrTypes: selectItemAttrTypes}, selectValues)
		diags.Append(slDiags...)

		// Build fields
		fieldValues := make([]attr.Value, len(t.Config.Fields))
		for j, f := range t.Config.Fields {
			fieldValues[j] = types.StringValue(f)
		}
		fieldsList, flDiags := types.ListValue(types.StringType, fieldValues)
		diags.Append(flDiags...)

		// Build config
		configObj, cfgDiags := types.ObjectValue(tileConfigAttrTypes, map[string]attr.Value{
			"name":           types.StringValue(t.Config.Name),
			"display_type":   types.StringValue(t.Config.DisplayType),
			"source":         stringOrNull(t.Config.Source),
			"group_by":       stringOrNull(t.Config.GroupBy),
			"where":          stringOrNull(t.Config.Where),
			"where_language": stringOrNull(t.Config.WhereLanguage),
			"granularity":    stringOrNull(t.Config.Granularity),
			"content":        stringPtrOrNull(t.Config.Content),
			"sort_order":     stringPtrOrNull(t.Config.SortOrder),
			"fields":         fieldsList,
			"select":         selectList,
		})
		diags.Append(cfgDiags...)

		configList, clDiags := types.ListValue(types.ObjectType{AttrTypes: tileConfigAttrTypes}, []attr.Value{configObj})
		diags.Append(clDiags...)

		// Build tile
		tileObj, tileDiags := types.ObjectValue(tileAttrTypes, map[string]attr.Value{
			"x":      types.Float64Value(t.X),
			"y":      types.Float64Value(t.Y),
			"w":      types.Float64Value(t.W),
			"h":      types.Float64Value(t.H),
			"config": configList,
		})
		diags.Append(tileDiags...)
		tileValues[i] = tileObj
	}
	tilesList, tilesDiags := types.ListValue(types.ObjectType{AttrTypes: tileAttrTypes}, tileValues)
	diags.Append(tilesDiags...)
	state.Tiles = tilesList

	// Tags
	if len(d.Tags) > 0 {
		tagValues := make([]attr.Value, len(d.Tags))
		for i, tag := range d.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		tagsList, tagDiags := types.ListValue(types.StringType, tagValues)
		diags.Append(tagDiags...)
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
			obj, objDiags := types.ObjectValue(filterAttrTypes, map[string]attr.Value{
				"type":       types.StringValue(f.Type),
				"name":       types.StringValue(f.Name),
				"expression": types.StringValue(f.Expression),
				"source_id":  types.StringValue(f.SourceID),
			})
			diags.Append(objDiags...)
			filterValues[i] = obj
		}
		filtersList, flDiags := types.ListValue(types.ObjectType{AttrTypes: filterAttrTypes}, filterValues)
		diags.Append(flDiags...)
		state.Filters = filtersList
	} else {
		state.Filters = types.ListNull(types.ObjectType{AttrTypes: filterAttrTypes})
	}

	// Saved filter values
	if len(d.SavedFilterValues) > 0 {
		sfvValues := make([]attr.Value, len(d.SavedFilterValues))
		for i, sfv := range d.SavedFilterValues {
			obj, objDiags := types.ObjectValue(savedFilterValueAttrTypes, map[string]attr.Value{
				"type":      stringOrNull(sfv.Type),
				"condition": types.StringValue(sfv.Condition),
			})
			diags.Append(objDiags...)
			sfvValues[i] = obj
		}
		sfvList, sfvDiags := types.ListValue(types.ObjectType{AttrTypes: savedFilterValueAttrTypes}, sfvValues)
		diags.Append(sfvDiags...)
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
