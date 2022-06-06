package services

import (
	"fmt"

	"github.com/choffmeister/csi-driver-truenas/internal/backends"
	"github.com/choffmeister/csi-driver-truenas/internal/backends/truenas"
)

func NewBackend() (backends.Backend, error) {
	backend := truenas.NewTruenasBackend()
	return &backend, nil
}

func NewBackendForCreateVolume(parameters map[string]string, secrets map[string]string) (backends.Backend, error) {
	backend, err := NewBackend()
	if err != nil {
		return nil, err
	}
	if err := backend.LoadParameters(parameters); err != nil {
		return nil, fmt.Errorf("unable load storage class parameters: %v", err)
	}
	if err := backend.LoadSecrets(secrets); err != nil {
		return nil, fmt.Errorf("unable load storage class provisioner secrets: %v", err)
	}
	return backend, nil
}

func NewBackendForDeleteVolume(secrets map[string]string) (backends.Backend, error) {
	backend, err := NewBackend()
	if err != nil {
		return nil, err
	}
	if err := backend.LoadSecrets(secrets); err != nil {
		return nil, fmt.Errorf("unable load storage class provisioner secrets: %v", err)
	}
	return backend, nil
}

func NewBackendForControllerExpandVolume(secrets map[string]string) (backends.Backend, error) {
	backend, err := NewBackend()
	if err != nil {
		return nil, err
	}
	if err := backend.LoadSecrets(secrets); err != nil {
		return nil, fmt.Errorf("unable load storage class provisioner secrets: %v", err)
	}
	return backend, nil
}

func NewBackendForNodeExpandVolume() (backends.Backend, error) {
	return NewBackend()
}

func NewBackendForNodePublish(context map[string]string, secrets map[string]string) (backends.Backend, error) {
	backend, err := NewBackend()
	if err != nil {
		return nil, err
	}
	if err := backend.LoadPublishContext(context); err != nil {
		return nil, fmt.Errorf("unable load publish context: %v", err)
	}
	if err := backend.LoadSecrets(secrets); err != nil {
		return nil, fmt.Errorf("unable load storage class provisioner secrets: %v", err)
	}
	return backend, nil
}

func NewBackendForNodeUnpublish() (backends.Backend, error) {
	return NewBackend()
}
