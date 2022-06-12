package test

import (
	"bytes"
	"os"
	"path"
	"text/template"

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
}

func LoadTestEnv(relFile string) TestEnv {
	if cwd, err := os.Getwd(); err != nil {
		panic(err)
	} else if err := godotenv.Load(path.Join(cwd, relFile)); err != nil {
		panic(err)
	}

	env := TestEnv{}
	env.TalosIP = os.Getenv("TALOS_IP")
	if env.TalosIP == "" {
		panic("env TALOS_IP is missing")
	}

	env.TruenasUrl = os.Getenv("TRUENAS_URL")
	if env.TruenasUrl == "" {
		panic("env TRUENAS_URL is missing")
	}

	env.TruenasApiKey = os.Getenv("TRUENAS_API_KEY")
	if env.TruenasApiKey == "" {
		panic("env TRUENAS_API_KEY is missing")
	}

	env.TruenasParentDataset = os.Getenv("TRUENAS_PARENT_DATASET")
	if env.TruenasParentDataset == "" {
		panic("env TRUENAS_PARENT_DATASET is missing")
	}

	env.ISCSIBaseIQN = os.Getenv("ISCSI_BASE_IQN")
	if env.ISCSIBaseIQN == "" {
		panic("env ISCSI_BASE_IQN is missing")
	}

	env.ISCSIPortalIP = os.Getenv("ISCSI_PORTAL_IP")
	if env.ISCSIPortalIP == "" {
		panic("env ISCSI_PORTAL_IP is missing")
	}

	env.ISCSIPortalPort = os.Getenv("ISCSI_PORTAL_PORT")
	if env.ISCSIPortalPort == "" {
		panic("env ISCSI_PORTAL_PORT is missing")
	}

	env.ISCSIPortalID = os.Getenv("ISCSI_PORTAL_ID")
	if env.ISCSIPortalID == "" {
		panic("env ISCSI_PORTAL_ID is missing")
	}

	env.ISCSIInitiatorID = os.Getenv("ISCSI_INITIATOR_ID")
	if env.ISCSIInitiatorID == "" {
		panic("env ISCSI_INITIATOR_ID is missing")
	}

	return env
}

func RenderTemplateFromEnv(env TestEnv, templateStr string) string {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		panic(err)
	}
	data := map[string]interface{}{
		"Env": env,
	}
	writer := bytes.Buffer{}
	if err := tmpl.Execute(&writer, data); err != nil {
		panic(err)
	}
	return writer.String()
}
