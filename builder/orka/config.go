//go:generate packer-sdc mapstructure-to-hcl2 -type Config,MockOptions

package orka

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const (
	defaultUserName = "admin"
	defaultPassword = "admin"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	CommConfig communicator.Config `mapstructure:",squash"`

	// Information on how to connect to the Orka API to issue a token & create VM.
	OrkaEndpoint        string `mapstructure:"orka_endpoint" required:"true"`
	OrkaUser            string `mapstructure:"orka_user" required:"true"`
	OrkaPassword        string `mapstructure:"orka_password" required:"true"`
	OrkaVMBuilderPrefix string `mapstructure:"orka_vm_builder_prefix"`
	OrkaVMBuilderName   string `mapstructure:"orka_vm_builder_name"`
	OrkaVMCPUCore       int    `mapstructure:"orka_vm_cpu_core"`

	// Name of the VM Config to launch from
	SourceImage string `mapstructure:"source_image" required:"true"`

	// The name of the resulting image. Defaults to
	// `packer-{{timestamp}}`
	// (see configuration templates for more info).
	ImageName string `mapstructure:"image_name" required:"false"`

	// Simulate create, for interanl dev/testing
	SimulateCreate bool `mapstructure:"simulate_create"`

	// Skips the pre-copy method, precopy is that the image can be comitted after the provisioners
	// have ran. This moves the copy logic up-front but also consumes more disk space
	// while the build is running. This was done as a workaround in Orka since the save
	// API method did not actually copy over the base image contents.
	ImagePrecopy bool        `mapstructure:"image_precopy" required:"false"`
	Mock         MockOptions `mapstructure:"mock" required:"false"`

	// Do not image after completion, for some manual testing, for internal dev/testing.
	NoCreateImage bool `mapstructure:"no_create_image"`

	// Do not delete after completion, for some manual testing, for internal dev/testing.
	NoDeleteVM bool `mapstructure:"no_delete_vm"`

	// Enable Boost IO Performance https://orkadocs.macstadium.com/docs/boost-io-performance
	OrkaVMBuilderEnableIOBoost *bool `mapstructure:"orka_enable_io_boost"`

	// Enable Orka IP Mapping for exposed IP networking
	EnableOrkaNodeIPMapping bool `mapstructure:"enable_orka_node_ip_mapping"`

	// Required if Enable Orka IP Mapping is enabled. Map of Node Ips to the external IP values.
	OrkaNodeIPMap map[string]string `mapstructure:"orka_node_ip_map"`
}

type MockOptions struct {
	ErrorType string `mapstructure:"error_type" required:"true"`
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

	// We always use SSH for Orka.
	c.CommConfig.Type = "ssh"

	// If we didn't specify a username, pull it from our defaults.
	if c.CommConfig.SSHUsername == "" {
		// log.Printf("No ssh username specified, using default: %s", defaultUserName)
		c.CommConfig.SSHUsername = defaultUserName
	}

	// If we didn't specify a password, pull it from our defaults.
	if c.CommConfig.SSHPassword == "" {
		// log.Printf("No ssh password specified, using default: %s", defaultPassword)
		c.CommConfig.SSHPassword = defaultPassword
	}

	// SSH should come up within' 10 seconds, but we'll give the timeout 5 minutes just incase.
	if c.CommConfig.SSHTimeout == 0 {
		c.CommConfig.SSHTimeout = 5 * time.Minute
	}

	// If our source image isn't set, this is a failure.
	if c.SourceImage == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("No source image specified! Please specify source_image in the builder options. This should be an orka_vm_name from 'orka vm configs'"))
	}

	// If our builder VM prefix wasn't given, default to packer.
	if c.OrkaVMBuilderName == "" {
		if c.OrkaVMBuilderPrefix == "" {
			name, err := interpolate.Render(fmt.Sprintf("packer-{{timestamp}}"), nil)
			if err != nil {
				return nil, err
			}
			c.OrkaVMBuilderName = name
		} else {
			name, err := interpolate.Render(fmt.Sprintf("%s-{{timestamp}}", c.OrkaVMBuilderPrefix), nil)
			if err != nil {
				return nil, err
			}
			c.OrkaVMBuilderName = name
		}
	}

	// If our image name isn't set, we'll use a default name.
	if c.ImageName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			return nil, err
		}
		// log.Printf("No Destination Image Specified. Using default: %s", name)
		c.ImageName = name
	}

	// If we didn't specify the number of cores, set it to the default of 3.
	if c.OrkaVMCPUCore == 0 {
		c.OrkaVMCPUCore = 3
	}

	if c.OrkaVMBuilderEnableIOBoost == nil {
		var defaultIOBoostValue = true
		c.OrkaVMBuilderEnableIOBoost = &defaultIOBoostValue
	}

	if es := c.CommConfig.Prepare(nil); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}
