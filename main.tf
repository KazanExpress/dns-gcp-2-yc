terraform {
  required_version = ">= 1.0.1"

  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = ">= 0.72.0"
    }
  }
}

provider "yandex" {
  token     = var.yc_token
  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
  zone      = var.yc_zone
}

variable "yc_token" {
  description = "The Yandex.Cloud API key."
  type        = string
  sensitive   = true
}

variable "yc_cloud_id" {
  description = "The Yandex.Cloud cloud id."
  type        = string
}

variable "yc_folder_id" {
  description = "The Yandex.Cloud folder id."
  type        = string
}

variable "yc_zone" {
  description = "The Yandex.Cloud availability zone."
  type        = string
  default     = "ru-central1-a"
}
