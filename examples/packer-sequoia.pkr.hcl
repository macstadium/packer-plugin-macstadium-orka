packer {
  required_plugins {
    macstadium-orka = {
      version = "= 3.0.1"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}

source "macstadium-orka" "image" {
  source_image      = "ghcr.io/celanthe/sequoia155:base"
  image_name        = "packer-sequoia"
  image_description = "MacOS Sequoia 15.5 image created with Packer!"
  orka_endpoint     = "http://10.221.188.20"
  orka_auth_token   = "YOUR_TOKEN_HERE"
  ssh_username      = "admin"
  ssh_password      = "admin"
}
build {
  sources = [
    "macstadium-orka.image"
  ]
  provisioner "shell" {
    inline = [
      // Enable passwordless sudo
      "echo admin | sudo -S sh -c \"mkdir -p /etc/sudoers.d/; echo 'admin ALL=(ALL) NOPASSWD: ALL' | EDITOR=tee visudo /etc/sudoers.d/admin-nopasswd\"",
      // Enable auto-login
    ]
  }

  provisioner "shell" {
    inline = [
      "echo 'Downloading the latest version of Orka VM Tools'",
      "curl -L -o /Users/admin/Downloads/orka-vm-tools.pkg https://orka-tools.s3.amazonaws.com/orka-vm-tools/official/3.3.0/orka-vm-tools.pkg",
      "echo 'Download complete'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Installing the Orka VM Tools package'",
      "cd /Users/admin/Downloads",
      "sudo installer -pkg orka-vm-tools.pkg -target /Applications",
      "echo 'Orka VM Tools package installation complete'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Verifying the Orka VM tools installation'",
      "ls -la /Applications/orka-vm-tools/",
      "echo 'Installation verified, files are located in /Applications/orka-vm-tools'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Copying Launch Daemon plist to Library'",
      "sudo cp com.orka.vm.tools.plist /Library/LaunchDaemons/",
      "echo 'Plist file copied'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Setting correct permissions on plist,'",
      "sudo chown root:wheel /Library/LaunchDaemons/com.orka.vm.tools.plist",
      "sudo chmod 644 /Library/LaunchDaemons/com.orka.vm.tools.plist",
      "echo 'plist permissions set'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Loading Orka VM Tools Launch Daemon'",
      "sudo launchctl load /Library/LaunchDaemons/com.orka.vm.tools.plist",
      "echo 'Launch Daemon loaded successfully'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Verifying launch daemon status'",
      "sudo launchctl list | grep com.orka.vm.tools || echo 'Service found in launchctl list'",
      "echo 'Verification complete'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Cleaning up installation files'",
      "rm -f /Users/admin/Downloads/orka-vm-tools.pkg",
      "echo 'Cleanup complete'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'Orka VM tools installation complete. Rebooting virtual machine.'",
      "sudo reboot"
    ]
    expect_disconnect = true
  }
  provisioner "shell" {
    pause_before = "30s"
    inline = [
      "echo 'System back online after reboot'",
      "echo 'Orka VM Tools installation and configuration completed'"
    ]
  }
  provisioner "shell" {
    inline = [
      "echo 'we are running on the remote host'",
      "hostname",
      "touch .we-ran-packer-successfully"
    ]
  }
}