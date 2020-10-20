package orka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepOrkaCreate struct {
	failed bool
}

func (s *stepOrkaCreate) createOrkaToken(state multistep.StateBag) (string, error) {
	config := state.Get("config").(*Config)
	user := config.OrkaUser
	password := config.OrkaPassword

	// HTTP Client.
	client := &http.Client{}

	reqData := TokenLoginRequest{user, password}
	reqDataJSON, _ := json.Marshal(reqData)
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "token"),
		bytes.NewBuffer(reqDataJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		e := fmt.Errorf("Error while logging into the Orka API")
		return "", e
	}

	var respData TokenLoginResponse
	respBodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(respBodyBytes, &respData)
	resp.Body.Close()

	return respData.Token, nil
}

func (s *stepOrkaCreate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	// ############################
	// # ORKA API LOGIN FOR TOKEN #
	// ############################

	ui.Say("Logging into Orka API endpoint.")

	token, err := s.createOrkaToken(state)

	if err != nil {
		ui.Error(fmt.Errorf("API login failed: %s", err).Error())
		state.Put("error", err)
		s.failed = true
		return multistep.ActionHalt
	}

	ui.Say("Logged in with token.")

	// Store the token in the data bag for cleanup later.
	// I am not sure how long these tokens actually last in Orka by default, but I would
	// assume as the build doesn't take hours and hours, it should still be valid by then.
	state.Put("token", token)

	// HTTP Client.
	client := &http.Client{}

	var actualImage string

	// ############################################################
	// # PRE-COPY SOURCE IMAGE TO NEW IMAGE THAT WILL GET CREATED #
	// ############################################################

	if !config.DoNotImage || !config.DoNotPrecopy {
		ui.Say(fmt.Sprintf("Pre-copying source image %s to destination image %s", config.SourceImage, config.ImageName))
		ui.Say("This can take awhile depending on how big the source image is; please wait ...")

		imageCopyRequestData := ImageCopyRequest{config.SourceImage, config.ImageName}
		imageCopyRequestDataJSON, _ := json.Marshal(imageCopyRequestData)
		imageCopyRequest, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/image/copy"),
			bytes.NewBuffer(imageCopyRequestDataJSON),
		)
		imageCopyRequest.Header.Set("Content-Type", "application/json")
		imageCopyRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		imageCopyResponse, err := client.Do(imageCopyRequest)

		if err != nil {
			e := fmt.Errorf("Error from API: %s", err)
			ui.Error(e.Error())
			state.Put("error", e)
			s.failed = true
			return multistep.ActionHalt
		}

		var imageCopyResponseData ImageCopyResponse
		imageCopyResponseBytes, _ := ioutil.ReadAll(imageCopyResponse.Body)
		json.Unmarshal(imageCopyResponseBytes, &imageCopyResponseData)
		imageCopyResponse.Body.Close()

		if imageCopyResponse.StatusCode != 200 {
			e := fmt.Errorf("Error from API: %s", imageCopyResponse.Status)
			ui.Error(e.Error())
			state.Put("error", e)
			s.failed = true
			return multistep.ActionHalt
		}

		ui.Say("Image copied.")
		actualImage = config.ImageName
	} else {
		if config.DoNotImage {
			ui.Say("Skipping source image pre-copy because of do_not_image being set.")
		} else {
			ui.Say("Skipping source image pre-copy because of do_not_precopy bieng set.")
		}
		actualImage = config.SourceImage
	}

	// #######################################
	// # CREATE THE BUILDER VM CONFIGURATION #
	// #######################################

	// Create the builder VM from a pre-existing base-image (required).

	ui.Say(fmt.Sprintf("Creating a temporary VM configuration: %s",
		config.OrkaVMBuilderName))
	ui.Say(fmt.Sprintf("Temporary VM configuration is using new, pre-copied base image: %s",
		config.ImageName))
	vmCreateConfigRequestData := VMCreateRequest{
		OrkaVMName:  config.OrkaVMBuilderName,
		OrkaVMImage: actualImage,
		OrkaImage:   config.OrkaVMBuilderName,
		OrkaCPUCore: 3,
		VCPUCount:   3,
	}
	vmCreateConfigRequestDataJSON, _ := json.Marshal(vmCreateConfigRequestData)
	vmCreateConfigRequest, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/create"),
		bytes.NewBuffer(vmCreateConfigRequestDataJSON),
	)
	vmCreateConfigRequest.Header.Set("Content-Type", "application/json")
	vmCreateConfigRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	vmCreateConfigResponse, err := client.Do(vmCreateConfigRequest)

	if err != nil {
		ui.Error(fmt.Errorf("Error while creating temporary VM configuration: %s", err).Error())
		return multistep.ActionHalt
	}

	var vmCreateConfigResponseData VMCreateResponse
	vmCreateConfigResponseBytes, _ := ioutil.ReadAll(vmCreateConfigResponse.Body)
	json.Unmarshal(vmCreateConfigResponseBytes, &vmCreateConfigResponseData)
	vmCreateConfigResponse.Body.Close()

	if vmCreateConfigResponse.StatusCode != 201 {
		state.Put("error", fmt.Errorf("Error from API while creating Orka VM: %s",
			vmCreateConfigResponse.Status).Error())
		s.failed = true
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created temporary VM configuration; message was: %s", vmCreateConfigResponseData.Message))

	// #################
	// # DEPLOY THE VM #
	// #################

	// If that succeeds, let's create a VM based on it, in order to build/pack.

	ui.Say(fmt.Sprintf("Creating temporary VM based on: %s", config.OrkaVMBuilderName))

	vmDeployRequestData := VMDeployRequest{config.OrkaVMBuilderName}
	vmDeployRequestDataJSON, _ := json.Marshal(vmDeployRequestData)
	vmDeployRequest, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/deploy"),
		bytes.NewBuffer(vmDeployRequestDataJSON),
	)
	vmDeployRequest.Header.Set("Content-Type", "application/json")
	vmDeployRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	vmDeployResponse, err := client.Do(vmDeployRequest)
	var vmDeployResponseData VMDeployResponse
	vmDeployResponseBodyBytes, _ := ioutil.ReadAll(vmDeployResponse.Body)
	json.Unmarshal(vmDeployResponseBodyBytes, &vmDeployResponseData)
	vmDeployResponse.Body.Close()

	if vmDeployResponse.StatusCode != 200 {
		state.Put(
			"error",
			fmt.Errorf("Error from API while deploying Orka VM: %s",
				vmDeployResponse.Status))
		s.failed = true
		return multistep.ActionHalt
	}

	// #########################
	// # STORE VM ID AND STATE #
	// #########################

	// Write the VM ID to our state databag for cleanup later.

	state.Put("vmid", vmDeployResponseData.VMId)

	ui.Say(fmt.Sprintf("Created VM with ID: %s",
		vmDeployResponseData.VMId))
	ui.Say(fmt.Sprintf("Server available at: %s:%s",
		vmDeployResponseData.IP, vmDeployResponseData.SSHPort))

	// Write to our state databag for pick-up by the ssh communicator.

	sshPort, _ := strconv.Atoi(vmDeployResponseData.SSHPort)

	state.Put("ssh_port", sshPort)
	state.Put("ssh_host", vmDeployResponseData.IP)

	// Continue processing
	return multistep.ActionContinue
}

