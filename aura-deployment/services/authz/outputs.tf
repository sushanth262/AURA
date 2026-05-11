output "image_name" {
  description = "Fully-qualified aura-authz image reference (name:tag)"
  value       = module.docker_image.image_name
}

output "image_digest" {
  description = "SHA-256 digest of the pushed aura-authz image"
  value       = module.docker_image.image_digest
}
