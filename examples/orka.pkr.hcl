packer {
  required_plugins {
    macstadium-orka = {
      version = "= 3.0.1"
      source  = "github.com/macstadium/macstadium-orka"
    }
  }
}
source "macstadium-orka" "image" {
  source_image      = "tahoe-beta-3.img"
  image_name        = "tahoe.img"
  cpu_count         = 4
  memory_gb         = 8
  disk_size_gb      = 90
  image_description = "MacOS 26 Tahoe beta 3 image created with Packer!"
  orka_endpoint     = "http://10.221.188.20"
  orka_auth_token   = "var.ORKA_TOKEN_ORKL10000025"
  ssh_username      = "admin"
  ssh_password      = "admin"
  }
  os_boot_command = [
    # hello, hola, bonjour, welcome, etc.
    "<wait60s><spacebar>",
    # Language: Switch to some other language first, e.g. "Italian" and then switch back to "English" to avoid automatically getting routed to UK English.
    #
    # [1]: Select Your Language
    "<wait30s>italiano<esc>english<enter>",
    # Select Your Country or Region
    "<wait60s>united states<leftShiftOn><tab><leftShiftOff><spacebar>",
    # Transfer Your Data to This Mac
    "<wait10s><tab><tab><tab><spacebar><tab><tab><spacebar>",
    # Written and Spoken Languages
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Accessibility
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Data & Privacy
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Create a Mac Account
    "<wait10s><tab><tab><tab><tab><tab><tab>Managed with Orka<tab>admin<tab>admin<tab>admin<tab><tab><spacebar><tab><tab><spacebar>",
    # Enable Voice Over
    "<wait120s><leftAltOn><f5><leftAltOff>",
    # Sign In with Your Apple ID
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Are you sure you want to skip signing in with an Apple ID?
    "<wait10s><tab><spacebar>",
    # Terms and Conditions
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # I have read and agree to the macOS Software License Agreement
    "<wait10s><tab><spacebar>",
    # Enable Location Services
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Are you sure you don't want to use Location Services?
    "<wait10s><tab><spacebar>",
    # Select Your Time Zone
    "<wait10s><tab><tab>EDT<enter><leftShiftOn><tab><tab><leftShiftOff><spacebar>",
    # Analytics
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Screen Time
    "<wait10s><tab><tab><spacebar>",
    # Siri
    "<wait10s><tab><spacebar><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Choose Your Look
    "<wait10s><leftShiftOn><tab><leftShiftOff><spacebar>",
    # Update Mac Automatically
    "<wait10s><tab><tab><spacebar>",
    # Welcome to Mac
    "<wait30s><spacebar>",
    # Disable Voice Over
    "<leftAltOn><f5><leftAltOff>",
    # Enable Keyboard navigation
    # This is so that we can navigate the System Settings app using the keyboard
    "<wait10s><leftAltOn><spacebar><leftAltOff>Terminal<enter>",
    "<wait10s>defaults write NSGlobalDomain AppleKeyboardUIMode -int 3<enter>",
    "<wait10s><leftAltOn>q<leftAltOff>",
    # Now that the installation is done, open "System Settings"
    "<wait10s><leftAltOn><spacebar><leftAltOff>System Settings<enter>",
    # Navigate to "Sharing"
    "<wait10s><leftCtrlOn><f2><leftCtrlOff><right><right><right><down>Sharing<enter>",
    # Navigate to "Screen Sharing" and enable it
    "<wait10s><tab><tab><tab><tab><tab><spacebar>",
    # Navigate to "Remote Login" and enable it
    "<wait10s><tab><tab><tab><tab><tab><tab><tab><tab><tab><tab><tab><tab><spacebar>",
    # Quit System Settings
    "<wait10s><leftAltOn>q<leftAltOff>",
  ]
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
provisioner "shell" {
  inline = [
  "echo 'Starting sys-daemon script execution'",
  "curl -o /Users/admin/Downloads/setup-sys-daemon.sh https://github.com/macstadium/monorepo-dev/blob/master/packages/orka-vm-arm-image/guest-scripts/setup-sys-daemon.sh",
  "sudo chmod +x /Users/admin/Downloads/setup-sys-daemon.sh",
  "sudo /Users/admin/Downloads/setup-sys-daemon.sh",
  "echo 'sys-daemon script execution complete'",
    ]
}
provisioner "shell" {
    inline = [
    "echo 'Validate new sysctl is running'",
    "sudo launchctl list sysctl",
    ]
}
provisioner "shell" {
    inline = [
    "echo 'Installing Homebrew'",
    "/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\",
    ]
}
provisioner "shell" {
  inline = [
  "echo 'Installing Orka VM Tools'",
  "brew install --cask orka-vm-tools",
    ]
}
