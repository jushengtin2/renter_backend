resource "google_artifact_registry_repository" "my_repo" {
  location      = "asia-east1"
  repository_id = "renter-app-repo"
  format        = "DOCKER"
  description   = "Docker repository for hazukashii backend"
}