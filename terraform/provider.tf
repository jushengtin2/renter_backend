terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.GCP_PROJECT_ID
  region  = var.GCP_REGION
}

resource "google_storage_bucket" "hazukashii_bucket" {
  name                        = var.GCS_BUCKET_NAME
  location                    = var.GCS_REGION
  uniform_bucket_level_access = true
   lifecycle {
    prevent_destroy = true
  }
}
