package orka

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/macstadium/packer-plugin-macstadium-orka/mocks"
)

const BuilderId = "orka"

const (
	DescriptionAnnotationKey = "orka.macstadium.com/description"
	DefaultOrkaNamespace     = "orka-default"
)

const (
	StateConfig     = "config"
	StateUi         = "ui"
	StateSshHost    = "ssh_host"
	StateSshPort    = "ssh_port"
	StateOrkaClient = "orka_client"
)

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
	state.Put("hook", hook) // needed for the common provisioning step
	state.Put(StateConfig, &b.config)
	state.Put(StateUi, ui)

	var commStep, provisionStep multistep.Step
	var client OrkaClient

	if b.config.Mock == (MockOptions{}) {
		if c, err := GetOrkaClient(b.config.OrkaEndpoint, b.config.OrkaAuthToken); err != nil {
			return nil, fmt.Errorf("failed to create k8s client: %w", err)
		} else {
			client = c
		}
		commStep = &communicator.StepConnect{
			Config:    &b.config.CommConfig,
			Host:      func(state multistep.StateBag) (string, error) { return state.Get(StateSshHost).(string), nil },
			SSHPort:   func(state multistep.StateBag) (int, error) { return state.Get(StateSshPort).(int), nil },
			SSHConfig: b.config.CommConfig.SSHConfigFunc(),
		}
		provisionStep = &commonsteps.StepProvision{}
	} else {
		client = &mocks.OrkaClient{ErrorType: b.config.Mock.ErrorType}
		commStep = &mocks.StepConnect{Host: b.config.CommConfig.Host()}
		provisionStep = &mocks.StepProvision{}
	}
	state.Put(StateOrkaClient, client)

	steps := []multistep.Step{
		&stepCreateVm{},
		commStep,
		provisionStep,
		&stepCreateImage{},
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

	// No errors, must've worked.
	return &Artifact{imageId: b.config.ImageName}, nil
}
