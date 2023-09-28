source "macstadium-orka" "image" {
	source_image    = "90GCatalinaSSH.img"
	image_name      = "packer-{{timestamp}}"
	orka_endpoint   = "http://10.221.188.100"
	orka_user       = "victor@ms.com"
	orka_password   = "password"
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
