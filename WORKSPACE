workspace(name = "go_infra_provisioner")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# Go rules
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "6dc2da7ab4cf5d7bfc7c949776b1b7c733f05e56edc4bcd9022bb249d2e2a996",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/v0.39.1/rules_go-v0.39.1.zip"],
)

# Gazelle
http_archive(
    name = "bazel_gazelle",
    sha256 = "727f3e4edd96ea20c29e8c2ca9e8d2af724d8c7778e7923a854b2c80952bc405",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.30.0/bazel-gazelle-v0.30.0.tar.gz"],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

# Initialize Go toolchain
go_rules_dependencies()

# Set specific Go version - make sure this matches your local Go version
go_register_toolchains(
    version = "1.21.1",
    nogo = "@//:nogo",  # Optional static analysis tool
)

# Initialize Gazelle
gazelle_dependencies()

# AWS SDK dependencies
load("@bazel_gazelle//:deps.bzl", "go_repository")

# AWS SDK dependencies will be added by gazelle