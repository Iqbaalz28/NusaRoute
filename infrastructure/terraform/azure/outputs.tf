output "resource_group" {
  value = azurerm_resource_group.rg.name
}

output "acr_login_server" {
  description = "Use this as the image registry prefix when building/pushing."
  value       = azurerm_container_registry.acr.login_server
}

output "acr_name" {
  value = azurerm_container_registry.acr.name
}

output "aks_name" {
  value = azurerm_kubernetes_cluster.aks.name
}

output "kubeconfig_path" {
  value = var.kubeconfig_out
}

output "get_credentials_cmd" {
  description = "Merge AKS creds into your default kubeconfig."
  value       = "az aks get-credentials --resource-group ${azurerm_resource_group.rg.name} --name ${azurerm_kubernetes_cluster.aks.name}"
}
