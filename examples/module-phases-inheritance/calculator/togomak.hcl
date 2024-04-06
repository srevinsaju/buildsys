togomak {
  version = 2
}

variable "a" {
  type = number
  description = "first variable"
}
variable "b" {
  type = number 
  description = "second variable"
}

variable "operation" {
  type = string 
  description = "Operation to perform, any of: [add, subtract, multiply, divide]"
}

stage "add" {
  if = var.operation == "add"
  script = "echo ${var.a} + ${var.b} is ${var.a + var.b}"
}

stage "paths" {
  script = <<-EOT
  echo path.module: ${path.module}
  echo path.cwd: ${path.cwd}
  echo path.root: ${ansifmt("green", path.root)}
  EOT
}
