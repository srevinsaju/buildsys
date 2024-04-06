togomak {
  version = 2
}

stage "colors" {
  for_each = toset([
    "error",
    "success",
    "warn",
    "warning",
    "info",
    "red",
    "green",
    "blue",
    "yellow",
    "bold",
    "italic",
    "cyan",
    "grey",
    "white",
    "magenta",
    "orange",
    "hi-green",
    "hi-blue",
    "hi-magenta",
    "hi-black",
    "hi-white",
    "hi-red",
    "hi-cyan",
    "hi-yellow",
  ])

  script = "echo ${ansifmt(each.key, "hello world")}"
}
