load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/whs-dot-hk/go-current-tempature
gazelle(name = "gazelle")

go_library(
    name = "go-current-tempature_lib",
    srcs = ["main.go"],
    importpath = "github.com/whs-dot-hk/go-current-tempature",
    visibility = ["//visibility:private"],
    deps = ["@com_github_gorilla_mux//:mux"],
)

go_binary(
    name = "go-current-tempature",
    embed = [":go-current-tempature_lib"],
    visibility = ["//visibility:public"],
)

load("@io_bazel_rules_docker//go:image.bzl", "go_image")

go_image(
    name = "go_image",
    embed = [":go-current-tempature_lib"],
    importpath = "github.com/whs-dot-hk/go-current-tempature",
    visibility = ["//visibility:public"],
)

load("@io_bazel_rules_docker//container:container.bzl", "container_push")
load("//:private.bzl", "AWS_ACCOUNT_ID", "AWS_REGION")

container_push(
    name = "push_go_image",
    format = "OCI",
    image = ":go_image",
    registry = "{accountId}.dkr.ecr.{region}.amazonaws.com".format(
        accountId = AWS_ACCOUNT_ID,
        region = AWS_REGION,
    ),
    repository = "go-current-tempature",
    tag = "latest",
)
