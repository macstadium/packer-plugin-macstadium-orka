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

type TokenLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenLoginResponse struct {
	Message string               `json:"message"`
	Token   string               `json:"token"`
	Errors  []OrkaResponseErrors `json:"errors"`
}

type OrkaResponseErrors struct {
	Message string `json:"message"`
}

type VirtualMachineCreateRequest struct {
	OrkaVMName  string `json:"orka_vm_name"`
	OrkaVMImage string `json:"orka_base_image"`
	OrkaImage   string `json:"orka_image"`
	OrkaCPUCore int    `json:"orka_cpu_core"`
	VCPUCount   int    `json:"vcpu_count"`
}

type VirtualMachineCreateResponse struct {
	Message string               `json:"message"`
	Errors  []OrkaResponseErrors `json:"errors"`
}

type VirtualMachineDeployRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type VirtualMachineDeployResponse struct {
	VMId    string `json:"vm_id"`
	IP      string `json:"ip"`
	SSHPort string `json:"ssh_port"`
}

type VirtualMachineDeleteRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type stepOrkaCreate struct {
	failed bool
}

func orkaToken(endpoint string, user string, password string) (string, error) {
	client := &http.Client{}

	reqData := TokenLoginRequest{user, password}
	reqDataJSON, _ := json.Marshal(reqData)
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", endpoint, "token"),
		bytes.NewBuffer(reqDataJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		e := fmt.Errorf("Error while logging into the Orka API")
		return "", e
	}

	defer resp.Body.Close()

	var orkaToken TokenLoginResponse
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &orkaToken)

	return orkaToken.Token, nil
}

// Run
func (s *stepOrkaCreate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Logging into Orka API endpoint.")

	token, err := orkaToken(config.OrkaEndpoint, config.OrkaUser, config.OrkaPassword)

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
	state.Put("orkaToken", token)

	// HTTP Client.
	client := &http.Client{}

	// Create the builder VM from a pre-existing base-image (required).
	ui.Say(fmt.Sprintf("Creating a temporary VM configuration for packer: %s", config.OrkaVMBuilderName))
	ui.Say(fmt.Sprintf("Temporary VM configuration is using base image: %s", config.SourceImage))

	createReqData := VirtualMachineCreateRequest{
		OrkaVMName:  config.OrkaVMBuilderName,
		OrkaVMImage: config.SourceImage,
		OrkaImage:   config.OrkaVMBuilderName,
		OrkaCPUCore: 3,
		VCPUCount:   3,
	}
	createReqDataJSON, _ := json.Marshal(createReqData)
	createReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/create"),
		bytes.NewBuffer(createReqDataJSON),
	)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	createResp, err := client.Do(createReq)

	if err != nil {
		state.Put("error", err)
		s.failed = true
		return multistep.ActionHalt
	}

	defer createResp.Body.Close()

	var createdOrkaVM VirtualMachineCreateResponse
	createRespBodyBytes, _ := ioutil.ReadAll(createResp.Body)
	json.Unmarshal(createRespBodyBytes, &createdOrkaVM)

	if createResp.StatusCode != 201 {
		state.Put("error", fmt.Errorf("Error from API while creating Orka VM: %s", createResp.Status))
		s.failed = true
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created temporary VM configuration with message: %s", createdOrkaVM.Message))

	// If that succeeds, let's create a VM based on it, in order to build/pack.
	ui.Say(fmt.Sprintf("Creating temporary VM based on: %s", config.OrkaVMBuilderName))

	deployReqData := VirtualMachineDeployRequest{config.OrkaVMBuilderName}
	deployReqDataJSON, _ := json.Marshal(deployReqData)
	deployReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/deploy"),
		bytes.NewBuffer(deployReqDataJSON),
	)
	deployReq.Header.Set("Content-Type", "application/json")
	deployReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	deployResp, err := client.Do(deployReq)

	if err != nil {
		state.Put("error", err)
		s.failed = true
		return multistep.ActionHalt
	}

	defer deployResp.Body.Close()

	var deployedOrkaVM VirtualMachineDeployResponse
	deployRespBodyBytes, _ := ioutil.ReadAll(deployResp.Body)
	json.Unmarshal(deployRespBodyBytes, &deployedOrkaVM)

	if deployResp.StatusCode != 200 {
		state.Put("error", fmt.Errorf("Error from API while deploying Orka VM: %s", deployResp.Status))
		s.failed = true
		return multistep.ActionHalt
	}

	// Write the VM ID to our state databag for cleanup later.
	state.Put("vmid", deployedOrkaVM.VMId)

	ui.Say(fmt.Sprintf("Created VM with ID: %s", deployedOrkaVM.VMId))
	ui.Say(fmt.Sprintf("Server available at: %s:%s", deployedOrkaVM.IP, deployedOrkaVM.SSHPort))

	// Write to our state databag for pick-up by the ssh communicator.
	sshPort, _ := strconv.Atoi(deployedOrkaVM.SSHPort)

	state.Put("ssh_port", sshPort)
	state.Put("ssh_host", deployedOrkaVM.IP)

	// Continue processing
	return multistep.ActionContinue
}

func (s *stepOrkaCreate) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if s.failed {
		ui.Say("There is nothing to clean up since the VM creation and deployment failed.")
		return
	}

	vmid := state.Get("vmid").(string)
	token := state.Get("orkaToken").(string)

	if config.DoNotDelete {
		ui.Say("We are skipping the deletion of the temporary VM and its configuration because of do_not_delete being set.")
		return
	}

	ui.Say(fmt.Sprintf("Removing temporary VM and its configuration: %s, %s", vmid, config.OrkaVMBuilderName))

	client := &http.Client{}
	deleteReqData := VirtualMachineDeleteRequest{config.OrkaVMBuilderName}
	deleteReqDataJSON, _ := json.Marshal(deleteReqData)
	deleteReq, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/%s", config.OrkaEndpoint, "resources/vm/purge"),
		bytes.NewBuffer(deleteReqDataJSON),
	)
	deleteReq.Header.Set("Content-Type", "application/json")
	deleteReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	deleteResp, err := client.Do(deleteReq)

	if err != nil {
		e := fmt.Errorf("Error while cleaning up and deleting Orka VM")
		state.Put("error", err)
		ui.Error(e.Error())
	}

	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != 200 {
		ui.Say("VM was not deleted due to API status code: " + deleteResp.Status)
	} else {
		ui.Say("VM deleted.")
	}
}
