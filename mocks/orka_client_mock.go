package mocks

import (
	"context"
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OrkaClient struct {
	ErrorType string
}

const (
	errorTypeCreate       = "creation failed"
	errorTypeDelete       = "deletion failed"
	errorTypeWaitForImage = "image creation error"
	errorTypeWaitForVm    = "vm deployment error"
)

var ErrorTypes = []string{
	errorTypeCreate,
	errorTypeDelete,
	errorTypeWaitForImage,
	errorTypeWaitForVm,
}

func (m OrkaClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return nil
}

func (m OrkaClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if m.ErrorType == errorTypeCreate {
		return errors.New(m.ErrorType)
	}
	return nil
}

func (m OrkaClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if m.ErrorType == errorTypeDelete {
		return errors.New(m.ErrorType)
	}

	return nil
}

func (m OrkaClient) WaitForVm(ctx context.Context, namespace, name string, timeout int) (string, int, error) {
	if m.ErrorType == errorTypeWaitForVm {
		return "", 0, errors.New(m.ErrorType)
	}

	return "1.2.3.4", 1234, nil
}

func (m OrkaClient) WaitForImage(ctx context.Context, name string) error {
	if m.ErrorType == errorTypeWaitForImage {
		return errors.New(m.ErrorType)
	}

	return nil
}
