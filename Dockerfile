# Build stage: build the Go application
FROM golang:1.23.4-alpine3.20 AS build

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to leverage caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code and build the application
COPY . ./
# Build with no CGO and output to a specific location
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Runtime stage: use a minimal base image for runtime and set up a non-root user
FROM jrottenberg/ffmpeg:7.1-ubuntu

# Install ca-certificates
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Create a non-root user and set up directories
RUN useradd --create-home --shell /bin/bash appuser
RUN mkdir -p /home/appuser/app/resources

# Set the working directory
WORKDIR /home/appuser/app

# Copy the built application from the build stage
COPY --from=build /go/bin/app .

# Ensure the executable has execute permissions
RUN chmod +x /home/appuser/app

# Set the user to the non-root user
USER appuser

# ENTRYPOINT should point to the executable, not a directory
ENTRYPOINT ["./app"]
