variable "image_name" {
  description = "Full image name without tag (provided by root module via registry prefix)"
  type        = string
}

variable "image_tag" {
  description = "Docker image tag"
  type        = string
  default     = "latest"
}

variable "context_path" {
  description = "Absolute path to the aura-frontend source directory (build context)"
  type        = string
}

variable "frontend_rebuild_stamp" {
  description = "Bumped from root module to force a new docker_image build."
  type        = string
  default     = "0"
}
