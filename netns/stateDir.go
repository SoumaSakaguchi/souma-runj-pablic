package netns

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go.sbk.wtf/runj/state"
)

const (
	netnsDir = "/var/run/netns"
)

func StateCreate() (*state.State, error) {
	err := os.MkdirAll(netnsDir, 0755)
	if err != nil {
		return nil, err
	}
	path, err := os.MkdirTemp(netnsDir, "netns")
	if err != nil {
		return nil, err
	}
	s := &state.State {
		ID:     filepath.Base(path),
		Bundle: path,
		Status: state.StatusCreating,
	}
	_, err = os.OpenFile(filepath.Join(path, "state.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	f, err := os.CreateTemp(path, "state")
	if err != nil {
		return nil, err
	}
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(f.Name())
		}
	}()
	d, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	_, err = f.Write(d)
	if err != nil {
		return nil, err
	}
	os.Rename(f.Name(), filepath.Join(path, "state.json"))
	return s, nil
}

func NsDir(id string) string {
	return filepath.Join(netnsDir, id)
}

func Remove(id string) error {
	return os.RemoveAll(NsDir(id))
}
