# Terraform Provider for ClickStack (HyperDX)

Terraform provider for managing [ClickStack](https://clickhouse.com/docs/clickstack) (HyperDX) observability resources.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Installation

```hcl
terraform {
  required_providers {
    clickstack = {
      source = "pleny-labs/clickstack"
    }
  }
}
```

## Configuration

```hcl
provider "clickstack" {
  endpoint = "https://clickstack.example.com"  # or CLICKSTACK_ENDPOINT env var
  api_key  = var.clickstack_api_key             # or CLICKSTACK_API_KEY env var
}
```

## Resources

### clickstack_dashboard

Manages dashboards with tiles, filters, and saved queries.

```hcl
resource "clickstack_dashboard" "example" {
  name = "Service Overview"
  tags = ["production"]

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
    }
  }
}
```

### clickstack_alert

Manages alerts with threshold-based triggers and webhook notifications.

```hcl
resource "clickstack_alert" "high_error_rate" {
  name           = "High Error Rate"
  threshold      = 100
  interval       = "5m"
  threshold_type = "above"
  source         = "tile"
  dashboard_id   = clickstack_dashboard.example.id

  channel {
    type       = "webhook"
    webhook_id = clickstack_webhook.slack.id
  }
}
```

### clickstack_webhook

Manages webhooks for alert notifications.

```hcl
resource "clickstack_webhook" "slack" {
  name        = "Slack Alerts"
  service     = "slack"
  url         = var.slack_webhook_url
  description = "Production alerts channel"
}
```

## Data Sources

### clickstack_sources

Lists all data sources.

```hcl
data "clickstack_sources" "all" {}
```

### clickstack_webhooks

Lists all webhooks.

```hcl
data "clickstack_webhooks" "all" {}
```

## Development

```bash
# Build
make build

# Run tests
make test

# Run acceptance tests (requires API credentials)
make testacc

# Generate documentation
make docs

# Install locally
make install
```

## Import

All resources support `terraform import`:

```bash
terraform import clickstack_dashboard.example <dashboard-id>
terraform import clickstack_alert.example <alert-id>
terraform import clickstack_webhook.example <webhook-id>
```
