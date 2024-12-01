# Use a base image with Bazel pre-installed
FROM l.gcr.io/google/bazel:6.4.0

# Install Go
RUN apt-get update && apt-get install -y \
    golang-go \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /workspace

# Copy Bazel configuration files
COPY .bazelrc WORKSPACE BUILD.bazel ./

# Copy Go module files
COPY go.mod go.sum ./

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY policies/ policies/
COPY terraform/ terraform/

# Build the application using Bazel
RUN bazel build //cmd/api:api

# Expose port
EXPOSE 8080

# Run the application
CMD ["bazel", "run", "//cmd/api:api"]