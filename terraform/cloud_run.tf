resource "google_cloud_run_v2_service" "backend" {
  name     = "hazukashii-backend"
  location = "asia-east1"

  template {
    containers {
      # 第一次執行先用 Google 的測試 image，之後 CI/CD 會覆蓋它
      image = "us-docker.pkg.dev/cloudrun/container/hello" 
      
      env {
        name = "DATABASE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.db_url.secret_id
            version = "latest"
          }
        }
      }
    }
  }
}