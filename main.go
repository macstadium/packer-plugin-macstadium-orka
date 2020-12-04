package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/lumoslabs/packer-builder-macstadium-orka/builder/orka"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(orka.Builder))
	server.Serve()
}
