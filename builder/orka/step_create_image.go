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
	token := state.Get("orkaToken").(string)

	if config.DoNotImage {
		ui.Say("We are skipping creating an image of the VM because of do_not_image being set.")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Creating new base image for VM: %s", vmid))
	ui.Say(fmt.Sprintf("Name of image being created: %s", config.ImageName))
	ui.Say("First, we must stop and then start (restart) the VM.")

	client := &http.Client{
		Timeout: time.Minute * 5,
	}

	stopVMRequestData := VMStopRequest{vmid}
	stopVMRequestDataJSON, _ := json.Marshal(stopVMRequestData)
	vmStopRequest, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/stop"),
		bytes.NewBuffer(stopVMRequestDataJSON),
	)
	vmStopRequest.Header.Set("Content-Type", "application/json")
	vmStopRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	vmStopResponse, err := client.Do(vmStopRequest)

	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	defer vmStopResponse.Body.Close()

	var vmStopResponseData VMStopResponse
	vmStopRespBytes, _ := ioutil.ReadAll(vmStopResponse.Body)
	json.Unmarshal(vmStopRespBytes, &vmStopResponseData)

	startVMRequestData := VMStartRequest{vmid}
	startVMRequestDataJSON, _ := json.Marshal(startVMRequestData)
	vmStartRequest, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/start"),
		bytes.NewBuffer(startVMRequestDataJSON),
	)
	vmStartRequest.Header.Set("Content-Type", "application/json")
	vmStartRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	vmStartResponse, err := client.Do(vmStartRequest)

	if err != nil {
		ui.Error(fmt.Errorf("Error while starting VM %s: %s", vmid, err).Error())
		return multistep.ActionHalt
	}

	defer vmStartRequest.Body.Close()

	var vmStartResponseData VMStartResponse
	vmStartResponseBytes, _ := ioutil.ReadAll(vmStartResponse.Body)
	json.Unmarshal(vmStartResponseBytes, &vmStartResponseData)

	// Now that the VM is stopped and restarted, we can re-image it to a new base image.
	ui.Say("VM restarted; creating image.")
	ui.Say("Please wait, this can take a little while ...")

	createImageReqData := ImageSaveRequest{vmid, config.ImageName}
	createImageReqDataJSON, _ := json.Marshal(createImageReqData)
	createImageReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/image/save"),
		bytes.NewBuffer(createImageReqDataJSON),
	)
	createImageReq.Header.Set("Content-Type", "application/json")
	createImageReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	createImageResp, err := client.Do(createImageReq)

	if err != nil {
		ui.Error(fmt.Errorf("Error while creating image: %s", err).Error())
		return multistep.ActionHalt
	}

	defer createImageResp.Body.Close()

	var imageSaveResponseData ImageSaveResponse
	createImageRespBytes, _ := ioutil.ReadAll(createImageResp.Body)
	json.Unmarshal(createImageRespBytes, &imageSaveResponseData)

	if createImageResp.StatusCode != 200 {
		ui.Error(fmt.Errorf("Image was not created due to API status code: %s", createImageResp.Status).Error())
	} else {
		ui.Say(fmt.Sprintf("Image created, response was: %s", imageSaveResponseData.Message))
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
