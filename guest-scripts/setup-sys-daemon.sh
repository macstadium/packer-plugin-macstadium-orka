#!/bin/bash

# This script will create Launch Daemon which will change on system load the net kernel parameters
# net.link.generic.system.hwcksum_* - Disable checksum offloading
# net.inet.tcp.tso - Disable Hardware TCP Segmentation Offload (TSO) and Hardware Large Receive Offload (LRO),
# nearly all hardware/drivers have issues with these settings, and they can lead to throughput issues.

sudo -s
cat > /Library/LaunchDaemons/sysctl.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
 <key>Label</key>
 <string>sysctl</string>
 <key>ProgramArguments</key>
 <array>
 <string>/usr/sbin/sysctl</string>
 <string>-w</string>
 <string>net.link.generic.system.hwcksum_tx=0</string>
 <string>net.link.generic.system.hwcksum_rx=0</string>
 <string>net.inet.tcp.tso=0</string>
 </array>
 <key>RunAtLoad</key>
 <true/>
</dict>
</plist>
EOF
launchctl load /Library/LaunchDaemons/sysctl.plist
