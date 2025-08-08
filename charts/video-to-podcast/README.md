# Video to Podcast Helm Chart

This Helm chart deploys the Video to Podcast Service on Kubernetes, which converts YouTube videos to podcast feeds that can be subscribed to in any podcast app.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- PV provisioner support in the underlying infrastructure (if persistence is enabled)

## Installing the Chart

To install the chart with the release name `my-video-to-podcast`:

```bash
helm install my-video-to-podcast ./charts/video-to-podcast
```

The command deploys Video to Podcast Service on the Kubernetes cluster with default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-video-to-podcast` deployment:

```bash
helm delete my-video-to-podcast
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Video to Podcast chart and their default values.

### Global Parameters

| Name                        | Description                                     | Value |
| --------------------------- | ----------------------------------------------- | ----- |
| `global.imageRegistry`      | Global Docker image registry                    | `""`  |
| `global.imagePullSecrets`   | Global Docker registry secret names as an array| `[]`  |

### API Service Parameters

| Name                                    | Description                                | Value                    |
| --------------------------------------- | ------------------------------------------ | ------------------------ |
| `api.replicaCount`                      | Number of API replicas to deploy          | `1`                      |
| `api.image.repository`                  | API image repository                       | `video-to-podcast-api`   |
| `api.image.tag`                         | API image tag (immutable tags are recommended) | `latest`            |
| `api.image.pullPolicy`                  | API image pull policy                      | `IfNotPresent`           |
| `api.service.type`                      | API service type                           | `ClusterIP`              |
| `api.service.port`                      | API service HTTP port                      | `8080`                   |
| `api.service.targetPort`                | API service target port                    | `8080`                   |
| `api.resources.limits.memory`           | The memory limit for the API container     | `1Gi`                    |
| `api.resources.limits.cpu`              | The CPU limit for the API container        | `1000m`                  |
| `api.resources.requests.memory`         | The requested memory for the API container | `512Mi`                  |
| `api.resources.requests.cpu`            | The requested CPU for the API container    | `500m`                   |
| `api.healthCheck.enabled`               | Enable health checks for API               | `true`                   |
| `api.healthCheck.path`                  | Health check path                          | `/v1/health`             |
| `api.config.server.port`                | API server port                            | `"8080"`                 |
| `api.config.server.baseUrl`             | API server base URL                        | `""`                     |
| `api.config.database.connectionString`  | Database connection string                 | `":memory:"`             |
| `api.config.storage.basePath`           | Storage base path                          | `"/app/data/resources"`  |
| `api.config.external.ytdlpCookiesFile`  | YouTube cookies file path                  | `""`                     |
| `api.config.feed.mode`                  | Feed generation mode                       | `"per_directory"`        |

### UI Service Parameters

| Name                                    | Description                                | Value                    |
| --------------------------------------- | ------------------------------------------ | ------------------------ |
| `ui.replicaCount`                       | Number of UI replicas to deploy           | `1`                      |
| `ui.image.repository`                   | UI image repository                        | `video-to-podcast-ui`    |
| `ui.image.tag`                          | UI image tag (immutable tags are recommended) | `latest`             |
| `ui.image.pullPolicy`                   | UI image pull policy                       | `IfNotPresent`           |
| `ui.service.type`                       | UI service type                            | `ClusterIP`              |
| `ui.service.port`                       | UI service HTTP port                       | `3000`                   |
| `ui.service.targetPort`                 | UI service target port                     | `3000`                   |
| `ui.resources.limits.memory`            | The memory limit for the UI container      | `512Mi`                  |
| `ui.resources.limits.cpu`               | The CPU limit for the UI container         | `500m`                   |
| `ui.resources.requests.memory`          | The requested memory for the UI container  | `256Mi`                  |
| `ui.resources.requests.cpu`             | The requested CPU for the UI container     | `250m`                   |
| `ui.healthCheck.enabled`                | Enable health checks for UI                | `true`                   |
| `ui.healthCheck.path`                   | Health check path                          | `/health`                |
| `ui.config.server.port`                 | UI server port                             | `"3000"`                 |
| `ui.config.api.baseUrl`                 | API base URL for UI (auto-configured)      | `""`                     |
| `ui.config.api.timeout`                 | API timeout                                | `"30s"`                  |

