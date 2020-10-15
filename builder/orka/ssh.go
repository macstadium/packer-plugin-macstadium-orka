package orka

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func CommHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		// Pull this from the database from our step_orka_create.
		return state.Get("ssh_host").(string), nil
	}
}

func CommPort(port int) func(multistep.StateBag) (int, error) {
	return func(state multistep.StateBag) (int, error) {
		// Pull this from the database from our step_orka_create.
		return state.Get("ssh_port").(int), nil
	}
}
