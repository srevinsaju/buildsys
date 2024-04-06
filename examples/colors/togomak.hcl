togomak {
  version = 2
}

stage "colors" {
  script = "echo ${ansifmt("success", "hello world")}"
}
