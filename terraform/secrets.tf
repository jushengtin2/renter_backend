# 1. 建立 Secret 容器
resource "google_secret_manager_secret" "db_url" {
  secret_id = "supabase-db-url"
  replication {
    automatic = true
  }
}

# 2. 授權 Cloud Run 讀取這個 Secret (IAM 權限)
resource "google_secret_manager_secret_iam_member" "secret_access" {
  secret_id = google_secret_manager_secret.db_url.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}