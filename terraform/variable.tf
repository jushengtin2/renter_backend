# #變數


variable "GCP_REGION" {
  type    = string
  default = "asia-east1"
}

variable "GCS_REGION" {
  type    = string
  default = "asia"
}

variable "GCP_PROJECT_ID" {
  description = "GCP Project ID"
  type        = string
  default     = "project-c0a608b0-a22f-495a-aa8"
}

variable "GCP_SERVICE_ACCOUNT_ID" {
  description = "GCP SA"
  type        = string
  default     = "projects/terraform-workflow-practice/serviceAccounts/terraform-manager@project-c0a608b0-a22f-495a-aa8.iam.gserviceaccount.com"

}

variable "GCS_BUCKET_NAME" {
  description = "Globally unique GCS bucket name"
  type        = string
  default     = "hazukashii-storage"
}
# variable "container_name" {
#   type    = string
#   default = "cloud-run-tf-ch4-6-3"
# }
