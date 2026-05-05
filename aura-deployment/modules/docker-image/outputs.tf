output "image_name" {
  description = "Fully-qualified image name including tag (e.g. ghcr.io/org/service:abc1234)"
  value       = docker_registry_image.this.name
}

output "image_digest" {
  description = "SHA-256 content digest of the remote image after push"
  value       = docker_registry_image.this.sha256_digest
}

output "image_id" {
  description = "Local Docker image ID (short SHA)"
  value       = docker_image.this.image_id
}
