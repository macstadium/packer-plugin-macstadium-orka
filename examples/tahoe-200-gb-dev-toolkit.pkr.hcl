packer {
  required_plugins {
    macstadium-orka = {
      version = "= 3.0.1"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}
variable "source_image" {
  default = "ghcr.io/macstadium/orka-images/tahoe:200-gb"
}
variable "image_name_prefix" {
  default = "packer"
}
variable "orka_endpoint" {
  default = env("ORKA_ENDPOINT")
}
variable "orka_auth_token" {
  default = env("ORKA_AUTH_TOKEN")
}
variable "ssh_username" {
  default = "admin"
}
variable "ssh_password" {
  default = "admin"
}

source "macstadium-orka" "image" {
  source_image      = var.source_image // This image has the latest version of Orka VM tools already pre-installed 
  image_name        = "${var.image_name_prefix}-{{timestamp}}"
  image_description = "MacOS Tahoe 26.1 200 GB base image, plus commonly-used MacOS developer tools, created with Packer!"
  orka_endpoint     = var.orka_endpoint
  orka_auth_token   = var.orka_auth_token
  ssh_username      = var.ssh_username
  ssh_password      = var.ssh_password
}

build {
  sources = [
    "macstadium-orka.image"
  ]

  provisioner "shell" {
    inline = [
      "echo 'Setting up passwordless sudo for admin user'",
      "echo '${var.ssh_password}' | sudo -S sh -c \"echo '${var.ssh_username} ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/${var.ssh_username}-nopasswd\"",
      "echo '${var.ssh_password}' | sudo -S chmod 0644 /etc/sudoers.d/${var.ssh_username}-nopasswd",
      "echo 'Installing Homebrew'",
      "echo '${var.ssh_password}' | sudo -S mkdir -p /opt/homebrew",
      "echo '${var.ssh_password}' | sudo -S chown -R ${var.ssh_username}:${var.ssh_username} /opt/homebrew",
      "NONINTERACTIVE=1 /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"",
      "echo 'Cleaning up passwordless sudo'",
      "echo '${var.ssh_password}' | sudo -S rm -f /etc/sudoers.d/${var.ssh_username}-nopasswd",
      "echo 'Homebrew installation completed'"
    ]
  }

  provisioner "shell" {
    inline = [
      "# Add Homebrew to PATH in shell configuration files, use Homebrew to install Fastlane, swiftlint, Git, swift, and Cocoapods",
      // Add or delete tools from this section as needed for your use case, XCodes will require your AppleID and password to install whichever version of XCode you specify.
      "echo >> /Users/${var.ssh_username}/.zprofile",
      "echo 'eval \"$(/opt/homebrew/bin/brew shellenv\"' >> /Users/${var.ssh_username}/.zprofile",
      "eval \"$(/opt/homebrew/bin/brew shellenv)\"",
      "brew install fastlane",
      "brew install git",
      "brew install cocoapods",
      "brew install swift",
      "",
      "# Note: Install Xcode manually or use xcodes after image is created",
      "# xcodes requires Command Line Tools which need interactive installation",
    ]
  }
}