locals {
  tagged_name = "${var.image_name}:${var.image_tag}"

  # Absolute paths work reliably with buildx; the legacy builder mishandles Dockerfile paths on Windows.
  _ctx_abs = replace(abspath(var.context_path), "\\", "/")
  _df_abs  = replace(abspath(var.dockerfile_path), "\\", "/")
}

# Build the image locally using the Docker daemon
resource "docker_image" "this" {
  name         = local.tagged_name
  keep_locally = false

  build {
    # Use Docker Buildx (same as `docker buildx build`) instead of the legacy builder, which often
    # breaks on Windows (wrong Dockerfile resolution, "unexpected EOF"). Requires Buildx (Docker Desktop includes it).
    builder    = "default"
    context    = local._ctx_abs
    dockerfile = local._df_abs

    build_args = var.build_args

    label = {
      "org.opencontainers.image.title"   = basename(var.image_name)
      "org.opencontainers.image.source"  = var.source_url
      "org.opencontainers.image.version" = var.image_tag
    }
  }

  # Rebuild whenever the tag changes (CI passes git SHA) or the Dockerfile changes
  triggers = {
    image_tag       = var.image_tag
    dockerfile_hash = filemd5(var.dockerfile_path)
  }
}

# Push the built image to the registry (GHCR)
resource "docker_registry_image" "this" {
  name          = docker_image.this.name
  keep_remotely = var.keep_remotely

  triggers = {
    image_id = docker_image.this.image_id
  }
}
