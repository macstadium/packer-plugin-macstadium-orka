package orka

import (
	"errors"
)

// Artifact represents an Orka image as the result of a Packer build.
type Artifact struct {
	imageId string
}

// BuilderId returns the builder Id.
func (*Artifact) BuilderId() string {
	return BuilderId
}

// Destroy destroys the image represented by the artifact.
func (a *Artifact) Destroy() error {
	return errors.New("Destroy not implemented")
}

// Files returns the files represented by the artifact.
func (*Artifact) Files() []string {
	return nil
}

// Id returns the VM UUID.
func (a *Artifact) Id() string {
	return a.imageId
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
	return a.imageId
}
