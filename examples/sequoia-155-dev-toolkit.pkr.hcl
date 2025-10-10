packer {
  required_plugins {
    macstadium-orka = {
      version = "= 3.0.1"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}
variable "source_image" {
  default = "ghcr.io/macstadium/orka-images/sequoia:latest"
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
variable "orka_vm_tools_version" {
 type    = string
 description = "Target Orka VM Tools version to install"
 default   = "3.5.0"
}

source "macstadium-orka" "image" {
  source_image      = var.source_image // This image has the latest version of Orka VM tools already pre-installed 
  image_name        = "${var.image_name_prefix}-{{timestamp}}"
  image_description = "MacOS Sequoia 15.5 developer tools image created with Packer!"
  orka_endpoint     = var.orka_endpoint
  orka_auth_token   = var.orka_auth_token
  ssh_username      = var.ssh_username
  ssh_password      = var.ssh_password
  orka_vm_tools_version = var.orka_vm_tools_version
}

build {
  sources = [
    "macstadium-orka.image"
  ]

  provisioner "shell" {
    inline = [
      "echo 'admin' | sudo -S sh -c 'echo \"admin ALL=(ALL) NOPASSWD: ALL\" >> /etc/sudoers'",
      "echo 'Installing Homebrew'",
      "NONINTERACTIVE=1 /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"",
    ]
  }

  provisioner "shell" {
    inline = [
      "# Update Orka VM Tools to specified version",
      "echo 'Checking current Orka VM Tools version...'",
      "brew list --versions macstadium/orka/orka-vm-tools || echo 'Orka VM Tools not installed via Homebrew'",
      "echo 'Ensuring MacStadium Orka tap is available...'",
      "brew tap macstadium/orka || true",
      "echo 'Uninstalling old Orka VM Tools if present...'",
      "brew uninstall macstadium/orka/orka-vm-tools --force --ignore-dependencies || true",
      "echo 'Installing Orka VM Tools version ${var.orka_vm_tools_version}...'",
      "brew install macstadium/orka/orka-vm-tools || true",
      "echo 'Verifying Orka VM Tools installation...'",
      "brew list --versions macstadium/orka/orka-vm-tools",
      "which orka-vm-tools || echo 'orka-vm-tools command not found in PATH'",
    ]
  }

  provisioner "shell" {
    inline = [
      "# Add Homebrew to PATH in shell configuration files, use Homebrew to install Fastlane, swiftlint, Git, swift, Cocoapods, and xcodes",
      // Add or delete tools from this section as needed for your use case, XCodes will require your AppleID and password to install whichever version of XCode you specify.
      "echo >> /Users/admin/.zprofile",
      "echo 'eval \"$(/opt/homebrew/bin/brew shellenv\"' >> /Users/admin/.zprofile",
      "eval \"$(/opt/homebrew/bin/brew shellenv)\"",
      "brew install fastlane",
      "brew install git",
      "brew install cocoapods",
      "brew install xcodesorg/made/xcodes",
      "brew install swift",
    ]
  }
}
