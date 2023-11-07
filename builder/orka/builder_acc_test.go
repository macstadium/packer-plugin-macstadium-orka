package orka

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/macstadium/packer-plugin-macstadium-orka/mocks"
)

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/orka/*.go  -timeout=180m
func init() {
	cmd := exec.Command("make", "-C", "../../", "rebuild")

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

//go:embed test-fixtures/template.pkr.hcl
var testBuilderHCL2Basic string

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{} = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func ErrorMockHCL(errorType string) string {
	return fmt.Sprintf(
		`source "macstadium-orka" "image" {
		orka_endpoint   = "http://10.221.188.100"
		orka_auth_token = "myauthtoken"
		source_image    = "90gbsonomassh.orkasi"
		image_name      = "my-packer-image"
		orka_vm_builder_namespace = "my-namespace"
		orka_vm_builder_name = "my-vm-name"
		no_create_image = false
		no_delete_vm    = false
		mock { error_type = "%s" }
	}

	build {
		sources = ["sources.macstadium-orka.image"]
		provisioner "shell" {
			inline = [
				"echo we are running on the remote host",
				"hostname",
				"touch .we-ran-packer-successfully"
			]
		}
	}
`, errorType)
}

func TestSuccessfulOrkaBuilder(t *testing.T) {
	testSuccessCase := &acctest.PluginTestCase{
		Name: "orka_builder_success_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderHCL2Basic,
		Type:     "macstadium-orka",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState.ExitCode() != 0 {
				return errors.New("exit code should be zero")
			}

			logsBytes, err := os.ReadFile(logfile)
			if err != nil {
				return err
			}

			logsString := string(logsBytes)

			const expectedVmCreationMessage = "Created VM [my-vm-name] in namespace [my-namespace]"
			if !strings.Contains(logsString, expectedVmCreationMessage) {
				return fmt.Errorf("does not contain the VM creation message: %q", expectedVmCreationMessage)
			}

			const expectedSshMessage = "SSH server will be available at [1.2.3.4:1234]"
			if !strings.Contains(logsString, expectedSshMessage) {
				return fmt.Errorf("does not contain the expected SSH server message: %q", expectedSshMessage)
			}

			const expectedImageSavedMessage = "image [my-packer-image] saved successfully"
			if !strings.Contains(logsString, expectedImageSavedMessage) {
				return fmt.Errorf("does not contain the image saved message: %q", expectedImageSavedMessage)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testSuccessCase)
}

func TestFailedOrkaBuilder(t *testing.T) {
	for _, errorType := range mocks.ErrorTypes {
		t.Run(errorType, func(t *testing.T) {
			testFailCase := &acctest.PluginTestCase{
				Name: fmt.Sprintf("orka_builder_error_test_%s", errorType),
				Setup: func() error {
					return nil
				},
				Teardown: func() error {
					return nil
				},
				Template: ErrorMockHCL(errorType),
				Type:     "macstadium-orka",
				Check: func(buildCommand *exec.Cmd, logfile string) error {
					if buildCommand.ProcessState.ExitCode() == 0 {
						return errors.New("exit code should not be zero")
					}

					logsBytes, err := os.ReadFile(logfile)
					if err != nil {
						return err
					}

					logsString := string(logsBytes)

					if !strings.Contains(logsString, errorType) {
						return fmt.Errorf("the log does not contain expected error %q", errorType)
					}

					return nil
				},
			}
			acctest.TestPlugin(t, testFailCase)
		})
	}
}
