package netns

import (
	"fmt"
	"os"
	"path/filepath"

	"go.sbk.wtf/runj/jail"
)

func NsCreateConfig(id string) (string, error) {
	config := &jail.Config {
		Name:        id,
		Root:        "/",
		VNet:        "new",
		ChildrenMax: 20,
	}
	cfg, err := jail.RenderConfig(config)
	if err != nil {
		return "", err
	}
	confPath := NsConfPath(config.Name)
	confFile, err := os.OpenFile(confPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("jail: config should not already exist: %w", err)
	}
	defer func() {
		confFile.Close()
		if err != nil {
			os.Remove(confFile.Name())
		}
	}()
	_, err = confFile.Write([]byte(cfg))
	if err != nil {
		return "", err
	}
	return confFile.Name(), nil
}

func NsConfPath(id string) string {
	return filepath.Join(NsDir(id), "jail.conf")
}
