#這個檔案是為了存放 Docker 鏡像的地方

resource "google_artifact_registry_repository" "my_repo" {
  location      = var.GCP_REGION
  repository_id = "hazukashii-artifact-registry"
  format        = "DOCKER"
  description   = "Docker repository"
}