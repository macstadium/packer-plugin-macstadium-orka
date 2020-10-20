package orka

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const BuilderId = "orka"

// Builder ...
type Builder struct {
	config Config
	runner multistep.Runner
}

// ConfigSpec ...
func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

// Prepare ...
func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)

	if errs != nil {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Setup the state bag and initial state for the steps.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Create our step pipeline.
	steps := []multistep.Step{
		new(stepOrkaCreate),
	}

	// Add our SSH Communicator after our steps.
	steps = append(
		steps,
		&communicator.StepConnect{
			Config:    &b.config.CommConfig,
			Host:      CommHost(b.config.CommConfig.Host()),
			SSHPort:   CommPort(b.config.CommConfig.Port()),
			SSHConfig: b.config.CommConfig.SSHConfigFunc(),
		},
	)

	// Add the typical common provisioner after that, then our create image.
	steps = append(
		steps,
		new(common.StepProvision),
		new(stepCreateImage))

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If it was cancelled, then just return.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, nil
	}

	// Check if we can describe the VM.
	vmid := state.Get("vmid").(string)

	if vmid == "" {
		err := fmt.Errorf("Unable to retrieve VMID")
		return nil, err
	}

	// No errors, must've worked.
	return &Artifact{
		imageId: b.config.ImageName,
	}, nil
}
