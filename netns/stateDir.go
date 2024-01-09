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

func StateCreate(id string) (*state.State, error) {
	s := &state.State {
		ID:     id,
		Bundle: Dir(id),
		Status: state.StatusCreating,
	}
	err := os.MkdirAll(Dir(id), 0755)
	if err != nil {
		return nil, err
	}

	_, err = os.OpenFile(filepath.Join(Dir(s.ID), state.stateFile), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	f, err := os.CreateTemp(Dir(s.ID), "state")
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
	os.Rename(f.Name(), filepath.Join(Dir(s.ID), state.stateFile))
	return s, nil
}

func Dir(id string) string {
	return filepath.Join(netnsDir, id)
}