### Persistence Parameters

| Name                                    | Description                                | Value           |
| --------------------------------------- | ------------------------------------------ | --------------- |
| `persistence.enabled`                   | Enable persistence using PVC               | `false`         |
| `persistence.resources.enabled`         | Enable persistent storage for resources    | `false`         |
| `persistence.resources.storageClass`    | PVC Storage Class for resources volume     | `""`            |
| `persistence.resources.accessMode`      | PVC Access Mode for resources volume       | `ReadWriteOnce` |
| `persistence.resources.size`            | PVC Storage Request for resources volume   | `10Gi`          |
| `persistence.database.enabled`          | Enable persistent storage for database     | `false`         |
| `persistence.database.storageClass`     | PVC Storage Class for database volume      | `""`            |
| `persistence.database.accessMode`       | PVC Access Mode for database volume        | `ReadWriteOnce` |
| `persistence.database.size`             | PVC Storage Request for database volume    | `1Gi`           |
| `persistence.cookies.enabled`           | Enable persistent storage for cookies      | `false`         |
| `persistence.cookies.storageClass`      | PVC Storage Class for cookies volume       | `""`            |
| `persistence.cookies.accessMode`        | PVC Access Mode for cookies volume         | `ReadWriteOnce` |
| `persistence.cookies.size`              | PVC Storage Request for cookies volume     | `100Mi`         |

### Ingress Parameters

| Name                       | Description                                | Value                    |
| -------------------------- | ------------------------------------------ | ------------------------ |
| `ingress.enabled`          | Enable ingress record generation           | `false`                  |
| `ingress.className`        | IngressClass that will be used             | `""`                     |
| `ingress.annotations`      | Additional annotations for the Ingress     | `{}`                     |
| `ingress.hosts[0].host`    | Default host for the ingress record        | `video-to-podcast.local` |
| `ingress.tls`              | TLS configuration for ingress              | `[]`                     |

### Autoscaling Parameters

| Name                                           | Description                                        | Value   |
| ---------------------------------------------- | -------------------------------------------------- | ------- |
| `autoscaling.api.enabled`                      | Enable Horizontal Pod Autoscaler for API          | `false` |
| `autoscaling.api.minReplicas`                  | Minimum number of API replicas                     | `1`     |
| `autoscaling.api.maxReplicas`                  | Maximum number of API replicas                     | `5`     |
| `autoscaling.api.targetCPUUtilizationPercentage` | Target CPU utilization percentage for API       | `80`    |
| `autoscaling.api.targetMemoryUtilizationPercentage` | Target Memory utilization percentage for API | `80`    |
| `autoscaling.ui.enabled`                       | Enable Horizontal Pod Autoscaler for UI           | `false` |
| `autoscaling.ui.minReplicas`                   | Minimum number of UI replicas                      | `1`     |
| `autoscaling.ui.maxReplicas`                   | Maximum number of UI replicas                      | `3`     |
| `autoscaling.ui.targetCPUUtilizationPercentage` | Target CPU utilization percentage for UI        | `80`    |
| `autoscaling.ui.targetMemoryUtilizationPercentage` | Target Memory utilization percentage for UI  | `80`    |

### Security Parameters

