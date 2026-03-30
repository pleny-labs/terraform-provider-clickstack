resource "clickstack_dashboard" "example" {
  name = "Service Overview"
  tags = ["production", "sre"]

  tiles {
    x = 0
    y = 0
    w = 12
    h = 8

    config {
      name           = "Request Rate"
      display_type   = "line"
      source         = data.clickstack_sources.all.sources[0].id
      group_by       = "ServiceName"
      where_language = "lucene"
      granularity    = "5 minute"

      select {
        agg_fn           = "count"
        value_expression = ""
      }
    }
  }

  tiles {
    x = 12
    y = 0
    w = 12
    h = 8

    config {
      name           = "Error Count"
      display_type   = "number"
      source         = data.clickstack_sources.all.sources[0].id
      where_language = "lucene"

      select {
        agg_fn                 = "count"
        value_expression       = ""
        agg_condition          = "SeverityText:ERROR"
        agg_condition_language = "lucene"
      }
    }
  }
}
