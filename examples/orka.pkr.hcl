variable "ORKA_IMAGE" {
  default = env("ORKA_IMAGE")
}
variable "ORKA_ENDPOINT" {
  default = env("ORKA_ENDPOINT")
}
variable "ORKA_TOKEN" {
  default = env("ORKA_TOKEN")
}
variable "ORKA_IMAGE_NAME_PREFIX" {
  default = "packer"
}

source "macstadium-orka" "image" {
  source_image      = var.ORKA_IMAGE
  image_name        = "${var.ORKA_IMAGE_NAME_PREFIX}-{{timestamp}}"
  image_description = "I was created with Packer !"
  orka_endpoint     = var.ORKA_ENDPOINT
  orka_auth_token   = var.ORKA_TOKEN
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
