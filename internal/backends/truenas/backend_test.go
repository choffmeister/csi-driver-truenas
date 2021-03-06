package truenas

import (
	"context"
	"testing"

	"github.com/choffmeister/csi-driver-truenas/internal/utils"
	"github.com/choffmeister/csi-driver-truenas/test"
	"github.com/stretchr/testify/assert"
)

func Test_TruenasBackend(t *testing.T) {
	var err error
	ctx := context.Background()

	backend := NewTruenasBackend()
	err = backend.LoadParameters(map[string]string{})
	assert.NoError(t, err)
	err = backend.LoadSecrets(storageClassSecretsFromEnv(test.LoadTestEnv()))
	assert.NoError(t, err)

	name := "csi-driver-truenas-test-" + utils.RandomString(8)
	id := ""
	t.Run("create volume", func(t *testing.T) {
		id, err = backend.CreateVolume(ctx, name, 128*1024*1024)
		assert.NoError(t, err)
	})

	t.Run("expand volume", func(t *testing.T) {
		err = backend.ExpandVolume(ctx, id, 2*128*1024*1024)
		assert.NoError(t, err)
	})

	t.Run("delete volume", func(t *testing.T) {
		err = backend.DeleteVolume(ctx, id)
		assert.NoError(t, err)
	})
}

func Test_TruenasBackend_Idempotency(t *testing.T) {
	var err error
	ctx := context.Background()

	backend := NewTruenasBackend()
	err = backend.LoadParameters(map[string]string{})
	assert.NoError(t, err)
	err = backend.LoadSecrets(storageClassSecretsFromEnv(test.LoadTestEnv()))
	assert.NoError(t, err)

	name := "csi-driver-truenas-test-" + utils.RandomString(8)
	id := ""
	t.Run("create volume", func(t *testing.T) {
		id, err = backend.CreateVolume(ctx, name, 128*1024*1024)
		assert.NoError(t, err)
		id, err = backend.CreateVolume(ctx, name, 128*1024*1024)
		assert.NoError(t, err)
	})

	t.Run("expand volume", func(t *testing.T) {
		err = backend.ExpandVolume(ctx, id, 2*128*1024*1024)
		assert.NoError(t, err)
		err = backend.ExpandVolume(ctx, id, 2*128*1024*1024)
		assert.NoError(t, err)
	})

	t.Run("delete volume", func(t *testing.T) {
		err = backend.DeleteVolume(ctx, id)
		assert.NoError(t, err)
		err = backend.DeleteVolume(ctx, id)
		assert.NoError(t, err)
	})
}

func storageClassSecretsFromEnv(env test.TestEnv) map[string]string {
	return map[string]string{
		"truenas-url":             env.TruenasUrl,
		"truenas-api-key":         env.TruenasApiKey,
		"truenas-parent-dataset":  env.TruenasParentDataset,
		"truenas-tls-skip-verify": "true",
		"iscsi-base-iqn":          env.ISCSIBaseIQN,
		"iscsi-portal-ip":         env.ISCSIPortalIP,
		"iscsi-portal-port":       env.ISCSIPortalPort,
		"iscsi-portal-id":         env.ISCSIPortalID,
		"iscsi-initiator-id":      env.ISCSIInitiatorID,
	}
}
