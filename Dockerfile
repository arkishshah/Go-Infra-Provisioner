# Use Ubuntu as base image
FROM ubuntu:22.04

# Install required packages including Bazel
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    golang-go \
    python3 \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Install Bazel
RUN curl -fsSL https://bazel.build/bazel-release.pub.gpg | gpg --dearmor > bazel.gpg \
    && mv bazel.gpg /etc/apt/trusted.gpg.d/ \
    && echo "deb [arch=amd64] https://storage.googleapis.com/bazel-apt stable jdk1.8" | tee /etc/apt/sources.list.d/bazel.list \
    && apt-get update && apt-get install -y bazel \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /workspace

# Copy Go module files first
COPY go.mod go.sum ./

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY terraform/ terraform/
COPY configs/ configs/

# Build the Go application directly (without Bazel for now)
RUN go build -o main cmd/api/main.go

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]