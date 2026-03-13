# 只管理 Secret 容器本身，不管理 secret version 內容。
resource "google_secret_manager_secret" "gcs_name" {
  secret_id = "hazukashii_picture_storage"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "gcs_url" {
  secret_id = "gcs-url"
  replication {
    auto {}
  }
}



resource "google_secret_manager_secret" "supabase_db_url" {
  secret_id = "supabase-db-url"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "supabase_storage_url" {
  secret_id = "supabase-storage-url"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "supabase_key_id" {
  secret_id = "supabase-key-id"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "supabase_access_key" {
  secret_id = "supabase-access-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "supabase_region" {
  secret_id = "supabase-region"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "clerk_jwks_url" {
  secret_id = "clerk-jwks-url"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "clerk_private_key" {
  secret_id = "clerk-private-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "clerk_webhook_signing_secret" {
  secret_id = "clerk-webhook-signing-secret"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "clerk_webhook_path" {
  secret_id = "clerk-webhook-path"
  replication {
    auto {}
  }
}
