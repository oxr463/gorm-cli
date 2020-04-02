package migrate

import (
	"os"
	"os/exec"
	"plugin"
	"strings"

	"github.com/pkg/errors"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func checkPluginExists(pluginName string) bool {
	_, err := os.Stat(MigrationTargetFolder + "/.plugins/" + pluginName)
	return !os.IsNotExist(err)
}

func buildPlugin(goFileName string) (string, error) {
	pluginName := strings.Replace(goFileName, ".go", ".so", -1)

	// Skip build plugin when exists.
	if checkPluginExists(pluginName) {
		return pluginName, nil
	}

	cmd := exec.Command(
		"go",
		"build",
		"-buildmode=plugin",
		"-o",
		MigrationTargetFolder+"/.plugins/"+pluginName,
		MigrationTargetFolder+"/"+goFileName,
	)

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return pluginName, nil
}

func getMigration(pluginName string) (Migration, error) {
	plug, err := plugin.Open(MigrationTargetFolder + "/.plugins/" + pluginName)
	if err != nil {
		return nil, errors.Wrap(err, "Open plugin filed.")
	}

	symMigration, err := plug.Lookup("Migration")
	if err != nil {
		return nil, errors.Wrap(err, "Migration wrong format - missing Migration declaration.")
	}

	var migration Migration
	migration, ok := symMigration.(Migration)
	if !ok {
		return nil, errors.Wrap(err, "Unexpected type from module symbol.")
	}
	return migration, nil
}

func removePlugin(goFileName string) error {
	pluginName := strings.Replace(goFileName, ".go", ".so", -1)
	return os.Remove(MigrationTargetFolder + "/.plugins/" + pluginName)
}

// MigrationTargetFolder is the migration folder target.
var MigrationTargetFolder string
