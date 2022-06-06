package services

const (
	PluginName    = "truenas.csi.choffmeister.de"
	PluginVersion = "0.1.0"

	// TODO move these to backend?
	DefaultVolumeSize int64 = 1024 * 1024 * 1024 // 1 GB
	MinVolumeSize     int64 = 1024 * 1024        // 1 MB
)
