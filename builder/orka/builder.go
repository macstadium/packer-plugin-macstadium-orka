package orka

import (
	"context"
	"errors"
	"net/http"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
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
}

// ConfigSpec ...
func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

// Prepare ...
func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)

	return nil, warnings, errs
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Setup the state bag and initial state for the steps.
	state := multistep.BasicStateBag{}
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var commStep, provisionStep multistep.Step
	var client HttpClient

	if b.config.Mock == (MockOptions{}) {
		client = &http.Client{}
		commStep = &communicator.StepConnect{
			Config:    &b.config.CommConfig,
			Host:      func(state multistep.StateBag) (string, error) { return state.Get(StateSshHost).(string), nil },
			SSHPort:   func(state multistep.StateBag) (int, error) { return state.Get(StateSshPort).(int), nil },
			SSHConfig: b.config.CommConfig.SSHConfigFunc(),
		}
		provisionStep = new(commonsteps.StepProvision)
	} else {
		client = &mocks.Client{ErrorType: b.config.Mock.ErrorType}
		commStep = &mocks.StepConnect{Host: b.config.CommConfig.Host()}
		provisionStep = &mocks.StepProvision{}
	}
	state.Put("client", client)

	// Create our step pipeline.
	steps := []multistep.Step{
		new(stepOrkaCreate),
		commStep,
		provisionStep,
		new(stepCreateImage),
	}

	// Run!
	runner := commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	runner.Run(ctx, &state)

	// If there was an error, return that.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If it was cancelled, then just return.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, nil
	}

	// Check if we can describe the VM.
	if state.Get("vmid").(string) == "" {
		return nil, errors.New("unable to retrieve VMID")
	}

	// No errors, must've worked.
	return &Artifact{imageId: b.config.ImageName}, nil
}
