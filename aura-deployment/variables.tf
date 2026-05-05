variable "ghcr_owner" {
  description = "GitHub username or org that owns the GHCR namespace (e.g. sushanth262)"
  type        = string
  default     = "sushanth262"
}

variable "github_token" {
  description = "GitHub Personal Access Token with write:packages scope. Set via TF_VAR_github_token."
  type        = string
  sensitive   = true
}

variable "image_tag" {
  description = "Docker image tag applied to every service image. Use a git SHA in CI (e.g. git rev-parse --short HEAD)."
  type        = string
  default     = "latest"
}

variable "repo_root_path" {
  description = "Path to the repository root (used as Docker build context so both aura-frontend/ and aura-deployment/ are accessible). Defaults to the parent of this directory."
  type        = string
  default     = ".."
}

variable "frontend_rebuild_stamp" {
  description = "Change this string (e.g. bump a number or use an ISO date) to force rebuild and push of the frontend Docker image even when source hashes are unchanged."
  type        = string
  default     = "0"
}