| Name                                    | Description                                | Value     |
| --------------------------------------- | ------------------------------------------ | --------- |
| `serviceAccount.create`                 | Specifies whether a service account should be created | `true` |
| `serviceAccount.annotations`            | Annotations to add to the service account  | `{}`      |
| `serviceAccount.name`                   | The name of the service account to use     | `""`      |
| `podSecurityContext.fsGroup`            | Set filesystem group ID                    | `1001`    |
| `podSecurityContext.runAsNonRoot`       | Run as non-root user                       | `true`    |
| `podSecurityContext.runAsUser`          | Set user ID                                | `1001`    |
| `podSecurityContext.runAsGroup`         | Set group ID                               | `1001`    |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation              | `false`   |
| `securityContext.capabilities.drop`     | Drop capabilities                          | `["ALL"]` |
| `securityContext.readOnlyRootFilesystem` | Set root filesystem as read-only          | `false`   |
| `securityContext.runAsNonRoot`          | Run as non-root user                       | `true`    |
| `securityContext.runAsUser`             | Set user ID                                | `1001`    |

### YouTube Cookies Parameters

| Name                           | Description                                | Value              |
| ------------------------------ | ------------------------------------------ | ------------------ |
| `youtubeCookies.enabled`       | Enable YouTube cookies secret              | `false`            |
| `youtubeCookies.secretName`    | Name of the secret containing cookies      | `youtube-cookies`  |
| `youtubeCookies.cookiesFile`   | Name of the cookies file                   | `youtube_cookies.txt` |

## Usage Examples

### Basic Installation

```bash
helm install my-video-to-podcast ./charts/video-to-podcast
```

### Installation with Persistence

```bash
helm install my-video-to-podcast ./charts/video-to-podcast \
  --set persistence.enabled=true \
  --set persistence.resources.enabled=true \
  --set persistence.database.enabled=true
```

### Installation with Ingress

```bash
helm install my-video-to-podcast ./charts/video-to-podcast \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=video-to-podcast.example.com
```

### Installation with Custom Values File

Create a `custom-values.yaml` file:

```yaml
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
    size: 50Gi
  database:
    enabled: true
    size: 5Gi

api:
  resources:
    requests:
      memory: 1Gi
      cpu: 500m
    limits:
      memory: 2Gi
      cpu: 1000m

autoscaling:
  api:
    enabled: true
    minReplicas: 2
    maxReplicas: 10
```

Then install:

```bash
helm install my-video-to-podcast ./charts/video-to-podcast -f custom-values.yaml
```

### Installation with YouTube Cookies

First, create a secret with your YouTube cookies:

```bash
kubectl create secret generic youtube-cookies \
  --from-file=youtube_cookies.txt=/path/to/your/cookies.txt
```

Then install with cookies enabled:

```bash
helm install my-video-to-podcast ./charts/video-to-podcast \
  --set youtubeCookies.enabled=true
```

## Accessing the Application

After installation, you can access the application:

1. **Port Forward (for testing)**:
   ```bash
   kubectl port-forward svc/my-video-to-podcast-ui 3000:3000
   kubectl port-forward svc/my-video-to-podcast-api 8080:8080
   ```
   Then access:
   - UI: http://localhost:3000
   - API: http://localhost:8080

2. **Via Ingress** (if enabled):
   Access via the configured ingress host.

3. **Via LoadBalancer** (if service type is LoadBalancer):
   ```bash
   kubectl get svc my-video-to-podcast-ui
   kubectl get svc my-video-to-podcast-api
   ```

## Upgrading

To upgrade the chart:

```bash
helm upgrade my-video-to-podcast ./charts/video-to-podcast
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/instance=my-video-to-podcast
```

### Check Logs

```bash
kubectl logs -l app.kubernetes.io/instance=my-video-to-podcast,app.kubernetes.io/component=api
kubectl logs -l app.kubernetes.io/instance=my-video-to-podcast,app.kubernetes.io/component=ui
```

### Check Services

```bash
kubectl get svc -l app.kubernetes.io/instance=my-video-to-podcast
```

### Check Ingress

```bash
kubectl get ingress -l app.kubernetes.io/instance=my-video-to-podcast
```

## License

This chart is licensed under the same license as the Video to Podcast Service project.
