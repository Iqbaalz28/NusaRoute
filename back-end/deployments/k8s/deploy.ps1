# deploy.ps1
# Script PowerShell untuk deploy manifest Kubernetes NusaRoute secara berurutan.

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "=== Memulai Deployment NusaRoute ke Kubernetes ===" -ForegroundColor Cyan

# 1. Buat Namespace terlebih dahulu
Write-Host "[1/3] Menerapkan Namespace..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir\namespace.yaml"

# 2. Hubungkan database & broker dari host ke Minikube
Write-Host "[2/3] Menerapkan manifest External Services..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir\external-services.yaml"

# 3. Terapkan semua Services, Deployments, Secrets, dan HPA
Write-Host "[3/3] Menerapkan manifest Services, Deployments, Secrets, dan HPA..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir\services.yaml"

Write-Host "=== Deployment Selesai! ===" -ForegroundColor Green
Write-Host "Gunakan 'kubectl get all -n nusaroute' untuk memantau status pod." -ForegroundColor Green
