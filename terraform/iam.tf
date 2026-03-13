#!!!!其實這邊已經可以不用google_secret_manager_secret 因為這個是先去secrets.tf讀那個變數的id然後去gcp secrets對照
#啊我已經在gcp有那些id了 其實就直接填入id也可


# 授權 Cloud Run 讀取所有會用到的 Secret
resource "google_secret_manager_secret_iam_member" "secret_access" {
  for_each = {
    gcs_name                     = google_secret_manager_secret.gcs_name.id
    gcs_url                      = google_secret_manager_secret.gcs_url.id
    supabase_db_url              = google_secret_manager_secret.supabase_db_url.id
    supabase_storage_url         = google_secret_manager_secret.supabase_storage_url.id
    supabase_key_id              = google_secret_manager_secret.supabase_key_id.id
    supabase_access_key          = google_secret_manager_secret.supabase_access_key.id
    supabase_region              = google_secret_manager_secret.supabase_region.id
    clerk_jwks_url               = google_secret_manager_secret.clerk_jwks_url.id
    clerk_private_key            = google_secret_manager_secret.clerk_private_key.id
    clerk_webhook_signing_secret = google_secret_manager_secret.clerk_webhook_signing_secret.id
    clerk_webhook_path           = google_secret_manager_secret.clerk_webhook_path.id
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:terraform-manager@project-c0a608b0-a22f-495a-aa8.iam.gserviceaccount.com"


}
