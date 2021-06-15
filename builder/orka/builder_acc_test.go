package orka

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
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

var ErrorTypes = map[string]string{
	"Login":       "false",
	"Logout":      "false",
	"VMCreate":    "false",
	"VMDeploy":    "false",
	"VMPurge":     "false",
	"ImageSave":   "false",
	"ImageCopy":   "true",
	"ImageCommit": "true",
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Env(t *testing.T) {
	if os.Getenv("PACKER_ACC") != "1" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}
}

func ErrorMockHCL(Bool string, ErrorType string) string {
	return fmt.Sprintf(
		`source "macstadium-orka" "image" {
		source_image    = "90GCatalinaSSH.img"
		image_name      = "packer-{{timestamp}}"
		orka_endpoint   = "http://10.221.188.100"
		orka_user       = "user@ms.com"
		orka_password   = "password"
		image_precopy   = %s
		simulate_create = false
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
	}`, Bool, ErrorType)
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
			if assert.NotNil(t, buildCommand.ProcessState) {
				assert.Equal(t, buildCommand.ProcessState.ExitCode(), 0, "Exit code should be zero")
			}

			logs, err := os.Open(logfile)
			assert.Nil(t, err)

			defer logs.Close()

			logsBytes, err := ioutil.ReadAll(logs)
			assert.Nil(t, err)
			logsString := string(logsBytes)

			assert.Contains(t, logsString, "Created VM [05ca969973999]")
			assert.Contains(t, logsString, "Image saved")
			return nil
		},
	}
	acctest.TestPlugin(t, testSuccessCase)
}

func TestFailedOrkaBuilder(t *testing.T) {
	for ErrorType, Bool := range ErrorTypes {
		testFailCase := &acctest.PluginTestCase{
			Name: "orka_builder_error_test",
			Setup: func() error {
				return nil
			},
			Teardown: func() error {
				return nil
			},
			Template: ErrorMockHCL(Bool, ErrorType),
			Type:     "macstadium-orka",
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if assert.NotNil(t, buildCommand.ProcessState) {
					assert.NotEqual(t, buildCommand.ProcessState.ExitCode(), 0, "Exit code should not be zero")
				}

				logs, err := os.Open(logfile)
				assert.Nil(t, err)

				defer logs.Close()

				logsBytes, err := ioutil.ReadAll(logs)
				assert.Nil(t, err)
				logsString := string(logsBytes)

				assert.Contains(t, logsString, "500 Internal Server Error")
				return nil
			},
		}
		acctest.TestPlugin(t, testFailCase)
	}
}
