package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/macstadium/packer-plugin-macstadium-orka/builder/orka"
	builderVersion "github.com/macstadium/packer-plugin-macstadium-orka/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(orka.Builder))
	pps.SetVersion(builderVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
