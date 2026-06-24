variable "subscription_id" {
  description = "Azure subscription ID. Leave empty to use the az CLI default."
  type        = string
  default     = ""
}

variable "prefix" {
  description = "Name prefix for all resources (must be globally-unique-friendly for ACR)."
  type        = string
  default     = "nusaroute"
}

variable "location" {
  description = "Azure region."
  type        = string
  default     = "Southeast Asia"
}

variable "kubernetes_version" {
  description = "AKS Kubernetes version. Empty = AKS default for the region."
  type        = string
  default     = ""
}

variable "node_count" {
  description = "Number of nodes in the default pool."
  type        = number
  default     = 3
}

variable "node_vm_size" {
  description = "VM size per node. Standard_B4ms = 4 vCPU / 16GB, burstable & cheap; good for a self-hosted-everything showcase."
  type        = string
  default     = "Standard_B4ms"
}

variable "acr_sku" {
  description = "Azure Container Registry SKU."
  type        = string
  default     = "Basic"
}

variable "kubeconfig_out" {
  description = "Path to write the AKS kubeconfig to (consumed by the addons layer)."
  type        = string
  default     = "../aks.kubeconfig"
}
