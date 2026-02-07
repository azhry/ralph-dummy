# Kubernetes Deployment Manifests

This directory contains production-ready Kubernetes manifests for deploying the Wedding Invitation Backend API.

## Files

| File | Description |
|------|-------------|
| `namespace.yaml` | Kubernetes namespace for the application |
| `secrets.yaml` | Sensitive configuration (API keys, database URLs) |
| `configmap.yaml` | Non-sensitive application configuration |
| `deployment.yaml` | Main application deployment with pod specs |
| `service.yaml` | Internal service for load balancing |
| `ingress.yaml` | External access via Ingress controller |
| `hpa.yaml` | Horizontal Pod Autoscaler for scaling |
| `k8s-deploy.sh` | Automated deployment script |

## Quick Start

### Prerequisites

- Kubernetes cluster (minikube, EKS, GKE, etc.)
- kubectl configured and connected to your cluster
- Ingress controller installed (nginx-ingress recommended)
- cert-manager installed for SSL certificates (optional)

### Deployment

1. **Update Secrets**: Edit `secrets.yaml` and replace base64 values with your actual secrets:

```bash
# Example: Update MongoDB URI
echo -n "mongodb://user:pass@host:27017/db" | base64
# Replace the mongodb-uri value in secrets.yaml
```

2. **Deploy Application**:

```bash
# Deploy with latest tag
./k8s-deploy.sh

# Deploy with specific version
./k8s-deploy.sh v1.2.0
```

3. **Verify Deployment**:

```bash
# Check pods
kubectl get pods -n wedding-invitation

# Check services
kubectl get services -n wedding-invitation

# Check ingress
kubectl get ingress -n wedding-invitation

# Check application logs
kubectl logs -n wedding-invitation -l app=wedding-api -f
```

## Configuration

### Environment Variables

The application uses both ConfigMaps and Secrets for configuration:

**ConfigMap (configmap.yaml)**:
- `mongodb-database`: MongoDB database name
- `aws-region`: AWS region for S3
- `s3-bucket`: S3 bucket name
- `email-provider`: Email service provider
- `rate-limit-requests`: Rate limit configuration
- And more...

**Secrets (secrets.yaml)**:
- `mongodb-uri`: Full MongoDB connection string
- `redis-url`: Redis connection string
- `jwt-secret`: JWT signing secret
- `jwt-refresh-secret`: JWT refresh token secret
- `sendgrid-api-key`: SendGrid API key
- And more...

### Customization

1. **Resource Limits**: Adjust memory and CPU requests/limits in `deployment.yaml`
2. **Replica Count**: Modify the `replicas` field in `deployment.yaml`
3. **Auto Scaling**: Update min/max replicas in `hpa.yaml`
4. **Ingress**: Update hostnames in `ingress.yaml` for your domain
5. **TLS**: Update TLS secret name and hosts in `ingress.yaml`

## Monitoring

### Health Checks

The deployment includes both liveness and readiness probes:

- **Liveness Probe**: `/health` endpoint every 10 seconds (after 30s delay)
- **Readiness Probe**: `/health` endpoint every 5 seconds (after 5s delay)

### Horizontal Pod Autoscaling

The HPA is configured to:
- Scale between 2-10 replicas
- Target 70% CPU utilization
- Target 80% memory utilization
- Scale up quickly, scale down gradually

## Security

### Pod Security

- Non-root user (UID 1000)
- Read-only root filesystem
- Drop all Linux capabilities
- Security context for filesystem group

### Network Security

- Network policies should be added for production
- Ingress configured with SSL redirect
- CORS headers configured via Ingress annotations
- Rate limiting via Ingress annotations

## Troubleshooting

### Common Issues

1. **Pods not starting**:
   ```bash
   kubectl describe pod -n wedding-invitation
   kubectl logs -n wedding-invitation <pod-name>
   ```

2. **Service not accessible**:
   ```bash
   kubectl get endpoints -n wedding-invitation
   kubectl describe service wedding-api-service -n wedding-invitation
   ```

3. **Ingress not working**:
   ```bash
   kubectl describe ingress wedding-api-ingress -n wedding-invitation
   kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx
   ```

### Cleanup

To remove the entire deployment:

```bash
kubectl delete namespace wedding-invitation
```

## Production Considerations

1. **Backup Strategy**: Implement MongoDB backup strategy
2. **Monitoring**: Add Prometheus/Grafana monitoring
3. **Logging**: Configure centralized logging (ELK stack)
4. **Network Policies**: Implement network segmentation
5. **Pod Disruption Budgets**: Add PDBs for high availability
6. **Resource Quotas**: Set namespace resource limits
7. **Image Security**: Use image scanning and trusted registry

## Integration with CI/CD

These manifests are designed to work with CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Deploy to Kubernetes
  run: |
    echo "${{ secrets.KUBECONFIG }}" | base64 -d > kubeconfig
    export KUBECONFIG=kubeconfig
    ./k8s/k8s-deploy.sh ${{ github.sha }}
```