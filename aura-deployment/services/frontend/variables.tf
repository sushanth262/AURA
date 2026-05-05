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
