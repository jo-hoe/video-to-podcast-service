# video-to-podcast-service

![Version: 1.9.1](https://img.shields.io/badge/Version-1.9.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.9.1](https://img.shields.io/badge/AppVersion-1.9.1-informational?style=flat-square)

Service for converting videos to podcast feeds

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| cookies.cookieContent | string | `""` | Base64 encoded content of the YouTube cookies file (optional) If provided, a secret will be automatically created with this content Example: cookieContent: "BASE64_ENCODED_COOKIE_STRING" |
| cookies.cookiePath | string | `"/app/data/cookies/youtube-cookies.txt"` | Path is relative to the mount point in the container |
| cookies.enabled | bool | `false` |  |
| cookies.secretName | string | `""` | Secret name containing the cookies file (optional) If provided, will use the existing secret instead of creating one Note: secretName takes precedence over cookieContent |
| database.connectionString | string | `"file:/app/data/database/video-to-podcast-service.db"` |  |
| database.driver | string | `"sqlite3"` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/jo-hoe/video-to-podcast-service"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"video-to-podcast.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"Prefix"` |  |
| ingress.tls | list | `[]` |  |
| livenessProbe.failureThreshold | int | `3` |  |
| livenessProbe.httpGet.path | string | `"/health"` |  |
| livenessProbe.httpGet.port | string | `"http"` |  |
| livenessProbe.initialDelaySeconds | int | `30` |  |
| livenessProbe.periodSeconds | int | `10` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| logLevel | string | `"info"` |  |
| media | object | `{"allowPartialDownloads":true,"maxParallelDownloads":1,"mediaPath":"/app/data/resources/media","tempPath":"/app/data/resources/temp"}` | Media configuration |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| persistence.accessMode | string | `"ReadWriteOnce"` | Access mode for the persistent volume |
| persistence.annotations | object | `{}` | Annotations for the PVC |
| persistence.enabled | bool | `false` | WARNING: Without persistence, all data (database, media files, cookies) will be lost on pod restart! |
| persistence.existingClaim | string | `""` | Existing claim name if you want to use an existing PVC |
| persistence.size | string | `"10Gi"` | Size of the persistent volume |
| persistence.storageClass | string | `""` | Storage class for the persistent volume |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.fsGroup | int | `1000` |  |
| readinessProbe.failureThreshold | int | `3` |  |
| readinessProbe.httpGet.path | string | `"/health"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| readinessProbe.initialDelaySeconds | int | `10` |  |
| readinessProbe.periodSeconds | int | `5` |  |
| readinessProbe.timeoutSeconds | int | `3` |  |
| resources.limits | object | `{"cpu":"1000m","memory":"1Gi"}` | We usually recommend not to specify default resources and to leave this as a conscious choice for the user. This also increases chances charts run on environments with little resources, such as Minikube. If you do want to specify resources, uncomment the following lines, adjust them as necessary, and remove the curly braces after 'resources:'. |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `false` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1000` |  |
| service.enabled | bool | `true` |  |
| service.port | int | `8081` |  |
| service.targetPort | int | `8080` |  |
| service.type | string | `"LoadBalancer"` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials? |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` |  |
| ytDlp | object | `{"binaryPvc":{"size":"128Mi","storageClass":""},"poTokenSidecar":{"enabled":true,"image":{"pullPolicy":"IfNotPresent","repository":"brainicism/bgutil-ytdlp-pot-provider","tag":"latest"},"resources":{"limits":{"cpu":"200m","memory":"256Mi"},"requests":{"cpu":"50m","memory":"128Mi"}}},"updateToNightly":false}` | yt-dlp configuration |
| ytDlp.binaryPvc | object | `{"size":"128Mi","storageClass":""}` | PVC used by the initContainer to store the yt-dlp binary. A separate small PVC avoids coupling the binary to the app data volume. The PVC is not a cache — it is a handoff mechanism between the initContainer (runs as root, writes the binary) and the main container (reads it as appuser). The initContainer re-downloads on every pod start, so a pod restart always picks up the latest build in the selected channel. This is intentional: when updateToNightly is true a restart is the mechanism to get a newer nightly. |
| ytDlp.poTokenSidecar | object | `{"enabled":true,"image":{"pullPolicy":"IfNotPresent","repository":"brainicism/bgutil-ytdlp-pot-provider","tag":"latest"},"resources":{"limits":{"cpu":"200m","memory":"256Mi"},"requests":{"cpu":"50m","memory":"128Mi"}}}` | PO token sidecar configuration. The bgutil-ytdlp-pot-provider HTTP server runs as a sidecar container and automatically supplies Proof-of-Origin tokens to yt-dlp, which makes traffic appear more legitimate to YouTube and reduces 403 errors.  Failure behavior (by design): - Sidecar crash: K8s restarts it via the liveness probe. While it is down   the bgutil plugin raises PoTokenProviderRejectedRequest (not a hard error)   so yt-dlp continues without a PO token — same behavior as without sidecar. - Invalid tokens (e.g. YouTube updates Botguard): downloads may 403, same as   without the sidecar. Use updateToNightly as the first mitigation lever. - Plugin goes unmaintained: graceful degradation as above. No hard dependency. |
| ytDlp.poTokenSidecar.enabled | bool | `true` | Enable the sidecar container that provides PO tokens to yt-dlp. |
| ytDlp.updateToNightly | bool | `false` | Pull the nightly build of yt-dlp instead of the version baked into the image. The initContainer runs as root and writes the binary to a dedicated PVC that is mounted read-only by the main container, so appuser never needs write access. Enable when the stable release is broken and a nightly fix is already available. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
