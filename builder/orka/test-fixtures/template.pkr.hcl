source "macstadium-orka" "image" {
  orka_endpoint   = "http://10.221.188.100"
  orka_auth_token = "myauthtoken"
  source_image    = "90gbsonomassh.orkasi"
  image_name      = "my-packer-image"
  orka_vm_builder_namespace = "my-namespace"
  orka_vm_builder_name = "my-vm-name"
  no_create_image = false
  no_delete_vm    = false
  mock { error_type = "none" }
}

build {
  sources = ["sources.macstadium-orka.image"]
  provisioner "shell" {
    inline = [
      "echo we are running on the remote host",
      "hostname",
      "touch .we-ran-packer-successfully"
    ]
  }
}
