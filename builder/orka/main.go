package orka

type OrkaResponseErrors struct {
	Message string `json:"message"`
}

type TokenLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenLoginResponse struct {
	Message string               `json:"message"`
	Token   string               `json:"token"`
	Errors  []OrkaResponseErrors `json:"errors"`
}

type ImageCopyRequest struct {
	Image   string `json:"image"`
	NewName string `json:"new_name"`
}

type ImageCopyResponse struct {
	Message string               `json:"message"`
	Errors  []OrkaResponseErrors `json:"errors"`
}

type ImageDeleteRequest struct {
	Image string `json:"image"`
}

type ImageDeleteResponse struct {
	Message string `json:"message"`
}

type VMCreateRequest struct {
	OrkaVMName  string `json:"orka_vm_name"`
	OrkaVMImage string `json:"orka_base_image"`
	OrkaImage   string `json:"orka_image"`
	OrkaCPUCore int    `json:"orka_cpu_core"`
	VCPUCount   int    `json:"vcpu_count"`
}

type VMCreateResponse struct {
	Message string               `json:"message"`
	Errors  []OrkaResponseErrors `json:"errors"`
}

type VMDeployRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type VMDeployResponse struct {
	VMId    string `json:"vm_id"`
	IP      string `json:"ip"`
	SSHPort string `json:"ssh_port"`
}

type VMPurgeRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type ImageCommitRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type ImageCommitResponse struct {
	Message string `json:"message"`
}

type ImageSaveRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
	NewName    string `json:"new_name"`
}

type ImageSaveResponse struct {
	Message string `json:"message"`
}

type VMStartRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type VMStartResponse struct {
	Message string `json:"message"`
}

type VMStopRequest struct {
	OrkaVMName string `json:"orka_vm_name"`
}

type VMStopResponse struct {
	Message string `json:"message"`
}
