output "registry_url" {
  description = "GHCR registry hostname"
  value       = local.registry_url
}

output "image_prefix" {
  description = "Base image path for all Aura images (e.g. ghcr.io/sushanth262)"
  value       = local.image_prefix
}
