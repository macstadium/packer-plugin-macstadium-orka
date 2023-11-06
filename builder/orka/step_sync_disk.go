package orka

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepSyncDisk struct {
}

func (s *stepSyncDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	var comm packer.Communicator
	if raw, ok := state.Get("communicator").(packer.Communicator); ok {
		comm = raw
	}

	var stderr bytes.Buffer

	ui.Say("Syncing disk changes...")

	// Start the command
	cmd := packer.RemoteCmd{Command: "sync", Stderr: &stderr}
	if err := comm.Start(ctx, &cmd); err != nil {
		err := fmt.Errorf("failed to sync disk changes: %w; stdErr=%q", err, stderr.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for it to complete
	status := cmd.Wait()

	if status != 0 {
		err := fmt.Errorf("failed to sync disk changes: status code %d; stdErr=%q", status, stderr.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Continue processing
	return multistep.ActionContinue
}

func (s *stepSyncDisk) Cleanup(multistep.StateBag) {
}
