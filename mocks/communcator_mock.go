package mocks 

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepConnect struct {
	Host string
}

func (s *StepConnect) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Using SSH communicator to connect: %s", s.Host))
	ui.Say("Waiting for SSH to become available...")
	ui.Say("Connected to to SSH!")
	return multistep.ActionContinue
}

func (s *StepConnect) Cleanup(state multistep.StateBag) {
}