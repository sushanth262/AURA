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

  image_name                         = "${module.registry.image_prefix}/aura-frontend"
  image_tag                          = var.image_tag
  context_path                       = abspath(var.repo_root_path)
  frontend_rebuild_stamp             = var.frontend_rebuild_stamp
  frontend_expo_public_api_base_url  = var.frontend_expo_public_api_base_url
  frontend_expo_public_ws_base_url   = var.frontend_expo_public_ws_base_url
}

module "aura_authz" {
  source = "./services/authz"

  image_name             = "${module.registry.image_prefix}/aura-authz"
  image_tag              = var.image_tag
  context_path           = abspath(var.repo_root_path)
  backend_rebuild_stamp  = var.backend_rebuild_stamp
}

module "aura_bff_api" {
  source = "./services/bff"

  image_name             = "${module.registry.image_prefix}/aura-bff-api"
  image_tag              = var.image_tag
  context_path           = abspath(var.repo_root_path)
  backend_rebuild_stamp  = var.backend_rebuild_stamp
}

module "aura_supervisor" {
  source = "./services/supervisor"

  image_name             = "${module.registry.image_prefix}/aura-supervisor"
  image_tag              = var.image_tag
  context_path           = abspath(var.repo_root_path)
  backend_rebuild_stamp  = var.backend_rebuild_stamp
}

module "aura_worker" {
  source = "./services/worker"

  image_name             = "${module.registry.image_prefix}/aura-worker"
  image_tag              = var.image_tag
  context_path           = abspath(var.repo_root_path)
  backend_rebuild_stamp  = var.backend_rebuild_stamp
}
