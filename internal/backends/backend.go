package backends

import (
	"context"
	"fmt"
	"strconv"
)

type Backend interface {
	LoadParameters(parameters map[string]string) error
	LoadSecrets(secrets map[string]string) error
	LoadPublishContext(context map[string]string) error
	CreateVolume(ctx context.Context, name string, size int64) (string, error)
	DeleteVolume(ctx context.Context, id string) error
	ExpandVolume(ctx context.Context, id string, size int64) error
	GetISCSISecrets() *ISCSISecrets
}

type ISCSISecrets struct {
	BaseIQN     string
	PortalIP    string
	PortalPort  int
	PortalId    int
	InitiatorId int
}

func LoadISCSISecrets(secrets map[string]string) (*ISCSISecrets, error) {
	baseIQN := secrets["iscsi-base-iqn"]
	if baseIQN == "" {
		return nil, fmt.Errorf("missing secret iscsi-base-iqn")
	}
	portalIP := secrets["iscsi-portal-ip"]
	if portalIP == "" {
		return nil, fmt.Errorf("missing secret iscsi-portal-ip")
	}
	portalPortStr := secrets["iscsi-portal-port"]
	portalPort := 3260
	if portalPortStr != "" {
		portalPortParsed, err := strconv.Atoi(portalPortStr)
		if err != nil {
			return nil, fmt.Errorf("malformed secret iscsi-portal-port: %w", err)
		}
		portalPort = portalPortParsed
	}
	portalIdStr := secrets["iscsi-portal-id"]
	if portalIdStr == "" {
		return nil, fmt.Errorf("missing secret iscsi-portal-id")
	}
	portalId, err := strconv.Atoi(portalIdStr)
	if err != nil {
		return nil, fmt.Errorf("malformed secret iscsi-portal-id: %w", err)
	}
	initiatorIdStr := secrets["iscsi-initiator-id"]
	if initiatorIdStr == "" {
		return nil, fmt.Errorf("missing secret iscsi-initiator-id")
	}
	initiatorId, err := strconv.Atoi(initiatorIdStr)
	if err != nil {
		return nil, fmt.Errorf("malformed secret iscsi-initiator-id: %w", err)
	}

	return &ISCSISecrets{
		BaseIQN:     baseIQN,
		PortalIP:    portalIP,
		PortalPort:  portalPort,
		PortalId:    portalId,
		InitiatorId: initiatorId,
	}, nil
}
