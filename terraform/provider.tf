provider "google" {
  credentials = file("gcp-key.json") # 記得把你的 JSON 金鑰放在這
  project     = "project-c0a608b0-a22f-495a-aa8"
  region      = "asia-east1"        
}

# 取得專案編號（用於之後的權限設定）
data "google_project" "project" {}

# 1. 建立 Workload Identity Pool
resource "google_iam_workload_identity_pool" "github_pool" {
  workload_identity_pool_id = "github-actions-pool"
  display_name              = "GitHub Actions Pool"
}

# 2. 建立 Workload Identity Provider (連結 GitHub)
resource "google_iam_workload_identity_pool_provider" "github_provider" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github_pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

# 3. 授權 GitHub 儲存庫使用服務帳戶
resource "google_service_account_iam_member" "wif_user" {
  service_account_id = "projects/${var.project_id}/serviceAccounts/terraform-manager@${var.project_id}.iam.gserviceaccount.com"
  role               = "roles/iam.workloadIdentityUser"
  member = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github_pool.name}/attribute.repository/jushengtin2/renter_backend"
}