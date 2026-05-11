variable "image_name" {
  description = "Full image name without tag (e.g. ghcr.io/org/aura-authz)"
  type        = string
}

variable "image_tag" {
  description = "Docker image tag"
  type        = string
  default     = "latest"
}

variable "context_path" {
  description = "Absolute path to repository root (Docker build context)"
  type        = string
}

variable "backend_rebuild_stamp" {
  description = "Bump from root module to force a new docker_image build."
  type        = string
  default     = "0"
}
