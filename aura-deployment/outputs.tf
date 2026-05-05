output "frontend_image" {
  description = "Fully-qualified frontend image reference pushed to GHCR"
  value       = module.frontend.image_name
}

output "frontend_digest" {
  description = "SHA-256 content digest of the pushed frontend image"
  value       = module.frontend.image_digest
}
