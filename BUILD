load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

##
## Libraries.
##
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
