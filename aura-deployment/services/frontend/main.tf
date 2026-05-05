module "docker_image" {
  source = "../../modules/docker-image"

  image_name      = var.image_name
  image_tag       = var.image_tag
  context_path    = var.context_path
  dockerfile_path = abspath("${path.module}/Dockerfile")

  build_args = {
    NODE_ENV = "production"
  }

  source_url    = "https://github.com/sushanth262/Aura"
  keep_remotely = true
}
