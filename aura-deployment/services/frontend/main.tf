locals {
  # Aura repo root (parent of aura-deployment): …/Aura
  repo_root = abspath("${path.module}/../../..")

  # Track app sources used by the Docker build context so asset/icon edits invalidate the image.
  aura_frontend_dir_globs = [
    "aura-frontend/app/**/*",
    "aura-frontend/src/**/*",
    "aura-frontend/icons/**/*",
    "aura-frontend/specs/**/*",
  ]
  aura_frontend_root_files = [
    "aura-frontend/package.json",
    "aura-frontend/tsconfig.json",
    "aura-frontend/app.json",
    "aura-frontend/babel.config.js",
  ]

  aura_frontend_rel_files = sort(distinct(concat(
    flatten([for g in local.aura_frontend_dir_globs : fileset(local.repo_root, g)]),
    [for f in local.aura_frontend_root_files : f if fileexists("${local.repo_root}/${f}")],
  )))

  aura_frontend_source_hash = length(local.aura_frontend_rel_files) > 0 ? substr(
    sha256(join("\n", [for f in local.aura_frontend_rel_files : "${f}:${filesha256("${local.repo_root}/${f}")}"])),
    0,
    40,
  ) : "empty"
}

module "docker_image" {
  source = "../../modules/docker-image"

  image_name      = var.image_name
  image_tag       = var.image_tag
  context_path    = var.context_path
  dockerfile_path = abspath("${path.module}/Dockerfile")

  build_args = {
    NODE_ENV = "production"
  }

  rebuild_triggers = {
    aura_frontend_source_hash = local.aura_frontend_source_hash
  }

  source_url    = "https://github.com/sushanth262/Aura"
  keep_remotely = true
}
