package orka

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	orkav1 "github.com/macstadium/packer-plugin-macstadium-orka/orkaapi/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type stepCreateVm struct {
	createVMFailed bool
}

func (s *stepCreateVm) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(StateConfig).(*Config)
	ui := state.Get(StateUi).(packer.Ui)
	client := state.Get(StateOrkaClient).(OrkaClient)

	ui.Say(fmt.Sprintf("Builder VM configuration will use base image [%s]", config.SourceImage))

	// #######################################
	// # CREATE THE BUILDER VM CONFIGURATION #
	// #######################################

	vmi := orkav1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: config.OrkaVMBuilderNamespace,
			Name:      config.OrkaVMBuilderName,
		},
		Spec: orkav1.VirtualMachineInstanceSpec{
			Image:       config.SourceImage,
			CPU:         config.OrkaVMCPUCore,
			Tag:         &config.OrkaVMTag,
			TagRequired: &config.OrkaVMTagRequired,
			LegacyIO:    config.OrkaLegacyIO,
			NetBoost:    config.OrkaNetBoost,
		},
	}

	ui.Say(fmt.Sprintf("Deploying a VM [%s] in namespace [%s]", config.OrkaVMBuilderName, config.OrkaVMBuilderNamespace))
	if err := client.Create(ctx, &vmi); err != nil {
		err := fmt.Errorf("failed to deploy a VM: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	sshHost, sshPort, err := client.WaitForVm(ctx, config.OrkaVMBuilderNamespace, config.OrkaVMBuilderName, config.PackerVMWaitTimeout)
	if err != nil {
		err := fmt.Errorf("failed to wait for the VM: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// #########################
	// # STORE VM ID AND STATE #
	// #########################

	// Write the VM ID to our state databag for cleanup later.

	ui.Say(fmt.Sprintf("Created VM [%s] in namespace [%s]", config.OrkaVMBuilderName, config.OrkaVMBuilderNamespace))

	if config.EnableOrkaNodeIPMapping {
		if newSshHost, ok := config.OrkaNodeIPMap[sshHost]; !ok {
			err := fmt.Errorf("VM IP [%s] is not tracked in the provided node IP map. Please provide a mapping for this VM", sshHost)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else {
			ui.Say(fmt.Sprintf("Found Internal VM IP in map [%s -> %s]", sshHost, newSshHost))
			sshHost = newSshHost
		}
	}

	ui.Say(fmt.Sprintf("SSH server will be available at [%s:%d]", sshHost, sshPort))

	// Write to our state databag for pick-up by the ssh communicator.
	state.Put(StateSshHost, sshHost)
	state.Put(StateSshPort, sshPort)

	// Continue processing
	return multistep.ActionContinue
}

func (s *stepCreateVm) Cleanup(state multistep.StateBag) {
	config := state.Get(StateConfig).(*Config)
	ui := state.Get(StateUi).(packer.Ui)
	client := state.Get(StateOrkaClient).(OrkaClient)

	if config.NoDeleteVM {
		ui.Say("We are skipping the deletion of the builder VM and its configuration because of do_not_delete being set")
		return
	}

	if s.createVMFailed {
		ui.Say("Nothing to cleanup because the builder VM creation, deployment and/or provisioning failed.")
		return
	}

	ui.Say(fmt.Sprintf("Cleaning up builder VM [%s] from namespace [%s]", config.OrkaVMBuilderName, config.OrkaVMBuilderNamespace))

	vmi := &orkav1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: config.OrkaVMBuilderNamespace,
			Name:      config.OrkaVMBuilderName,
		},
	}
	if err := client.Delete(context.Background(), vmi); err != nil {
		state.Put("error", err)
		ui.Error(fmt.Errorf("failed to delete builder VM: %w", err).Error())
	}
}
