//go:generate mapstructure-to-hcl2 -type Config

package orka

import (
	// "fmt"
	"time"
	// "log"
	"errors"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const (
	defaultUserName     = "admin"
	defaultPassword     = "admin"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	CommConfig communicator.Config `mapstructure:",squash"`
	// Name of the VM Config to launch from
	SourceImage string `mapstructure:"source_image" required:"true"`
	// The name of the resulting image. Defaults to
	// `packer-{{timestamp}}`
	// (see configuration templates for more info).
	ImageName string `mapstructure:"image_name" required:"false"`
	// Simulate create, for interanl dev/testing
	SimulateCreate bool `mapstructure:"simulate_create"`
	// Do not image after completion, for some manual testing, for internal dev/testing
	DoNotImage bool `mapstructure:"do_not_image"`
	// Do not delete after completion, for some manual testing, for internal dev/testing
	DoNotDelete bool `mapstructure:"do_not_delete"`

}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{},
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError
	
	// We always use SSH for Orka
	c.CommConfig.Type = "ssh"
	
	// If we didn't specify a username, pull it from our defaults
	if c.CommConfig.SSHUsername == "" {
		// log.Printf("No ssh username specified, using default: %s", defaultUserName)
		c.CommConfig.SSHUsername = defaultUserName
	}

	// If we didn't specify a password, pull it from our defaults
	if c.CommConfig.SSHPassword == "" {
		// log.Printf("No ssh password specified, using default: %s", defaultPassword)
		c.CommConfig.SSHPassword = defaultPassword
	}
	
	// SSH should come up within' 10 seconds, but we'll give the timeout 5 minutes just incase
	if c.CommConfig.SSHTimeout == 0 {
		c.CommConfig.SSHTimeout = 5 * time.Minute
	}

	// If our source image isn't set, this is a failure
	if c.SourceImage == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("No Source Image Specified, please specify source_image in the builder options.  This should be an orka_vm_name from 'orka vm configs'"))
	}

	// If our image name isn't set, we'll use a default name
	if c.ImageName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			return nil, err
		}
		// log.Printf("No Destination Image Specified. Using default: %s", name)
		c.ImageName = name
	}
		
	if es := c.CommConfig.Prepare(nil); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}
