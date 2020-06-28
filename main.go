package main

import (
	"github.com/andrewfarley/packer-builder-macstadium-orka/builder/orka"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(orka.Builder))
	server.Serve()
}
