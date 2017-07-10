load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_prefix")

go_prefix("github.com/vrischmann/dumpserver")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/dustin/go-humanize:go_default_library",
        "//vendor/github.com/vrischmann/flagutil:go_default_library",
    ],
)

go_binary(
    name = "dumpserver",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
