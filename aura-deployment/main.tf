provider "docker" {
  registry_auth {
    address  = "ghcr.io"
    username = var.ghcr_owner
    password = var.github_token
  }
}

# ── Registry metadata ─────────────────────────────────────────────────────────
module "registry" {
  source     = "./infrastructure/registry"
  ghcr_owner = var.ghcr_owner
}

# ── Services ──────────────────────────────────────────────────────────────────
module "frontend" {
  source = "./services/frontend"

  image_name   = "${module.registry.image_prefix}/aura-frontend"
  image_tag    = var.image_tag
  context_path = abspath(var.repo_root_path)
}
