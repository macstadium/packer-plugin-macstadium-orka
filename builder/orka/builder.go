package orka

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/macstadium/packer-plugin-macstadium-orka/mocks"
)

const BuilderId = "orka"

type HttpClient interface {
    Do(req *http.Request) (*http.Response, error)
}

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

	// Check if mock block is empty
	if b.config.Mock == (MockOptions{}) {
		state.Put("client", &http.Client{})
	} else {
		ErrorType := b.config.Mock.ErrorType
		state.Put("client", &mocks.Client{ErrorType: ErrorType})
	}

	// Create our step pipeline.
	steps := []multistep.Step{
		new(stepOrkaCreate),
	}

	// Iniitialize communicatior
	var comm = &communicator.StepConnect{
			Config:    &b.config.CommConfig,
			Host:      CommHost(b.config.CommConfig.Host()),
			SSHPort:   CommPort(b.config.CommConfig.Port()),
			SSHConfig: b.config.CommConfig.SSHConfigFunc(),
	}


	// Add our SSH Communicator after our steps.
	if b.config.Mock == (MockOptions{}) {
		steps = append(
			steps,
			comm,
			new(commonsteps.StepProvision),
			new(stepCreateImage),
		)
	} else {
		MockComm := &mocks.StepConnect{
			Host: b.config.CommConfig.Host(),
		}
		steps = append(
			steps,
			MockComm,
			new(mocks.StepProvision),
			new(stepCreateImage),
		)

	}

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
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
