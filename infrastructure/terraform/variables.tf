variable "kubeconfig_path" {
  description = "Path to the kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "istio_version" {
  description = "Version of Istio to install"
  type        = string
  default     = "1.20.0"
}

variable "argocd_version" {
  description = "Version of ArgoCD helm chart"
  type        = string
  default     = "5.51.6"
}

variable "prometheus_version" {
  description = "Version of kube-prometheus-stack helm chart"
  type        = string
  default     = "55.5.1"
}
