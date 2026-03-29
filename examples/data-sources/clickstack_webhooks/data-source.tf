data "clickstack_webhooks" "all" {}

output "webhook_names" {
  value = data.clickstack_webhooks.all.webhooks[*].name
}