func (s *stepOrkaCreate) precopyImageDelete(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	token := state.Get("token").(string)

	client := &http.Client{}

	imageDeleteRequestData := ImageDeleteRequest{config.OrkaVMBuilderName}
	imageDeleteRequestDataJSON, _ := json.Marshal(imageDeleteRequestData)
	imageDeleteRequest, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/image/delete"),
		bytes.NewBuffer(imageDeleteRequestDataJSON),
	)
	imageDeleteRequest.Header.Set("Content-Type", "application/json")
	imageDeleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	imageDeleteResponse, err := client.Do(imageDeleteRequest)

	if err != nil {
		e := fmt.Errorf("Error from API: %s", err)
		ui.Error(e.Error())
		state.Put("error", e)
	}

	if imageDeleteResponse.StatusCode != 200 {
		e := fmt.Errorf("Error from API: %s", imageDeleteResponse.Status)
		ui.Error("VM was not purged.")
		ui.Error(e.Error())
	} else {
		ui.Say("VM purged.")
	}

	imageDeleteResponse.Body.Close()
}

func (s *stepOrkaCreate) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	token := state.Get("token").(string)

	if config.DoNotDelete {
		ui.Say("We are skipping the deletion of the temporary VM and its configuration because of do_not_delete being set.")
		if !config.DoNotPrecopy {
			ui.Say(fmt.Sprintf("Pre-copy was performed; image %s will be left and not removed.",
				config.ImageName))
		}
		return
	} else if s.failed && !config.DoNotPrecopy {
		ui.Say(fmt.Sprintf("Pre-copy was performed; cleaning up image %s", config.ImageName))
		s.precopyImageDelete(state)
	} else if s.failed && config.DoNotPrecopy {
		ui.Say("There is nothing to clean up since the VM creation and deployment failed.")
		return
	}

	vmid := state.Get("vmid").(string)

	ui.Say(fmt.Sprintf("Removing temporary VM and its configuration: %s, %s", vmid, config.OrkaVMBuilderName))

	client := &http.Client{}
	vmPurgeRequestData := VMPurgeRequest{config.OrkaVMBuilderName}
	vmPurgeRequestDatJSON, _ := json.Marshal(vmPurgeRequestData)
	vmPurgeRequest, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/purge"),
		bytes.NewBuffer(vmPurgeRequestDatJSON),
	)
	vmPurgeRequest.Header.Set("Content-Type", "application/json")
	vmPurgeRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	vmPurgeResponse, err := client.Do(vmPurgeRequest)

	if err != nil {
		ui.Error(fmt.Errorf("Error while cleaning up, deleting and purging Orka VM").Error())
		state.Put("error", err)
	}

	if vmPurgeResponse.StatusCode != 200 {
		ui.Say(fmt.Sprintf("VM was not purged due to API status code: %s", vmPurgeResponse.Status))
	} else {
		ui.Say("VM purged.")
	}

	vmPurgeResponse.Body.Close()
}
