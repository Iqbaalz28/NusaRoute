#!/bin/bash
# deploy.sh
# Script untuk deploy manifest Kubernetes NusaRoute secara berurutan.

# Menghentikan script jika ada error
set -e

# Tentukan path absolute/relative ke folder manifest k8s
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Memulai Deployment NusaRoute ke Kubernetes ==="

# 1. Buat Namespace terlebih dahulu
echo "[1/3] Menerapkan Namespace..."
kubectl apply -f "$SCRIPT_DIR/namespace.yaml"

# 2. Hubungkan database & broker dari host ke Minikube
echo "[2/3] Menerapkan manifest External Services..."
kubectl apply -f "$SCRIPT_DIR/external-services.yaml"

# 3. Terapkan semua Services, Deployments, Secrets, dan HPA
echo "[3/3] Menerapkan manifest Services, Deployments, Secrets, dan HPA..."
kubectl apply -f "$SCRIPT_DIR/services.yaml"

echo "=== Deployment Selesai! ==="
echo "Gunakan 'kubectl get all -n nusaroute' untuk memantau status pod."
