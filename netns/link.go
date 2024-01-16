package netns

import (
	"os"
	"path/filepath"

	"go.sbk.wtf/runj/state"
)

const (
	netnsLink = "netns"
)

func CreateSymlink(jailID string, netnsID string) (error) {
	err := os.Symlink(filepath.join(state.Dir(jailID), netnsLink), NsDir(netnsID))
	if err != nil {
		return err
	}
	return nil
}
