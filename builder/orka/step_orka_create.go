package orka

import (
	"context"
	"fmt"
	"encoding/json"
	// "log"
	// "os"
	"strings"
	"strconv"
	"errors"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepOrkaCreate struct{}

// TODO: This string parsing/extraction logic seems solid, but could probably be simplified.
//       although effort here seems useless because I'm hoping MacStadium makes their Orka
//       CLI support JSON ouptut for all commands soon, so we can just centralize on JSON
func ExtractIPHost(output string) (string, int, error) {
	// log.Println("ExtractIPHost: " + output)

	// Split into fields by whitespace
	myfields := strings.Fields(output);
	// log.Println("Fields")
	// log.Println(myfields[4])

	// Predefine variables (pre-typed) for below
	ip := ""
	port := 0
	status := "invalid"

  for _, value := range myfields {
  // for index, value := range myfields {
    // log.Printf("Character %d of GoT is = %s\n", index, value)

		// If we're at our IP
		if status == "ip" {
			ip = value
			status = "invalid"
			continue
		}
		// If we're at our port
		if status == "port" {
			// Convert to int
			portinteger, err := strconv.Atoi(value)
			port = portinteger
			if err != nil {
				err := fmt.Errorf("Error converting port to int: %s", err)
				return "", 0, err
			}
			status = "invalid"
			continue
		}
		// From the tokenization, see if the next record is IP or port (for ssh)
		if value == "IP:" {
			status = "ip"
		}
		if value == "SSH:" {
			status = "port"
		}
  }

	// log.Println("extracted ip: " + ip)	
	// log.Println("extracted port: ")
	// log.Println(port)
	
	if ip == "" || port == 0 { 
		err := fmt.Errorf("Error was unable to parse out IP or Port")
		return "", 0, err
	}

	// Return pre-typed variables ready to go
	return ip, port, nil
}

// Run our VM
func (s *stepOrkaCreate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)


	source_image := config.SourceImage
	ui.Say("Creating VM based on: " + source_image + " ...")
	
	err := errors.New("test")
	results := ""
	
	// Run our command to create an VM on Orka, simulated if necessary
	if config.SimulateCreate {
		ui.Say("Simulating create...")
		results = string(` orka vm deploy

                 VM name: macos-catalina-10-15-5 Successfully deployed VM
    Node name (optional):                         Connection Info
                Replicas: 1                        IP: 10.221.188.11
            [Attach ISO]:                         Ports
         [Attached disk]:                          VNC: 6000
       VNC Console (Y/n):                          SSH: 8823
                                                   Screen Sharing: 5901`);
		err = nil
	} else {
			// ui.Say("Actually creating VM...")
			results, err = RunCommand( 
					"orka","vm","deploy",
					"-v",source_image,
					"--vnc",
					"-y")
	}

	// Check for errors
	if err != nil {
		myerr := fmt.Errorf("Error while creating VM: %s\n%s", err, results)
		state.Put("error", err)
		ui.Error(myerr.Error())
		return multistep.ActionHalt
	}
	
	// Grab our source image
	ui.Say("Looking up VMID...")
	jsonString, _ := RunCommand(
		"orka","vm","list","--json")
		
	// TODO: This whole JSON parsing below is just a mess, it's a mess in general
	//       with golang.  If anyone wants to improve this please do it.
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(jsonString), &jsonMap)
	topLevelJson := jsonMap["virtual_machine_resources"].([]interface {})
	
	isInHere := false
	wasInHere := false

	// Search for a VM Name the same as our image (how Orka works)
	// NOTE/TODO: There is a sort-of bug in the logic here below, that it will
	//            detect all VMs that were launched from the same launch config.  This
	//            potentially could lead to this plugin shutting down one of your VMs
	//            that you didn't intend.  I'm not 100% sure how to fix this, other than
	//            just recommending that you create a new launch config any time for your
	//            automation, and only use it for your automation.  This MAY be fixable when
	//            I switch over to use the Orka API directly instead of their CLI utility 
	//            assuming that their API has more information than their CLI tool does 
	//            (eg: via vmid immediately returned on vm creation response)
  for _, value := range topLevelJson {
		isInHere = false
    // log.Printf("Character %d of toplevel is = %s\n", index, value)
		myval := value.(map[string]interface{})
		// First, check if it's in here
	  for subindex, subvalue := range myval {
			if string(subindex) == "virtual_machine_name" {
				// log.Printf("Found VM Name")
				// log.Printf(subvalue.(string))
				if subvalue.(string) == source_image {
					// log.Printf("This is a valid image")
					isInHere = true
				}
			// } else {
			// 	log.Printf("Found Subindex: " + string(subindex))
			}
		}
		
		if isInHere {
		  for subindex, subvalue := range myval {
				if string(subindex) == "status" && isInHere == true {
					// log.Printf("Found VM Status")
					for _, statusvalue := range subvalue.([]interface{}) {
				    // log.Printf("Index %d of status value is = %s\n", statusindex, statusvalue)
						mystatusval := statusvalue.(map[string]interface{})
						for substatusindex, substatusvalue := range mystatusval {
							if string(substatusindex) == "virtual_machine_id" {
								vmid := string(substatusvalue.(string))
								ui.Say("Launched VM: " + vmid)
								state.Put("vmid", vmid)
								wasInHere = true
							}
					    // log.Printf("Index %d of substatus value is = %s\n", substatusindex, substatusvalue)
						}
					}
				}
			}
		}
	}
	
	// Incase we were unable to get vmid
	if wasInHere == false {
		myerr := fmt.Errorf("Error while looking up vmid, unable to find it")
		state.Put("error", err)
		ui.Error(myerr.Error())
		return multistep.ActionHalt
	}

	// Extract the IP/Port from the string
	ip, port, err := ExtractIPHost(results)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	
	ui.Say("Server Available At: " + ip + ":" + strconv.Itoa(port))

	// Write to our state databag for pick-up by the ssh communicator
	state.Put("ssh_port", port)
	state.Put("ssh_host", ip)
	
	// Continue processing
	return multistep.ActionContinue
}

func (s *stepOrkaCreate) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	
	vmid := state.Get("vmid").(string)

	if config.DoNotDelete {
		ui.Say("We are skipping to delete the VM because of do_not_delete being set")
		return
	}
	
	ui.Say("Removing old VM: " + vmid)

	results, err := RunCommand( 
			"orka","vm","delete",
			"--vm",vmid,
			"-y")
	
	if err != nil {
		myerr := fmt.Errorf("Error while destroying VM: %s\n%s", err, results)
		state.Put("error", err)
		ui.Error(myerr.Error())
	} else {
		ui.Say("Removing old VM Complete")
	}
}
