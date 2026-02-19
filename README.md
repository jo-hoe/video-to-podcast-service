# Video To Podcast Service

[![Test Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/test/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/video-to-podcast-service)](https://goreportcard.com/report/github.com/jo-hoe/video-to-podcast-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/video-to-podcast-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/video-to-podcast-service?branch=main)

Video To Podcast Service is a backend service that downloads video files (currently only from YouTube), extracts and converts them into audio files, and organizes them into podcast feeds accessible via RSS. The service exposes a REST API for adding new videos, listing available podcast feeds, retrieving audio files, and deleting podcast items.

## Deployment Options

### Docker Deployment

See the sections below for Docker and Docker Compose deployment.

### Kubernetes (k3s) Deployment

The service can be deployed to Kubernetes (specifically tested on k3s) using Helm:

```bash
# Basic installation
helm install video-to-podcast ./charts/video-to-podcast-service

# With custom values
helm install video-to-podcast ./charts/video-to-podcast-service -f custom-values.yaml
```

For detailed Kubernetes deployment instructions, see the [Helm Chart README](./charts/video-to-podcast-service/README.md).

### Local k3d Development Cluster

For local development and testing, you can use k3d (k3s in Docker):

```bash
# Start k3d cluster and deploy the service
make start-k3d

# Access the service
# The service will be available at http://localhost:8080

# Stop the cluster
make stop-k3d

# Restart the cluster
make restart-k3d
```

## How to Use

### Initial Setup

After cloning the repository, install the git hooks:

```bash
make install-hooks
```

This installs a pre-commit hook that automatically runs `go fmt` on all Go files before each commit, ensuring consistent code formatting across the project.

### Start the Service

You can start the service using `make` (recommended):

```bash
make start
```

Or use Docker directly:

```bash
docker build . -t v2p
docker run --rm -p 8080:8080 v2p
```

Or with Docker Compose (includes optional mail webhook):

```bash
make start-service
# or
make start-services-rebuild
```

### Resources

All downloaded resources are placed in the `resources` directory. Podcasts are organized in subdirectories named after the channel the video belongs to. Each feed has its own directory containing audio files and the RSS XML.

### Temporary Files

During video download and processing, temporary files are stored in a configurable temp directory:

- **Configuration**: Set via the `mediaConfig.TempPath` setting in the application configuration (defaults to `./mount/resources/temp`)
- **Cleanup**: Temporary directories are automatically cleaned up after processing completes

## API Usage

The service exposes a REST API. See [`openapi.yaml`](./openapi.yaml) for the full OpenAPI/Swagger specification.

## Linting

The project uses `golangci-lint` for linting. See <https://golangci-lint.run/docs/welcome/install/> for installation instructions.

To run linting locally:

```bash
golangci-lint run ./...
```

## Limitations

- Only YouTube is supported as a video source.
- Google may block certain IPs (e.g., from cloud providers), resulting in errors like `403` or age restriction issues. See [this GitHub issue](https://github.com/kkdai/youtube/issues/343#issuecomment-2347950479) for more details.

## Future Work

- Add iTunes to
  - add tags for podcast images (requires custom XML generation)
  - set length in podcast metadata
- Provide ticketing/progress feedback via API
- Auto-chapterize videos without chapters

## Relevant Links

- [ID3 Tags](https://www.exiftool.org/TagNames/ID3.html)
- [Example podcast](https://feeds.libsyn.com/230510/rss)
