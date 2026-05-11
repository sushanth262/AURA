locals {
  repo_root = abspath("${path.module}/../../..")

  aura_backend_dir_globs = [
    "aura-backend/**/*.go",
    "aura-backend/internal/fixturesdata/*.yaml",
  ]
  aura_backend_root_files = [
    "aura-backend/go.mod",
  ]

  aura_backend_rel_files = sort(distinct(concat(
    flatten([for g in local.aura_backend_dir_globs : fileset(local.repo_root, g)]),
    [for f in local.aura_backend_root_files : f if fileexists("${local.repo_root}/${f}")],
  )))

  aura_backend_source_hash = length(local.aura_backend_rel_files) > 0 ? substr(
    sha256(join("\n", [for f in local.aura_backend_rel_files : "${f}:${filesha256("${local.repo_root}/${f}")}"])),
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
    CACHEBUST = var.backend_rebuild_stamp
  }

  rebuild_triggers = {
    aura_backend_source_hash = local.aura_backend_source_hash
    backend_rebuild_stamp    = var.backend_rebuild_stamp
  }

  source_url    = "https://github.com/sushanth262/Aura"
  keep_remotely = true
}
