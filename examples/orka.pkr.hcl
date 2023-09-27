variable "ORKA_IMAGE" {
  default = env("ORKA_IMAGE")
}
variable "ORKA_ENDPOINT" {
  default = env("ORKA_ENDPOINT")
}
variable "ORKA_USER" {
  default = env("ORKA_USER")
}
variable "ORKA_PASSWORD" {
  default = env("ORKA_PASSWORD")
}
variable "ORKA_IMAGE_NAME_PREFIX" {
  default = "packer"
}

source "macstadium-orka" "image" {
  source_image    = var.ORKA_IMAGE
  image_name      = "${var.ORKA_IMAGE_NAME_PREFIX}-{{timestamp}}"
  orka_endpoint   = var.ORKA_ENDPOINT
  orka_user       = var.ORKA_USER
  orka_password   = var.ORKA_PASSWORD
  simulate_create = false
  no_create_image = false
  no_delete_vm    = false
}

build {
  sources = [
    "macstadium-orka.image"
  ]
   provisioner "shell" {
    inline = [
      "echo we are running on the remote host",
      "hostname",
      "touch .we-ran-packer-successfully"
    ]
  }
}
