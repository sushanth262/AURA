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

variable "frontend_expo_public_api_base_url" {
  description = "Optional. Passed to expo export as EXPO_PUBLIC_API_BASE_URL (must include /v1 suffix). Leave empty to keep Dockerfile default (localhost)."
  type        = string
  default     = ""
}

variable "frontend_expo_public_ws_base_url" {
  description = "Optional. Passed as EXPO_PUBLIC_WS_BASE_URL (wss:// for HTTPS BFF sites). Leave empty for Dockerfile default."
  type        = string
  default     = ""
}
