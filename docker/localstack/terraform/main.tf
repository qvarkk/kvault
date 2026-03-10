variable "access_key" {
  type = string
}

variable "secret_key" {
  type = string
}

variable "default_region" {
  type = string
}

variable "s3_endpoint" {
  type = string
}

variable "bucket_name" {
  type = string
}

provider "aws" {

  access_key                  = var.access_key
  secret_key                  = var.secret_key
  region                      = var.default_region

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true

  endpoints {
    s3             = var.s3_endpoint
  }
}

resource "aws_s3_bucket" "kvault" {
  bucket = var.bucket_name
}