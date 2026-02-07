#!/bin/bash

# Kubernetes Deployment Script for Wedding Invitation API
# Usage: ./k8s-deploy.sh [version]

set -e

# Configuration
VERSION=${1:-latest}
NAMESPACE="wedding-invitation"
K8S_DIR="k8s"

echo "🚀 Deploying Wedding Invitation API version: $VERSION"
echo "📁 Kubernetes manifests directory: $K8S_DIR"
echo "🏷️  Target namespace: $NAMESPACE"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if k8s directory exists
if [ ! -d "$K8S_DIR" ]; then
    echo "❌ Kubernetes manifests directory not found: $K8S_DIR"
    exit 1
fi

# Create namespace if it doesn't exist
echo "📦 Creating namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Apply manifests in order
echo "🔧 Applying Kubernetes manifests..."

echo "  📋 Applying secrets..."
kubectl apply -f $K8S_DIR/secrets.yaml

echo "  ⚙️  Applying configmap..."
kubectl apply -f $K8S_DIR/configmap.yaml

echo "  🚀 Applying deployment..."
# Update image tag if version is provided
if [ "$VERSION" != "latest" ]; then
    sed -i.bak "s|image: wedding-api:latest|image: wedding-api:$VERSION|" $K8S_DIR/deployment.yaml
    kubectl apply -f $K8S_DIR/deployment.yaml
    mv $K8S_DIR/deployment.yaml.bak $K8S_DIR/deployment.yaml
else
    kubectl apply -f $K8S_DIR/deployment.yaml
fi

echo "  🌐 Applying service..."
kubectl apply -f $K8S_DIR/service.yaml

echo "  🛡️  Applying ingress..."
kubectl apply -f $K8S_DIR/ingress.yaml

echo "  📈 Applying HPA..."
kubectl apply -f $K8S_DIR/hpa.yaml

# Wait for deployment to be ready
echo "⏳ Waiting for deployment to be ready..."
kubectl rollout status deployment/wedding-api -n $NAMESPACE --timeout=300s

# Show deployment status
echo ""
echo "✅ Deployment completed successfully!"
echo ""
echo "📊 Current deployment status:"
kubectl get pods -n $NAMESPACE -l app=wedding-api

echo ""
echo "🌐 Service status:"
kubectl get service -n $NAMESPACE

echo ""
echo "🛡️  Ingress status:"
kubectl get ingress -n $NAMESPACE

echo ""
echo "📈 HPA status:"
kubectl get hpa -n $NAMESPACE

echo ""
echo "🔍 To check logs:"
echo "  kubectl logs -n $NAMESPACE -l app=wedding-api -f"

echo ""
echo "🔄 To restart deployment:"
echo "  kubectl rollout restart deployment/wedding-api -n $NAMESPACE"

echo ""
echo "🗑️  To cleanup:"
echo "  kubectl delete namespace $NAMESPACE"

echo ""
echo "🎉 Wedding Invitation API is now deployed!"