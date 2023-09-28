package orka

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	orkav1 "github.com/macstadium/packer-plugin-macstadium-orka/orkaapi/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type stepCreateImage struct{}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(StateConfig).(*Config)
	ui := state.Get(StateUi).(packer.Ui)
	orkaClient := state.Get(StateOrkaClient).(OrkaClient)

	vmNamespace := config.OrkaVMBuilderNamespace
	vmName := config.OrkaVMBuilderName

	ctx, cancel := context.WithTimeout(ctx, 5*time.Hour)
	defer cancel()

	if config.NoCreateImage {
		ui.Say("Skipping image creation because of 'no_create_image' being set")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Image creation is using VM [%s] in namespace [%s]", vmName, vmNamespace))
	ui.Say(fmt.Sprintf("Saving new image [%s]", config.ImageName))
	ui.Say("Please wait as this can take a little while...")

	image := &orkav1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: DefaultOrkaNamespace,
			Name:      config.ImageName,
			Annotations: map[string]string{
				DescriptionAnnotationKey: config.ImageDescription,
			},
		},
		Spec: orkav1.ImageSpec{
			Source:          vmName,
			SourceNamespace: vmNamespace,
			SourceType:      orkav1.Vm,
			Destination:     config.ImageName,
		},
	}

	if config.ImageForceOverwrite {
		if err := client.IgnoreNotFound(orkaClient.Delete(ctx, image)); err != nil {
			err := fmt.Errorf("failed to delete existing VM image: %w", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if err := orkaClient.Create(ctx, image); err != nil {
		err := fmt.Errorf("failed to create a VM save request: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := orkaClient.WaitForImage(ctx, config.ImageName); err != nil {
		err := fmt.Errorf("failed to save the image: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("image [%s] saved successfully", config.ImageName))

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	ui := state.Get(StateUi).(packer.Ui)
	config := state.Get(StateConfig).(*Config)
	orkaClient := state.Get(StateOrkaClient).(OrkaClient)

	image := &orkav1.Image{}

	err := orkaClient.Get(context.Background(), client.ObjectKey{Namespace: DefaultOrkaNamespace, Name: config.ImageName}, image)
	if err == nil && image.Status.State == orkav1.Failed {
		ui.Say(fmt.Sprintf("Cleaning up image [%s]", config.ImageName))
		if err := orkaClient.Delete(context.Background(), image); err != nil {
			ui.Error(fmt.Sprintf("failed to delete image [%s]: %s", config.ImageName, err.Error()))
		}
	}
}
