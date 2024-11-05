package conf

import "strings"

// PathConfig holds all the necessary file paths
type PathConfig struct {
	ModName        string
	MediaDirectory string
	Backup         string
	IndexFile      string
}

func VariableBuilder(a int) *PathConfig {
	// Maps to store the paths and module names for each argument
	configMap := map[int][]string{
		1: {"/NAS4", "MMD"},
		2: {"/NAS2/priv", "PMV"},
		3: {"/NAS3/samba", "Fancam"},
		4: {"/NAS2/priv", "AV"},
		0: {"debug", "test"},
	}

	config, ok := configMap[a]
	if !ok {
		// logfatal
		return nil
	}

	mediaDirectory := config[0] + "/" + config[1]
	backup := "/home/efirlus/goproject/Logs/backup/backup_" + config[1] + ".bkp"
	indexFile := config[0] + "/watch/" + strings.ToLower(config[1]) + ".dpl"

	return &PathConfig{
		ModName:        config[1],
		MediaDirectory: mediaDirectory,
		Backup:         backup,
		IndexFile:      indexFile,
	}
}
