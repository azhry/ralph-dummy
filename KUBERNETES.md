# Kubernetes Deployment Guide

## Prerequisites

1. **Kubernetes Cluster** (v1.20+)
2. **kubectl** configured to connect to your cluster
3. **Ingress Controller** (nginx-ingress recommended)
4. **cert-manager** for SSL certificate management
5. **StorageClass** configured for persistent volumes

## Quick Deployment

### 1. Prepare Your Secrets

Edit `secret.yaml` and replace the base64-encoded values with your actual production secrets:

```bash
# Encode your MongoDB URI
echo -n "mongodb://username:password@mongodb-service:27017/wedding_invitations" | base64

# Generate JWT secrets
openssl rand -base64 32 | base64

# Encode other secrets
echo -n "your-sendgrid-api-key" | base64
echo -n "your-aws-access-key" | base64
echo -n "your-aws-secret-key" | base64
```

### 2. Update Configuration

Edit `configmap.yaml` and update:
- `allowed-origins`: Your frontend domain(s)
- `cdn-url`: Your CDN domain
- `email-from`: Your email address
- Storage and other settings

### 3. Deploy the Application

```bash
# Apply all configurations
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f service.yaml
kubectl apply -f deployment.yaml
kubectl apply -f ingress.yaml

# Or deploy all at once
kubectl apply -f .
```

### 4. Verify Deployment

```bash
# Check deployment status
kubectl get deployment wedding-invitation-backend
kubectl get pods -l app=wedding-invitation-backend

# Check services
kubectl get service wedding-invitation-backend-service

# Check ingress
kubectl get ingress wedding-invitation-backend-ingress

# Check logs
kubectl logs -l app=wedding-invitation-backend -f
```

## Configuration Options

### Environment Variables

The application uses the following configuration hierarchy:
1. **Secrets** (sensitive data): Database URLs, API keys, JWT secrets
2. **ConfigMap** (configuration): URLs, timeouts, feature flags
3. **Defaults** (built-in): Fallback values

### Storage

- **Persistent Volume**: 10Gi for file uploads (configurable)
- **Storage Class**: Uses `fast-ssd` by default (adjust for your cluster)
- **Upload Directory**: Mounted at `/app/uploads`

### Networking

- **Internal Service**: ClusterIP on port 80
- **External Access**: Via Ingress on port 443 (HTTPS)
- **Load Balancing**: 3 replicas by default
- **Health Checks**: Liveness and readiness probes enabled

## SSL/TLS Setup

### Using cert-manager (Recommended)

1. Install cert-manager:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

2. Create ClusterIssuer (example for Let's Encrypt):
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@yourdomain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### Manual TLS Certificate

If not using cert-manager, create your TLS secret:
```bash
kubectl create secret tls wedding-api-tls \
  --key=tls.key \
  --cert=tls.crt \
  --namespace=default
```

## Scaling Configuration

### Horizontal Scaling

Edit `deployment.yaml` to adjust replicas:
```yaml
spec:
  replicas: 5  # Increase for higher load
```

### Resource Limits

Adjust resource requests/limits based on your needs:
```yaml
resources:
  requests:
    memory: "512Mi"    # Increase for memory-intensive workloads
    cpu: "500m"        # Increase for CPU-intensive workloads
  limits:
    memory: "1Gi"      # Maximum memory per pod
    cpu: "1000m"       # Maximum CPU per pod
```

## Monitoring and Logging

### Application Logs

```bash
# View real-time logs
kubectl logs -l app=wedding-invitation-backend -f

# View logs for specific pod
kubectl logs <pod-name> --previous  # For crashed pods
```

### Health Monitoring

The application exposes health endpoints:
- `/health` - Basic health check
- `/health/detailed` - Detailed system status

Access via: `https://api.yourdomain.com/health`

## Troubleshooting

### Common Issues

1. **Pods not starting**
   ```bash
   kubectl describe pod <pod-name>
   kubectl logs <pod-name>
   ```

2. **Secrets not working**
   ```bash
   kubectl get secret wedding-secrets -o yaml
   ```

3. **Ingress not working**
   ```bash
   kubectl get ingress wedding-invitation-backend-ingress -o yaml
   kubectl describe ingress wedding-invitation-backend-ingress
   ```

4. **Persistent volume issues**
   ```bash
   kubectl get pvc wedding-uploads-pvc
   kubectl describe pvc wedding-uploads-pvc
   ```

### Database Connection Issues

Ensure your MongoDB service is accessible:
```bash
# Test connectivity from pod
kubectl exec -it <pod-name> -- nc -zv mongodb-service 27017
```

### Performance Tuning

1. **Increase replicas** for higher throughput
2. **Adjust resource limits** based on usage patterns
3. **Enable HPA** (Horizontal Pod Autoscaler):
   ```bash
   kubectl autoscale deployment wedding-invitation-backend \
     --cpu-percent=70 \
     --min=3 \
     --max=10
   ```

## Backup and Recovery

### Data Backup

Regular backups of:
- MongoDB database
- Uploaded files (PVC)
- Kubernetes secrets and configmaps

### Disaster Recovery

1. Restore MongoDB from backup
2. Re-deploy application manifests
3. Restore uploaded files to PVC

## Security Considerations

1. **Network Policies**: Restrict inter-pod communication
2. **Pod Security**: Run as non-root (already configured)
3. **Secrets Management**: Use external secret store if possible
4. **Image Security**: Use signed images and vulnerability scanning
5. **RBAC**: Configure proper role-based access control

## Production Checklist

- [ ] All secrets properly encoded and configured
- [ ] SSL/TLS certificates configured
- [ ] Resource limits set appropriately
- [ ] Health checks passing
- [ ] Monitoring and alerting configured
- [ ] Backup strategy implemented
- [ ] Security policies applied
- [ ] Load testing performed
- [ ] DNS records configured to point to ingress
- [ ] CORS settings updated for production domains