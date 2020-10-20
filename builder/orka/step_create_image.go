package orka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type ImageCommitRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type ImageCommitResponse struct {
	Message string `json:"message"`
}

type ImageSaveRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
	NewName    string `json:"new_name"`
}

type ImageSaveResponse struct {
	Message string `json:"message"`
}

type VMStartRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type VMStartResponse struct {
	Message string `json:"message"`
}

type VMStopRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type VMStopResponse struct {
	Message string `json:"message"`
}

type stepCreateImage struct {
	imageID string
	failed  bool
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	vmid := state.Get("vmid").(string)
	token := state.Get("token").(string)

	if config.DoNotImage {
		ui.Say("We are skipping commit image of the VM because of do_not_image being set.")
		return multistep.ActionContinue
	}

	// HTTP Client.

	client := &http.Client{
		Timeout: time.Minute * 5,
	}

	ui.Say(fmt.Sprintf("Comitting base image for VM: %s", vmid))
	ui.Say(fmt.Sprintf("Name of image being comitted: %s", config.ImageName))
	ui.Say("We must stop and then start (restart) the VM first.")

	stopVMRequestData := VMStopRequest{vmid}
	stopVMRequestDataJSON, _ := json.Marshal(stopVMRequestData)
	vmStopRequest, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/stop"),
		bytes.NewBuffer(stopVMRequestDataJSON),
	)
	vmStopRequest.Header.Set("Content-Type", "application/json")
	vmStopRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	ui.Say("Stopping and waiting 10 seconds...")
	vmStopResponse, err := client.Do(vmStopRequest)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	var vmStopResponseData VMStopResponse
	vmStopRespBytes, _ := ioutil.ReadAll(vmStopResponse.Body)
	json.Unmarshal(vmStopRespBytes, &vmStopResponseData)
	vmStopResponse.Body.Close()
	time.Sleep(time.Second * 10)

	// startVMRequestData := VMStartRequest{vmid}
	// startVMRequestDataJSON, _ := json.Marshal(startVMRequestData)
	// vmStartRequest, err := http.NewRequest(
	// 	http.MethodPost,
	// 	fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/start"),
	// 	bytes.NewBuffer(startVMRequestDataJSON),
	// )
	// vmStartRequest.Header.Set("Content-Type", "application/json")
	// vmStartRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	// ui.Say("Starting and waiting 30 seconds...")
	// vmStartResponse, err := client.Do(vmStartRequest)
	// if err != nil {
	// 	ui.Error(fmt.Errorf("Error while starting VM %s: %s", vmid, err).Error())
	// 	return multistep.ActionHalt
	// }
	// var vmStartResponseData VMStartResponse
	// vmStartResponseBytes, _ := ioutil.ReadAll(vmStartResponse.Body)
	// json.Unmarshal(vmStartResponseBytes, &vmStartResponseData)
	// vmStartRequest.Body.Close()
	// time.Sleep(time.Second * 30)

	// Now that the VM is stopped, we can commit it.
	ui.Say("VM stopped; comitting image.")
	ui.Say("Please wait, this can take a little while ...")

	imageCommitRequestData := ImageCommitRequest{vmid}
	imageCommitRequestDataJSON, _ := json.Marshal(imageCommitRequestData)
	imageCommitRequest, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/image/commit"),
		bytes.NewBuffer(imageCommitRequestDataJSON),
	)
	imageCommitRequest.Header.Set("Content-Type", "application/json")
	imageCommitRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	imageCommitResponse, err := client.Do(imageCommitRequest)
	if err != nil {
		ui.Error(fmt.Errorf("Error while comitting image: %s", err).Error())
		return multistep.ActionHalt
	}
	var imageCommitResponseData ImageCommitResponse
	imageCommitResponseBytes, _ := ioutil.ReadAll(imageCommitResponse.Body)
	json.Unmarshal(imageCommitResponseBytes, &imageCommitResponseData)
	imageCommitResponse.Body.Close()

	if imageCommitResponse.StatusCode != 200 {
		e := fmt.Errorf("Error from API: %s", imageCommitResponse.Status)
		ui.Error(e.Error())
	} else {
		ui.Say(fmt.Sprintf("Image comitted, response was: %s", imageCommitResponseData.Message))
	}

	s.imageID = vmid

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	if s.imageID == "" {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui.Say("We should maybe delete the image here...?")

	// _, err := client.ImageApi.ImageDelete(context.TODO(), s.imageID)
	// if err != nil {
	// 	ui.Error(fmt.Sprintf("error deleting image '%s' - consider deleting it manually: %s",
	// 		s.imageID, formatOpenAPIError(err)))
	// }
}
