package orka

import (
	"context"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateImage struct {
	imageID string
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	vmid := state.Get("vmid").(string)

	ui.Say("Creating image for VM: " + vmid)
	ui.Say("Name of image being created: " + config.ImageName)
	ui.Say("Please wait, this can take 1-5 minutes...")
	
	result, err := RunCommand( 
			"orka","image","save",
			"-v",vmid,
			"-b",config.ImageName,
			"-y")
		
  if err != nil {
      ui.Error(err.Error())
      return multistep.ActionHalt
  }
	
	ui.Say(result)

	s.imageID = vmid

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	if s.imageID == "" {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say("We should maybe delete the image here...?")

	// _, err := client.ImageApi.ImageDelete(context.TODO(), s.imageID)
	// if err != nil {
	// 	ui.Error(fmt.Sprintf("error deleting image '%s' - consider deleting it manually: %s",
	// 		s.imageID, formatOpenAPIError(err)))
	// }
}
