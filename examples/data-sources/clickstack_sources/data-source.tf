data "clickstack_sources" "all" {}

output "source_ids" {
  value = data.clickstack_sources.all.sources[*].id
}
