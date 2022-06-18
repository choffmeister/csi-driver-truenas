package truenas

import (
	"context"
	"fmt"
	"strings"

	"github.com/choffmeister/csi-driver-truenas/internal/backends"
)

var _ backends.Backend = (*TruenasBackend)(nil)

type TruenasBackend struct {
	secrets    *TruenasSecrets
	httpClient *TruenasHttpClient
}

func NewTruenasBackend() TruenasBackend {
	return TruenasBackend{}
}

type TruenasSecrets struct {
	Url           string
	ApiKey        string
	TLSSkipVerify bool
	ParentDataset string
	ISCSI         backends.ISCSISecrets
}

func (b *TruenasBackend) LoadParameters(parameters map[string]string) error {
	return nil
}

func (b *TruenasBackend) LoadSecrets(secrets map[string]string) error {
	url := secrets["truenas-url"]
	if url == "" {
		return fmt.Errorf("missing secret truenas-url")
	}
	apiKey := secrets["truenas-api-key"]
	if apiKey == "" {
		return fmt.Errorf("missing secret truenas-api-key")
	}
	tlsSkipVerify := secrets["truenas-tls-skip-verify"] == "true"
	parentDataset := secrets["truenas-parent-dataset"]
	if apiKey == "" {
		return fmt.Errorf("missing secret truenas-parent-dataset")
	}
	iscsi, err := backends.LoadISCSISecrets(secrets)
	if err != nil {
		return err
	}

	b.secrets = &TruenasSecrets{
		Url:           url,
		ApiKey:        apiKey,
		TLSSkipVerify: tlsSkipVerify,
		ParentDataset: parentDataset,
		ISCSI:         *iscsi,
	}
	b.httpClient = NewTruenasHttpClient(url, apiKey, tlsSkipVerify)

	return nil
}

func (b *TruenasBackend) LoadPublishContext(context map[string]string) error {
	return nil
}

func (b *TruenasBackend) CreateVolume(ctx context.Context, name string, size int64) (string, error) {
	datasetName := fmt.Sprintf("%s/%s", b.secrets.ParentDataset, name)
	if dataset, err := b.httpClient.PoolDatasetPost(ctx, datasetName, size); err != nil && !strings.Contains(err.Error(), "already exists") {
		return "", fmt.Errorf("unable to create dataset: %v", err)
	} else if err == nil && dataset.Id != datasetName {
		return "", fmt.Errorf("expected dataset id to equal name: got %s", dataset.Id)
	}

	targetId := 0
	existingTargets, err := b.httpClient.ISCSITargetGet(ctx, 1000)
	if err != nil {
		return "", fmt.Errorf("unable to list iscsi targets: %v", err)
	}
	for _, existingTarget := range *existingTargets {
		if existingTarget.Name == name {
			targetId = existingTarget.Id
			break
		}
	}
	if targetId == 0 {
		target, err := b.httpClient.ISCSITargetPost(ctx, name, b.secrets.ISCSI.PortalId, b.secrets.ISCSI.InitiatorId)
		if err != nil {
			return "", fmt.Errorf("unable to create iscsi target: %v", err)
		}
		targetId = target.Id
	}

	extentId := 0
	existingExtents, err := b.httpClient.ISCSIExtentGet(ctx, 1000)
	if err != nil {
		return "", fmt.Errorf("unable to list iscsi extents: %v", err)
	}
	for _, existingExtent := range *existingExtents {
		if existingExtent.Name == name {
			extentId = existingExtent.Id
			break
		}
	}
	if extentId == 0 {
		extent, err := b.httpClient.ISCSIExtentPost(ctx, name, "zvol/"+datasetName)
		if err != nil {
			return "", fmt.Errorf("unable to create iscsi extent: %v", err)
		}
		extentId = extent.Id
	}

	targetExtentId := 0
	existingTargetExtents, err := b.httpClient.ISCSITargetExtendGet(ctx, 1000)
	if err != nil {
		return "", fmt.Errorf("unable to list iscsi target extents: %v", err)
	}
	for _, existingTargetExtent := range *existingTargetExtents {
		if existingTargetExtent.Target == targetId && existingTargetExtent.Extent == extentId {
			targetExtentId = existingTargetExtent.Id
			break
		}
	}
	if targetExtentId == 0 {
		_, err := b.httpClient.ISCSITargetExtendPost(ctx, targetId, extentId)
		if err != nil {
			return "", fmt.Errorf("unable to create iscsi target extent: %v", err)
		}
	}

	return datasetName, nil
}

func (b *TruenasBackend) DeleteVolume(ctx context.Context, id string) error {
	if err := b.httpClient.PoolDatasetIdIdDelete(ctx, id, false, false); err != nil && !strings.Contains(err.Error(), "does not exist") {
		return fmt.Errorf("unable to delete dataset: %v", err)
	}
	return nil
}

func (b *TruenasBackend) ExpandVolume(ctx context.Context, id string, size int64) error {
	if _, err := b.httpClient.PoolDatasetPutVolsize(ctx, id, size); err != nil {
		return fmt.Errorf("unable to resize dataset: %v", err)
	}

	return nil
}

func (b *TruenasBackend) CommentVolume(ctx context.Context, id string, comment string) error {
	if _, err := b.httpClient.PoolDatasetPutComments(ctx, id, comment); err != nil {
		return fmt.Errorf("unable to set dataset comment: %v", err)
	}

	return nil
}

func (b *TruenasBackend) GetISCSISecrets() *backends.ISCSISecrets {
	return &b.secrets.ISCSI
}
