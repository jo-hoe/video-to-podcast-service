# video-to-podcast

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.0](https://img.shields.io/badge/AppVersion-1.0.0-informational?style=flat-square)

A Helm chart for Video to Podcast Service - Convert YouTube videos to podcast feeds

**Homepage:** <https://github.com/jo-hoe/video-to-podcast-service>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| jo-hoe |  | <https://github.com/jo-hoe> |

## Source Code

* <https://github.com/jo-hoe/video-to-podcast-service>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| api | object | `{"affinity":{},"config":{"database":{"connectionString":":memory:"},"external":{"ytdlpCookiesFile":""},"feed":{"mode":"per_directory"},"server":{"baseUrl":"","port":"8080"},"storage":{"basePath":"/app/data/resources"}},"healthCheck":{"enabled":true,"failureThreshold":5,"initialDelaySeconds":30,"path":"/v1/health","periodSeconds":30,"successThreshold":1,"timeoutSeconds":10},"image":{"pullPolicy":"IfNotPresent","repository":"video-to-podcast-api","tag":"latest"},"nodeSelector":{},"replicaCount":1,"resources":{"limits":{"cpu":"1000m","memory":"1Gi"},"requests":{"cpu":"500m","memory":"512Mi"}},"service":{"annotations":{},"port":8080,"targetPort":8080,"type":"ClusterIP"},"tolerations":[]}` | API Service configuration |
| api.affinity | object | `{}` | Affinity rules for API pods (node/pod affinity/anti-affinity) |
| api.config | object | `{"database":{"connectionString":":memory:"},"external":{"ytdlpCookiesFile":""},"feed":{"mode":"per_directory"},"server":{"baseUrl":"","port":"8080"},"storage":{"basePath":"/app/data/resources"}}` | Application runtime configuration for API |
| api.config.database.connectionString | string | `":memory:"` | Database connection string (":memory:" for in-memory / sqlite path for file) |
| api.config.external.ytdlpCookiesFile | string | `""` | Optional path to yt-dlp cookies file inside container |
| api.config.feed.mode | string | `"per_directory"` | Feed generation mode (e.g., "per_directory") |
| api.config.server.baseUrl | string | `""` | External base URL for API (e.g., ingress URL); empty uses cluster service |
| api.config.server.port | string | `"8080"` | Port the API server listens on (string for templating consistency) |
| api.config.storage.basePath | string | `"/app/data/resources"` | Base path in container filesystem for resources (audio, etc.) |
| api.healthCheck | object | `{"enabled":true,"failureThreshold":5,"initialDelaySeconds":30,"path":"/v1/health","periodSeconds":30,"successThreshold":1,"timeoutSeconds":10}` | Liveness/Readiness probe configuration for API |
| api.healthCheck.enabled | bool | `true` | Enable/disable health probes for API |
| api.healthCheck.failureThreshold | int | `5` | Consecutive failures before marking container unhealthy |
| api.healthCheck.initialDelaySeconds | int | `30` | Initial delay before starting health checks (seconds) |
| api.healthCheck.path | string | `"/v1/health"` | HTTP path used for API health endpoint |
| api.healthCheck.periodSeconds | int | `30` | Interval between health checks (seconds) |
| api.healthCheck.successThreshold | int | `1` | Consecutive successes before marking container healthy |
| api.healthCheck.timeoutSeconds | int | `10` | Probe timeout (seconds) |
| api.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for API |
| api.image.repository | string | `"video-to-podcast-api"` | API container image repository |
| api.image.tag | string | `"latest"` | API image tag to deploy |
| api.nodeSelector | object | `{}` | Map of node labels for API pods to be scheduled on specific nodes |
| api.replicaCount | int | `1` | Number of API replicas to run |
| api.resources.limits | object | `{"cpu":"1000m","memory":"1Gi"}` | Resource limits for API container |
| api.resources.requests | object | `{"cpu":"500m","memory":"512Mi"}` | Resource requests for API container |
| api.service.annotations | object | `{}` | Additional annotations for the API Service |
| api.service.port | int | `8080` | Service port exposed by the API Service |
| api.service.targetPort | int | `8080` | Target container port for the API |
| api.service.type | string | `"ClusterIP"` | Kubernetes Service type for API (ClusterIP, NodePort, LoadBalancer) |
| api.tolerations | list | `[]` | Tolerations for API pods to schedule onto tainted nodes |
| autoscaling | object | `{"api":{"enabled":false,"maxReplicas":5,"minReplicas":1,"targetCPUUtilizationPercentage":80,"targetMemoryUtilizationPercentage":80},"ui":{"enabled":false,"maxReplicas":3,"minReplicas":1,"targetCPUUtilizationPercentage":80,"targetMemoryUtilizationPercentage":80}}` | Horizontal Pod Autoscaling (HPA) configuration |
| autoscaling.api.enabled | bool | `false` | Enable HPA for API deployment |
| autoscaling.api.maxReplicas | int | `5` | Maximum number of API replicas |
| autoscaling.api.minReplicas | int | `1` | Minimum number of API replicas |
| autoscaling.api.targetCPUUtilizationPercentage | int | `80` | Target average CPU utilization percentage for scaling |
| autoscaling.api.targetMemoryUtilizationPercentage | int | `80` | Target average memory utilization percentage for scaling |
| autoscaling.ui.enabled | bool | `false` | Enable HPA for UI deployment |
| autoscaling.ui.maxReplicas | int | `3` | Maximum number of UI replicas |
| autoscaling.ui.minReplicas | int | `1` | Minimum number of UI replicas |
| autoscaling.ui.targetCPUUtilizationPercentage | int | `80` | Target average CPU utilization percentage for scaling |
| autoscaling.ui.targetMemoryUtilizationPercentage | int | `80` | Target average memory utilization percentage for scaling |
| extraEnvVars | list | `[]` | Additional environment variables injected into both API and UI pods |
| extraVolumeMounts | list | `[]` | Additional volume mounts for containers (match names with extraVolumes) |
| extraVolumes | list | `[]` | Additional volumes to mount into pods (for advanced customization) |
| global | object | `{"imagePullSecrets":[],"imageRegistry":""}` | Global configuration applied across the chart (if supported by templates) |
| global.imagePullSecrets | list | `[]` | List of secrets to use for pulling images from private registries |
| global.imageRegistry | string | `""` | Optional global image registry override (e.g., "my-registry.io") |
| ingress | object | `{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"video-to-podcast.local","paths":[{"path":"/","pathType":"Prefix","service":"ui"},{"path":"/v1","pathType":"Prefix","service":"api"}]}],"tls":[]}` | Ingress configuration to expose UI/API via HTTP(S) |
| ingress.annotations | object | `{}` | Additional annotations for ingress resources |
| ingress.className | string | `""` | IngressClass name (e.g., "nginx") |
| ingress.enabled | bool | `false` | Enable ingress resources |
| ingress.hosts | list | `[{"host":"video-to-podcast.local","paths":[{"path":"/","pathType":"Prefix","service":"ui"},{"path":"/v1","pathType":"Prefix","service":"api"}]}]` | kubernetes.io/tls-acme: "true" |
| ingress.tls | list | `[]` | TLS configuration for ingress (list of hosts mapped to secrets) |
| monitoring | object | `{"serviceMonitor":{"enabled":false,"interval":"30s","labels":{},"namespace":"","scrapeTimeout":"10s"}}` | Monitoring configuration (Prometheus ServiceMonitor) |
| monitoring.serviceMonitor.enabled | bool | `false` | Enable creation of a ServiceMonitor for scraping metrics |
| monitoring.serviceMonitor.interval | string | `"30s"` | Scrape interval (e.g., "30s") |
| monitoring.serviceMonitor.labels | object | `{}` | Additional labels to attach to ServiceMonitor |
| monitoring.serviceMonitor.namespace | string | `""` | Namespace where ServiceMonitor should be created (empty = same namespace) |
| monitoring.serviceMonitor.scrapeTimeout | string | `"10s"` | Scrape timeout (e.g., "10s") |
| networkPolicy | object | `{"egress":[],"enabled":false,"ingress":[]}` | Network Policy configuration to restrict traffic |
| networkPolicy.egress | list | `[]` | Egress rules for allowing outgoing traffic (advanced) |
| networkPolicy.enabled | bool | `false` | Enable creation of NetworkPolicy resources |
| networkPolicy.ingress | list | `[]` | Ingress rules for allowing incoming traffic (advanced) |
| persistence | object | `{"cookies":{"accessMode":"ReadWriteOnce","annotations":{},"enabled":false,"size":"100Mi","storageClass":""},"database":{"accessMode":"ReadWriteOnce","annotations":{},"enabled":false,"size":"1Gi","storageClass":""},"enabled":false,"resources":{"accessMode":"ReadWriteOnce","annotations":{},"enabled":false,"size":"10Gi","storageClass":""}}` | Persistence configuration for application data |
| persistence.cookies | object | `{"accessMode":"ReadWriteOnce","annotations":{},"enabled":false,"size":"100Mi","storageClass":""}` | Storage for cookies |
| persistence.cookies.accessMode | string | `"ReadWriteOnce"` | Access mode for cookies PVC |
| persistence.cookies.annotations | object | `{}` | Additional annotations for cookies PVC |
| persistence.cookies.enabled | bool | `false` | Enable PVC for cookies storage |
| persistence.cookies.size | string | `"100Mi"` | Requested size for cookies PVC |
| persistence.cookies.storageClass | string | `""` | StorageClass for cookies PVC (empty uses cluster default) |
| persistence.database | object | `{"accessMode":"ReadWriteOnce","annotations":{},"enabled":false,"size":"1Gi","storageClass":""}` | Storage for database |
| persistence.database.accessMode | string | `"ReadWriteOnce"` | Access mode for database PVC |
| persistence.database.annotations | object | `{}` | Additional annotations for database PVC |
| persistence.database.enabled | bool | `false` | Enable PVC for database storage |
| persistence.database.size | string | `"1Gi"` | Requested size for database PVC |
| persistence.database.storageClass | string | `""` | StorageClass for database PVC (empty uses cluster default) |
| persistence.enabled | bool | `false` | Enable PersistentVolumeClaims for data persistence |
| persistence.resources | object | `{"accessMode":"ReadWriteOnce","annotations":{},"enabled":false,"size":"10Gi","storageClass":""}` | Storage for resources (audio files, etc.) |
| persistence.resources.accessMode | string | `"ReadWriteOnce"` | Access mode for resources PVC |
| persistence.resources.annotations | object | `{}` | Additional annotations for resources PVC |
| persistence.resources.enabled | bool | `false` | Enable PVC for resources storage |
| persistence.resources.size | string | `"10Gi"` | Requested size for resources PVC |
| persistence.resources.storageClass | string | `""` | StorageClass for resources PVC (empty uses cluster default) |
| podAnnotations | object | `{}` | Additional annotations to add to pods |
| podDisruptionBudget | object | `{"enabled":false,"minAvailable":1}` | Pod Disruption Budget for high availability |
| podDisruptionBudget.enabled | bool | `false` | Enable creation of a PodDisruptionBudget |
| podDisruptionBudget.minAvailable | int | `1` | Minimum number of pods that must be available |
| podLabels | object | `{}` | Additional labels to add to pods |
| podSecurityContext | object | `{"fsGroup":1001,"runAsGroup":1001,"runAsNonRoot":true,"runAsUser":1001}` | Pod Security Context applied to all pods (unless overridden) |
| podSecurityContext.fsGroup | int | `1001` | Filesystem group ID for mounted volumes |
| podSecurityContext.runAsGroup | int | `1001` | Primary group ID the container runs as |
| podSecurityContext.runAsNonRoot | bool | `true` | Run containers as non-root user if true |
| podSecurityContext.runAsUser | int | `1001` | User ID the container runs as |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":false,"runAsNonRoot":true,"runAsUser":1001}` | Container Security Context defaults |
| securityContext.allowPrivilegeEscalation | bool | `false` | Prevent privilege escalation |
| securityContext.capabilities | object | `{"drop":["ALL"]}` | Linux capabilities to drop |
| securityContext.readOnlyRootFilesystem | bool | `false` | Mount container root filesystem as read-only |
| securityContext.runAsNonRoot | bool | `true` | Run containers as non-root user if true |
| securityContext.runAsUser | int | `1001` | User ID the container runs as |
| serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | Service Account configuration |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | Name of the service account to use (if empty, a name is generated) |
| ui | object | `{"affinity":{},"config":{"api":{"baseUrl":"","timeout":"30s"},"server":{"port":"3000"}},"healthCheck":{"enabled":true,"failureThreshold":5,"initialDelaySeconds":20,"path":"/health","periodSeconds":30,"successThreshold":1,"timeoutSeconds":10},"image":{"pullPolicy":"IfNotPresent","repository":"video-to-podcast-ui","tag":"latest"},"nodeSelector":{},"replicaCount":1,"resources":{"limits":{"cpu":"500m","memory":"512Mi"},"requests":{"cpu":"250m","memory":"256Mi"}},"service":{"annotations":{},"port":3000,"targetPort":3000,"type":"ClusterIP"},"tolerations":[]}` | UI Service configuration |
| ui.affinity | object | `{}` | Affinity rules for UI pods (node/pod affinity/anti-affinity) |
| ui.config | object | `{"api":{"baseUrl":"","timeout":"30s"},"server":{"port":"3000"}}` | Application runtime configuration for UI |
| ui.config.api.timeout | string | `"30s"` | Timeout for API requests from UI |
| ui.config.server.port | string | `"3000"` | Port the UI server listens on (string for templating consistency) |
| ui.healthCheck | object | `{"enabled":true,"failureThreshold":5,"initialDelaySeconds":20,"path":"/health","periodSeconds":30,"successThreshold":1,"timeoutSeconds":10}` | Liveness/Readiness probe configuration for UI |
| ui.healthCheck.enabled | bool | `true` | Enable/disable health probes for UI |
| ui.healthCheck.failureThreshold | int | `5` | Consecutive failures before marking container unhealthy |
| ui.healthCheck.initialDelaySeconds | int | `20` | Initial delay before starting health checks (seconds) |
| ui.healthCheck.path | string | `"/health"` | HTTP path used for UI health endpoint |
| ui.healthCheck.periodSeconds | int | `30` | Interval between health checks (seconds) |
| ui.healthCheck.successThreshold | int | `1` | Consecutive successes before marking container healthy |
| ui.healthCheck.timeoutSeconds | int | `10` | Probe timeout (seconds) |
| ui.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for UI |
| ui.image.repository | string | `"video-to-podcast-ui"` | UI container image repository |
| ui.image.tag | string | `"latest"` | UI image tag to deploy |
| ui.nodeSelector | object | `{}` | Node scheduling hints for UI Map of node labels for UI pods to be scheduled on specific nodes |
| ui.replicaCount | int | `1` | Number of UI replicas to run |
| ui.resources.limits | object | `{"cpu":"500m","memory":"512Mi"}` | Resource limits for UI container |
| ui.resources.requests | object | `{"cpu":"250m","memory":"256Mi"}` | Resource requests for UI container |
| ui.service.annotations | object | `{}` | Additional annotations for the UI Service |
| ui.service.port | int | `3000` | Service port exposed by the UI Service |
| ui.service.targetPort | int | `3000` | Target container port for the UI |
| ui.service.type | string | `"ClusterIP"` | Kubernetes Service type for UI (ClusterIP, NodePort, LoadBalancer) |
| ui.tolerations | list | `[]` | Tolerations for UI pods to schedule onto tainted nodes |
| youtubeCookies | object | `{"cookiesFile":"youtube_cookies.txt","enabled":false,"secretName":"youtube-cookies"}` | YouTube cookies secret configuration (optional) |
| youtubeCookies.cookiesFile | string | `"youtube_cookies.txt"` | File name inside the secret containing cookies |
| youtubeCookies.enabled | bool | `false` | Enable mounting a secret containing YouTube cookies |
| youtubeCookies.secretName | string | `"youtube-cookies"` | Name of the Kubernetes Secret with the cookies |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
