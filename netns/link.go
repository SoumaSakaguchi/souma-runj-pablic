package netns

import (
	"fmt"
	"os"
	"path/filepath"

	"go.sbk.wtf/runj/state"
)

const (
	netnsLink = "netns"
)

func CreateSymlink(jailID string, netnsID string) error {
	err := os.Symlink(filepath.Join(state.Dir(jailID), netnsLink), NsDir(netnsID))
	if err != nil {
		return err
	}
	return nil
}

func LoadSymlink(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	var realPath string
	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		realPath, err = os.Readlink(path)
		if err != nil{
			return "", err
		}
	} else {
		return "", fmt.Errorf("Path is not symboliclink")
	}
	return filepath.Base(realPath), nil
}
