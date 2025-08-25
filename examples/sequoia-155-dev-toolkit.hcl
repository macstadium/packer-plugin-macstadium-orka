packer {
  required_plugins {
    macstadium-orka = {
      version = "= 3.0.1"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}
variable "source_image" {
  default = env("PKR_VAR_source_image")
}
variable "image_name" {
  default = env("PKR_VAR_image_name")
}
variable "orka_endpoint" {
  default = env("PKR_VAR_orka_endpoint")
}
variable "orka_auth_token" {
  default = env("PKR_VAR_orka_auth_token")
}
variable "ssh_username" {
  default = env("PKR_VAR_ssh_username")
}
variable "ssh_password" {
  default = env("PKR_VAR_ssh_password")
}

source "macstadium-orka" "image" {
  source_image      = var.source_image // This image has the latest version of Orka VM tools already pre-installed 
  image_name        = var.image_name
  image_description = "MacOS Sequoia 15.5 image created with Packer!"
  orka_endpoint     = var.orka_endpoint
  orka_auth_token   = var.orka_auth_token
  ssh_username      = var.ssh_username
  ssh_password      = var.ssh_password
}

}
build {
  sources = [
    "macstadium-orka.image"
  ]

  provisioner "shell" {
    inline = [
      "echo 'Installing Homebrew'",
      "/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"",
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
      "brew install swiftlint",
      "brew install swift",
    ]
  }
