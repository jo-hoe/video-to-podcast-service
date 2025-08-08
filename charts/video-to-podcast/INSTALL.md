# Installation Guide for Video to Podcast Helm Chart

This guide provides step-by-step instructions for installing the Video to Podcast service using Helm.

## Prerequisites

1. **Kubernetes Cluster**: Ensure you have access to a Kubernetes cluster (v1.19+)
2. **Helm**: Install Helm 3.2.0 or later
3. **kubectl**: Configured to access your Kubernetes cluster
4. **Container Images**: Build and push the container images to a registry

## Building Container Images

Before installing the chart, you need to build and push the container images:

```bash
# Build API image
docker build -f Dockerfile.api -t your-registry/video-to-podcast-api:v1.0.0 .
docker push your-registry/video-to-podcast-api:v1.0.0

# Build UI image
docker build -f Dockerfile.ui -t your-registry/video-to-podcast-ui:v1.0.0 .
docker push your-registry/video-to-podcast-ui:v1.0.0
```

## Installation Steps

### 1. Basic Installation

For a quick test installation with default settings:

```bash
helm install video-to-podcast ./charts/video-to-podcast \
  --set api.image.repository=your-registry/video-to-podcast-api \
  --set api.image.tag=v1.0.0 \
  --set ui.image.repository=your-registry/video-to-podcast-ui \
  --set ui.image.tag=v1.0.0
```

### 2. Development Installation

For development with NodePort services and relaxed security:

```bash
helm install video-to-podcast ./charts/video-to-podcast \
  -f charts/video-to-podcast/values-dev.yaml \
  --set api.image.repository=your-registry/video-to-podcast-api \
  --set api.image.tag=v1.0.0 \
  --set ui.image.repository=your-registry/video-to-podcast-ui \
  --set ui.image.tag=v1.0.0
```

Access the application:
- UI: `http://<node-ip>:30300`
- API: `http://<node-ip>:30080`

### 3. Production Installation

For production deployment with persistence, ingress, and autoscaling:

```bash
# Create namespace
kubectl create namespace video-to-podcast

# Install with production values
helm install video-to-podcast ./charts/video-to-podcast \
  --namespace video-to-podcast \
  -f charts/video-to-podcast/values-prod.yaml \
  --set api.image.repository=your-registry/video-to-podcast-api \
  --set api.image.tag=v1.0.0 \
  --set ui.image.repository=your-registry/video-to-podcast-ui \
  --set ui.image.tag=v1.0.0 \
  --set ingress.hosts[0].host=your-domain.com
```

### 4. Installation with Persistence

To enable data persistence:

```bash
helm install video-to-podcast ./charts/video-to-podcast \
  --set persistence.enabled=true \
  --set persistence.resources.enabled=true \
  --set persistence.resources.size=50Gi \
  --set persistence.database.enabled=true \
  --set persistence.database.size=5Gi \
  --set api.config.database.connectionString="/app/data/database/podcast.db" \
  --set api.image.repository=your-registry/video-to-podcast-api \
  --set api.image.tag=v1.0.0 \
  --set ui.image.repository=your-registry/video-to-podcast-ui \
  --set ui.image.tag=v1.0.0
```

### 5. Installation with YouTube Cookies

If you need to access private or age-restricted videos:

```bash
# Create secret with cookies
kubectl create secret generic youtube-cookies \
  --from-file=youtube_cookies.txt=/path/to/your/cookies.txt

# Install with cookies enabled
helm install video-to-podcast ./charts/video-to-podcast \
  --set youtubeCookies.enabled=true \
  --set api.config.external.ytdlpCookiesFile="/app/cookies/youtube_cookies.txt" \
  --set api.image.repository=your-registry/video-to-podcast-api \
  --set api.image.tag=v1.0.0 \
  --set ui.image.repository=your-registry/video-to-podcast-ui \
  --set ui.image.tag=v1.0.0
```

## Verification

After installation, verify the deployment:

```bash
# Check pods
kubectl get pods -l app.kubernetes.io/instance=video-to-podcast

# Check services
kubectl get svc -l app.kubernetes.io/instance=video-to-podcast

# Check ingress (if enabled)
kubectl get ingress -l app.kubernetes.io/instance=video-to-podcast

# Check logs
kubectl logs -l app.kubernetes.io/instance=video-to-podcast,app.kubernetes.io/component=api
kubectl logs -l app.kubernetes.io/instance=video-to-podcast,app.kubernetes.io/component=ui
```

## Accessing the Application

### Via Port Forward (for testing)

```bash
kubectl port-forward svc/video-to-podcast-ui 3000:3000
kubectl port-forward svc/video-to-podcast-api 8080:8080
```

Then access:
- UI: http://localhost:3000
- API: http://localhost:8080

### Via Ingress

If ingress is enabled, access via the configured hostname.

### Via NodePort (development)

If using NodePort services, access via:
- UI: `http://<node-ip>:30300`
- API: `http://<node-ip>:30080`

## Upgrading

To upgrade the deployment:

```bash
helm upgrade video-to-podcast ./charts/video-to-podcast \
  --set api.image.tag=v1.1.0 \
  --set ui.image.tag=v1.1.0
```

## Uninstalling

To uninstall the deployment:

```bash
helm uninstall video-to-podcast
```

To also remove persistent volumes:

```bash
kubectl delete pvc -l app.kubernetes.io/instance=video-to-podcast
```

## Troubleshooting

### Common Issues

1. **Image Pull Errors**: Ensure images are built and pushed to the correct registry
2. **Permission Errors**: Check security contexts and service account permissions
3. **Storage Issues**: Verify storage classes and PVC creation
4. **Network Issues**: Check ingress configuration and DNS resolution

### Debug Commands

```bash
# Describe pods for detailed information
kubectl describe pods -l app.kubernetes.io/instance=video-to-podcast

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp

# Check resource usage
kubectl top pods -l app.kubernetes.io/instance=video-to-podcast
```

## Configuration Examples

### Custom Values File

Create a `my-values.yaml` file:

```yaml
api:
  image:
    repository: your-registry/video-to-podcast-api
    tag: v1.0.0
  resources:
    requests:
      memory: 1Gi
      cpu: 500m
    limits:
      memory: 2Gi
      cpu: 1000m

ui:
  image:
    repository: your-registry/video-to-podcast-ui
    tag: v1.0.0

ingress:
  enabled: true
  hosts:
    - host: video-to-podcast.example.com
      paths:
        - path: /
          pathType: Prefix
          service: ui
        - path: /v1
          pathType: Prefix
          service: api

persistence:
  enabled: true
  resources:
    enabled: true
    size: 100Gi
  database:
    enabled: true
    size: 10Gi
```

Then install:

```bash
helm install video-to-podcast ./charts/video-to-podcast -f my-values.yaml
```

## Support

For issues and questions:
1. Check the application logs
2. Verify Kubernetes resources
3. Review the chart documentation
4. Check the project repository for known issues
