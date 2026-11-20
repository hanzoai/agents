package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// HistoryStorage defines how workflow state is persisted.
// In production, this maps to hanzoai/s3 or VFS.
type HistoryStorage interface {
	SaveState(id string, state WorkflowState) error
	LoadState(id string) (*WorkflowState, error)
}

type VFSStorage struct {
	BasePath string
}

func (v *VFSStorage) SaveState(id string, state WorkflowState) error {
	path := filepath.Join(v.BasePath, fmt.Sprintf("%s.json", id))
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (v *VFSStorage) LoadState(id string) (*WorkflowState, error) {
	path := filepath.Join(v.BasePath, fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var state WorkflowState
	err = json.Unmarshal(data, &state)
	return &state, err
}
