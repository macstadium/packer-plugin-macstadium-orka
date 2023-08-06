package orka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateImage struct {
	failedCommit bool
	failedSave   bool
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	vmid := state.Get("vmid").(string)
	token := state.Get("token").(string)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

	defer cancel()

	if config.NoCreateImage {
		ui.Say("Skipping image creation because of 'no_create_image' being set")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Image creation is using VM ID [%s]", vmid))
	ui.Say(fmt.Sprintf("Image name is [%s]", config.ImageName))

	// HTTP Client.
	client := state.Get("client").(HttpClient)

	if config.ImagePrecopy {
		// If we are using the pre-copy logic, then we just re-commit the image back.

		ui.Say("Committing existing image since pre-copy is being used")
		ui.Say("Please wait as this can take a little while...")

		imageCommitRequestData := ImageCommitRequest{vmid}
		imageCommitRequestDataJSON, _ := json.Marshal(imageCommitRequestData)
		imageCommitRequest, _ := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/image/commit"),
			bytes.NewBuffer(imageCommitRequestDataJSON),
		)
		imageCommitRequest.Header.Set("Content-Type", "application/json")
		imageCommitRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		imageCommitResponse, err := client.Do(imageCommitRequest)

		if err != nil {
			s.failedCommit = true
			e := fmt.Errorf("%s [%s]", OrkaAPIRequestErrorMessage, err)
			ui.Error(e.Error())
			state.Put("error", e)
			return multistep.ActionHalt
		}

		defer imageCommitResponse.Body.Close()

		var imageCommitResponseData ImageCommitResponse
		imageCommitResponseBytes, _ := io.ReadAll(imageCommitResponse.Body)
		json.Unmarshal(imageCommitResponseBytes, &imageCommitResponseData)

		if imageCommitResponse.StatusCode != http.StatusOK {
			s.failedCommit = true
			e := fmt.Errorf("Error committing image [%s]", imageCommitResponse.Status)
			ui.Error(e.Error())
			state.Put("error", e)
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Image comitted [%s] [%s]", imageCommitResponse.Status, imageCommitResponseData.Message))
	} else {
		// By default we use the save endpoint to generate a new base image from
		// the running VM's current image.

		ui.Say(fmt.Sprintf("Saving new image [%s]", config.ImageName))
		ui.Say("Please wait as this can take a little while...")

		imageSaveRequestData := ImageSaveRequest{vmid, config.ImageName}
		imageSaveRequestDataJSON, _ := json.Marshal(imageSaveRequestData)
		imageSaveRequest, _ := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/image/save"),
			bytes.NewBuffer(imageSaveRequestDataJSON),
		)
		imageSaveRequest.Header.Set("Content-Type", "application/json")
		imageSaveRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		imageSaveResponse, err := client.Do(imageSaveRequest)

		if err != nil {
			s.failedSave = true
			e := fmt.Errorf("%s [%s]", OrkaAPIRequestErrorMessage, err)
			ui.Error(e.Error())
			state.Put("error", e)
			return multistep.ActionHalt
		}

		defer imageSaveResponse.Body.Close()

		var imageSaveResponseData ImageSaveResponse
		imageSaveResponseBytes, _ := io.ReadAll(imageSaveResponse.Body)
		json.Unmarshal(imageSaveResponseBytes, &imageSaveResponseData)

		if imageSaveResponse.StatusCode != http.StatusOK {
			s.failedSave = true
			e := fmt.Errorf("%s [%s]", OrkaAPIResponseErrorMessage, imageSaveResponse.Status)
			ui.Error(e.Error())
			ui.Error(imageSaveResponseData.Errors[0].Message)
			state.Put("error", e)
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Image saved [%s] [%s]", imageSaveResponse.Status, imageSaveResponseData.Message))
	}

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	vmid := state.Get("vmid").(string)
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if s.failedCommit || s.failedSave {
		// TODO: Automatically clean up? Make a user-flag?
		ui.Say("Commit or save failed - please check Orka to see if any artifacts were left behind")
		return
	}

	if !cancelled && !halted {
		return
	}

	if vmid == "" {
		return
	}
}
