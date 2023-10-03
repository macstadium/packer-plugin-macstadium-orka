package v1

// SetDefaultsVirtualMachineInstance sets the default values on VirtualMachineInstance
func SetDefaultsVirtualMachineInstance(vmi *VirtualMachineInstance) {
	SetDefaultsVirtualMachineInstanceStatus(vmi)
}

// SetDefaultsVirtualMachineInstanceStatus sets the default values on VirtualMachineInstanceStatus
func SetDefaultsVirtualMachineInstanceStatus(vmi *VirtualMachineInstance) {
	status := &vmi.Status
	if status.Phase == "" {
		vmi.Status = NewVirtualMachineInstanceStatus()
	}
}

// NewVirtualMachineInstanceStatus returns a new VirtualMachineInstanceStatus
func NewVirtualMachineInstanceStatus() VirtualMachineInstanceStatus {
	return VirtualMachineInstanceStatus{
		Phase: VMPending,
	}
}
