package other

import (
	"mc-portable-launcher/src/config"
	"mc-portable-launcher/src/ux"
	"os"
	"path/filepath"
)

func getInstallationHash() string {
	return config.LAUNCHER_VERSION + "-" + config.VERSION_DIR_STRING
}

func OpenMarkerFileAndCheckVersion(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		ux.ShowErrorLog(err, "read version marker")
		return false
	}
	if string(content) == getInstallationHash() {
		return true
	}
	return false
}

func MarkInstallationComplete(path string) {
	markerPath := filepath.Join(path, "INSTALLATION_COMPLETE")

	f, err := os.Create(markerPath)
	if err != nil {
		ux.ShowErrorLog(err, "mark installation complete")
		return
	}

	defer f.Close()

	content := getInstallationHash()

	_, err = f.WriteString(content)
	if err != nil {
		ux.ShowErrorLog(err, "write mark installation complete")
		return
	}
}
