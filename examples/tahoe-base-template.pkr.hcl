packer {
  required_plugins {
    macstadium-orka = {
      version = "= 3.0.1"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}

variable "source_image" {
  default = "ghcr.io/macstadium/orka-images/tahoe:15.0"
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
  source_image      = var.source_image
  image_name        = "${var.image_name_prefix}-{{timestamp}}"
  image_description = "MacOS Tahoe base OS image created with Packer!"
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
    execute_command = "echo 'admin' | sudo -S sh -c '{{ .Vars }} {{ .Path }}'"
    inline = [
      "echo 'Installing Homebrew'",
      "echo 'Setting up temporary passwordless sudo for admin user'",
      "echo 'admin ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/packer-temp",
      "echo 'Create homebrew directory with proper permissions'",
      "mkdir -p /opt/homebrew",
      "chown -R admin:admin /opt/homebrew",
      "echo 'Installing Homebrew as admin user'",
      "# Switch to admin user and install Homebrew with full interactive mode",
      "su - admin -c '/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"'",
      "echo 'Add Homebrew to PATH in shell configuration files, install Orka VM Tools'",
      "echo >> /Users/admin/.zprofile",
      "echo 'eval \"$(/opt/homebrew/bin/brew shellenv)\"' >> /Users/admin/.zprofile",
      "eval \"$(/opt/homebrew/bin/brew shellenv)\"",
      "echo 'Installing Orka VM Tools'",
      "su - admin -c '/opt/homebrew/bin/brew install --cask orka-vm-tools'",
      "echo 'Removing temporary passwordless sudo'",
      "rm -f /etc/sudoers.d/packer-temp",
      "echo 'Homebrew and Orka VM Tools installation completed'"
    ]
  }

  provisioner "shell" {
    inline = [
      "echo 'Starting sys-daemon script execution'",
      "curl -o /Users/admin/Downloads/setup-sys-daemon.sh https://raw.githubusercontent.com/macstadium/packer-plugin-macstadium-orka/refs/heads/main/guest-scripts/setup-sys-daemon.sh",
      "echo 'admin' | sudo -S chmod +x /Users/admin/Downloads/setup-sys-daemon.sh",
      "echo 'admin' | sudo -S /Users/admin/Downloads/setup-sys-daemon.sh",
      "echo 'sys-daemon script execution complete'"
    ]
  }

  provisioner "shell" {
    inline = [
      "echo 'Validate new sysctl is running'",
      "echo 'admin' | sudo -S launchctl list sysctl"
    ]
  }

  provisioner "shell" {
    execute_command = "echo 'admin' | sudo -S sh -c '{{ .Vars }} {{ .Path }}'"
    inline = [
      "echo 'Uninstalling Homebrew to clean up the image...'",
      "echo 'Temporarily allowing sudo for cleanup:'",
      "rm -f /etc/sudoers.d/restrict-admin",
      "echo 'admin ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/cleanup-temp",
      "chmod 440 /etc/sudoers.d/cleanup-temp",
      "echo 'Running Homebrew uninstall script:'",
      "su - admin -c '/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/uninstall.sh)\"'",
      "echo 'Removing Homebrew directories manually if they still exist:'",
      "rm -rf /opt/homebrew",
      "rm -rf /usr/local/Homebrew",
      "rm -rf /usr/local/Cellar",
      "rm -rf /usr/local/Caskroom",
      "echo 'Cleaning up Homebrew references from shell configuration:'",
      "sed -i '' '/brew shellenv/d' /Users/admin/.zprofile 2>/dev/null || true",
      "sed -i '' '/opt\\/homebrew/d' /Users/admin/.zprofile 2>/dev/null || true",
      "echo 'Re-applying sudo restrictions:'",
      "rm -f /etc/sudoers.d/cleanup-temp",
      "echo 'admin ALL=(ALL) !ALL' > /etc/sudoers.d/restrict-admin",
      "chmod 440 /etc/sudoers.d/restrict-admin",
      "echo 'Homebrew uninstall completed'"
    ]
  }
}
  
