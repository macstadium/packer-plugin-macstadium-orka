package models

// OrkaVMPushRequestModel describes the expected JSON input data for the vm push operation.
type OrkaVMPushRequestModel struct {
	ImageReference string `json:"imageReference" binding:"required" example:"ghcr.io/organization-name/orka-images/base:latest"`
}

// OrkaVMPushResponseModel describes the JSON response data for the vm push operation.
type OrkaVMPushResponseModel struct {
	JobName string `json:"jobName"`
}
