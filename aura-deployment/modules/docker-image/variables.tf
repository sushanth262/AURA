variable "image_name" {
  description = "Full image name without tag (e.g. ghcr.io/sushanth262/aura-frontend)"
  type        = string
}

variable "image_tag" {
  description = "Image tag. Use a git SHA in CI for immutable, traceable images."
  type        = string
  default     = "latest"
}

variable "context_path" {
  description = "Absolute path to the Docker build context directory"
  type        = string
}

variable "dockerfile_path" {
  description = "Absolute path to the Dockerfile"
  type        = string
}

variable "build_args" {
  description = "Build-time ARG values passed to docker build (--build-arg)"
  type        = map(string)
  default     = {}
}

variable "source_url" {
  description = "OCI org.opencontainers.image.source label value"
  type        = string
  default     = ""
}

variable "keep_remotely" {
  description = "Whether to keep the remote image when the Terraform resource is destroyed"
  type        = bool
  default     = true
}
