# Azure infrastructure layer for NusaRoute: Resource Group + ACR + AKS.
# Run this BEFORE the platform-addons layer (../). It provisions the cluster;
# the addons layer (Istio/ArgoCD/Prometheus) then targets the kubeconfig this
# layer writes out.
terraform {
  required_version = ">= 1.3"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.110"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "azurerm" {
  features {}
  # Auth comes from `az login` (Azure CLI). Optionally set subscription_id via
  # the variable or ARM_SUBSCRIPTION_ID env var.
  subscription_id = var.subscription_id != "" ? var.subscription_id : null
}
