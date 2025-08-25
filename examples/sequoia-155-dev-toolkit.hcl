packer {
  required_plugins {
    macstadium-orka = {
      version = "= 3.0.1"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}

source "macstadium-orka" "image" {
  source_image      = "ghcr.io/macstadium/orka-images/sequoia:latest" // This image has the latest version of Orka VM tools already pre-installed 
  image_name        = "packer-sequoia-155-dev-toolkit"
  image_description = "MacOS Sequoia 15.5 image created with Packer!"
  orka_endpoint     = "http://10.221.188.20"
  orka_auth_token   = "YOUR_ORKA_SERVICE_ACCOUNT_TOKEN_HERE"
  ssh_username      = "admin"
  ssh_password      = "admin"
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
