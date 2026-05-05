# GHCR is fully managed by GitHub — no Terraform resource creation is required.
# This module provides canonical registry coordinates used by every service module.

locals {
  registry_url = "ghcr.io"
  image_prefix = "${local.registry_url}/${var.ghcr_owner}"
}
