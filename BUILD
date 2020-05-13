load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

##
## Binaries.
##
go_binary(
    name = "main",
    srcs = ["main.go"],
    deps = [
        ":breedgraph",
        ":flower",
    ],
)

##
## Libraries.
##
go_library(
    name = "breedgraph",
    srcs = ["breed_graph.go"],
    importpath = "github.com/BranLwyd/acnh_flowers/breedgraph",
    deps = [":flower"],
)

go_library(
    name = "flower",
    srcs = ["flower.go"],
    importpath = "github.com/BranLwyd/acnh_flowers/flower",
)

go_test(
    name = "flower_test",
    timeout = "short",
    srcs = ["flower_test.go"],
    embed = [":flower"],
)
