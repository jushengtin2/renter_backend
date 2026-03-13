# #這個檔案定義 Cloud Run Service。
# #關鍵內容：

# #指定使用哪個 Container Image（由 artifact.tf 提供路徑）。
# #設定環境變數、記憶體（RAM）、CPU 以及自動縮放（Autoscaling）規則。
# #設定服務是否開放給大眾存取（IAM 權限）。

resource "google_cloud_run_v2_service" "hazukashii_backend" {
  name       = "hazukashii-backend"
  location   = var.GCP_REGION
  depends_on = [google_secret_manager_secret_iam_member.secret_access] #因為iam要先載入（有很多參數）才能跑cloudrun

  template {
    containers {
      #先用 Google 的測試 image
      image = "us-docker.pkg.dev/cloudrun/container/hello"

      env {
        name = "GCS_BUCKET_NAME"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.gcs_name.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "GCS_DB_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.gcs_url.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "SUPABASE_DB_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.supabase_db_url.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "SUPABASE_STORAGE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.supabase_storage_url.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "SUPABASE_KEY_ID"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.supabase_key_id.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "SUPABASE_ACCESS_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.supabase_access_key.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "SUPABASE_REGION"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.supabase_region.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "CLERK_JWKS_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.clerk_jwks_url.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "CLERK_PRIVATE_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.clerk_private_key.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "CLERK_WEBHOOK_SIGNING_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.clerk_webhook_signing_secret.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "CLERK_WEBHOOK_PATH"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.clerk_webhook_path.secret_id
            version = "latest"
          }
        }
      }
    }
  }
}
