package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/progrium/rig/pkg/node"
)

type JSONProvider struct {
	Filepath string
}

func NewJSONProvider(path string) *JSONProvider {
	return &JSONProvider{
		Filepath: path,
	}
}

func (s *JSONProvider) Exists() bool {
	if _, err := os.Stat(s.Filepath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (s *JSONProvider) LoadAll() ([]node.Raw, error) {
	b, err := os.ReadFile(s.Filepath)
	if err != nil {
		return nil, err
	}
	var d []node.Raw
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *JSONProvider) SaveAll(nodes []node.Raw) error {
	for _, n := range nodes {
		// sanity check each node to find/debug unsupported types
		_, err := json.Marshal(n)
		if err != nil {
			return fmt.Errorf("%s:%s %w [%s]", n.ID, n.Component, err, n.Name)
		}
	}
	b, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}
	// TODO: write to new file and rename for safety
	if err := os.WriteFile(s.Filepath, b, 0644); err != nil {
		return err
	}
	return nil
}

func (s *JSONProvider) Save(node node.Raw) error {
	return nil // no-op?
}

func (s *JSONProvider) Close() error {
	return nil
}
