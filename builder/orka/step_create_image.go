package orka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	orkav1 "github.com/macstadium/packer-plugin-macstadium-orka/orkaapi/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type stepCreateImage struct{}

const (
	imageSaveTimeout    time.Duration = 5 * time.Hour
	waitForSaveMessage  string        = "Please wait as this can take a little while..."
)

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(StateConfig).(*Config)
	ui := state.Get(StateUi).(packer.Ui)

	if config.NoCreateImage {
		ui.Say("Skipping image creation because of 'no_create_image' being set")
		return multistep.ActionContinue
	}

	var oci bool
	if _, err := reference.ParseNamed(config.ImageName); err == nil {
		oci = true
	}

	if oci {
		return imageSaveOCI(ctx, state, config)
	} else {
		return imageSaveNFS(ctx, state, config)
	}
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

func imageSaveNFS(ctx context.Context, state multistep.StateBag, config *Config) multistep.StepAction {
	ui := state.Get(StateUi).(packer.Ui)
	orkaClient := state.Get(StateOrkaClient).(OrkaClient)

	vmNamespace := config.OrkaVMBuilderNamespace
	vmName := config.OrkaVMBuilderName

	ctx, cancel := context.WithTimeout(ctx, imageSaveTimeout)
	defer cancel()

	ui.Say(fmt.Sprintf("Image creation is using VM [%s] in namespace [%s]", vmName, vmNamespace))
	ui.Say(fmt.Sprintf("Saving new image [%s]", config.ImageName))
	ui.Say(waitForSaveMessage)

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

func imageSaveOCI(ctx context.Context, state multistep.StateBag, config *Config) multistep.StepAction {
	ui := state.Get(StateUi).(packer.Ui)
	orkaClient := state.Get(StateOrkaClient).(OrkaClient)

	vmNamespace := config.OrkaVMBuilderNamespace
	vmName := config.OrkaVMBuilderName

	ctx, cancel := context.WithTimeout(ctx, imageSaveTimeout)
	defer cancel()

	vmPushAPIPath := fmt.Sprintf("/api/v1/namespaces/%s/vms/%s/push", vmNamespace, vmName)
	endpoint, err := url.JoinPath(config.OrkaEndpoint, vmPushAPIPath)
	if err != nil {
		err := fmt.Errorf("failed to generate VM Push API Endpoint: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	reqModel := orkav1.OrkaVMPushRequestModel{ImageReference: config.ImageName}
	reqJSON, err := json.Marshal(reqModel)
	if err != nil {
		err := fmt.Errorf("failed to marshal VM Push request: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqJSON))
	if err != nil {
		err := fmt.Errorf("failed to create VM push request: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.OrkaAuthToken))

	cl := &http.Client{}
	response, err := cl.Do(req)
	if err != nil {
		err := fmt.Errorf("failed to send VM push request: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer response.Body.Close()

	ui.Say(fmt.Sprintf("Image push is using VM [%s] in namespace [%s]", vmName, vmNamespace))
	ui.Say(fmt.Sprintf("Pushing new image to registry [%s]", config.ImageName))

	body, err := io.ReadAll(response.Body)
	if err != nil {
		err := fmt.Errorf("failed to read request response body: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if response.StatusCode != http.StatusOK {
		var obj metav1.Status
		if err := json.Unmarshal(body, &obj); err != nil {
			err := fmt.Errorf("failed to unmarshal VM push response error. Failed with status code %d: %w", response.StatusCode, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		err := fmt.Errorf("VM push failed with error code %d: %w", response.StatusCode, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var r orkav1.OrkaVMPushResponseModel
	if err := json.Unmarshal(body, &r); err != nil {
		err := fmt.Errorf("failed to unmarshal VM push response: %w", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("image [%s] push began successfully.", config.ImageName))
	ui.Say(waitForSaveMessage)

	err = orkaClient.WaitForPush(ctx, config.OrkaVMBuilderNamespace, r.JobName)
	if err != nil {
		ui.Error(fmt.Sprintf("image [%s] push failed: %s", config.ImageName, err))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("image [%s] push finshed successfully.", config.ImageName))

	return multistep.ActionContinue
}
