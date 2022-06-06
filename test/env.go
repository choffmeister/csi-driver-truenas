package test

import (
	"log"
	"os"
	"path"

	"github.com/joho/godotenv"
)

type TestEnv struct {
	TalosIP              string
	TruenasUrl           string
	TruenasApiKey        string
	TruenasParentDataset string
	ISCSIBaseIQN         string
	ISCSIPortalIP        string
	ISCSIPortalPort      string
	ISCSIPortalID        string
	ISCSIInitiatorID     string
	StorageClassSecrets  map[string]string
}

func LoadTestEnv() TestEnv {
	if cwd, err := os.Getwd(); err != nil {
		log.Panic(err)
	} else if err := godotenv.Load(path.Join(cwd, "..", "..", "test.env")); err != nil {
		log.Panic(err)
	}

	env := TestEnv{}
	env.TalosIP = os.Getenv("TALOS_IP")
	if env.TalosIP == "" {
		log.Panic("env TALOS_IP is missing")
	}

	env.TruenasUrl = os.Getenv("TRUENAS_URL")
	if env.TruenasUrl == "" {
		log.Panic("env TRUENAS_URL is missing")
	}

	env.TruenasApiKey = os.Getenv("TRUENAS_API_KEY")
	if env.TruenasApiKey == "" {
		log.Panic("env TRUENAS_API_KEY is missing")
	}

	env.TruenasParentDataset = os.Getenv("TRUENAS_PARENT_DATASET")
	if env.TruenasParentDataset == "" {
		log.Panic("env TRUENAS_PARENT_DATASET is missing")
	}

	env.ISCSIBaseIQN = os.Getenv("ISCSI_BASE_IQN")
	if env.ISCSIBaseIQN == "" {
		log.Panic("env ISCSI_BASE_IQN is missing")
	}

	env.ISCSIPortalIP = os.Getenv("ISCSI_PORTAL_IP")
	if env.ISCSIPortalIP == "" {
		log.Panic("env ISCSI_PORTAL_IP is missing")
	}

	env.ISCSIPortalPort = os.Getenv("ISCSI_PORTAL_PORT")
	if env.ISCSIPortalPort == "" {
		log.Panic("env ISCSI_PORTAL_PORT is missing")
	}

	env.ISCSIPortalID = os.Getenv("ISCSI_PORTAL_ID")
	if env.ISCSIPortalID == "" {
		log.Panic("env ISCSI_PORTAL_ID is missing")
	}

	env.ISCSIInitiatorID = os.Getenv("ISCSI_INITIATOR_ID")
	if env.ISCSIInitiatorID == "" {
		log.Panic("env ISCSI_INITIATOR_ID is missing")
	}

	env.StorageClassSecrets = map[string]string{
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

	return env
}
