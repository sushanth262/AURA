output "frontend_image" {
  description = "Fully-qualified frontend image reference pushed to GHCR"
  value       = module.frontend.image_name
}

output "frontend_digest" {
  description = "SHA-256 content digest of the pushed frontend image"
  value       = module.frontend.image_digest
}

output "aura_authz_image" {
  description = "Fully-qualified aura-authz image (GHCR)"
  value       = module.aura_authz.image_name
}

output "aura_authz_digest" {
  description = "Digest for aura-authz"
  value       = module.aura_authz.image_digest
}

output "aura_bff_api_image" {
  description = "Fully-qualified aura-bff-api image (GHCR)"
  value       = module.aura_bff_api.image_name
}

output "aura_bff_api_digest" {
  description = "Digest for aura-bff-api"
  value       = module.aura_bff_api.image_digest
}

output "aura_supervisor_image" {
  description = "Fully-qualified aura-supervisor image (GHCR)"
  value       = module.aura_supervisor.image_name
}

output "aura_supervisor_digest" {
  description = "Digest for aura-supervisor"
  value       = module.aura_supervisor.image_digest
}

output "aura_worker_image" {
  description = "Fully-qualified aura-worker image (GHCR)"
  value       = module.aura_worker.image_name
}

output "aura_worker_digest" {
  description = "Digest for aura-worker"
  value       = module.aura_worker.image_digest
}
