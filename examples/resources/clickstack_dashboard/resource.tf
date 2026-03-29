resource "clickstack_dashboard" "example" {
  name = "Service Overview"
  tags = ["production", "sre"]

  tiles {
    name = "Request Rate"
    x    = 0
    y    = 0
    w    = 6
    h    = 3

    config {
      display_type = "line"
      source_id    = data.clickstack_sources.all.sources[0].id

      select {
        agg_fn           = "count"
        value_expression = "*"
      }

      group_by = ["ServiceName"]
    }
  }

  tiles {
    name = "Error Count"
    x    = 6
    y    = 0
    w    = 6
    h    = 3

    config {
      display_type = "number"
      source_id    = data.clickstack_sources.all.sources[0].id

      select {
        agg_fn           = "count"
        value_expression = "*"
        where            = "SeverityText = 'ERROR'"
        where_language   = "sql"
        alias            = "Errors"
      }
    }
  }
}
