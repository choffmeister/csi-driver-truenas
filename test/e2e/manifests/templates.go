package manifests

import _ "embed"

var (
	//go:embed "consumer.tmpl.yaml"
	ConsumerTemplate string
	//go:embed "secret.tmpl.yaml"
	SecretTemplate string
)
