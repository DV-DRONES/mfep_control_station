package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CollectGarbage removes JSON and YAML files
// garbage value is the boolean true.
func CollectGarbage(root string) ([]string, error) {
	var deleted []string
	var errs []error

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			errs = append(errs, walkErr)
			return nil
		}
		if entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension != ".json" && extension != ".yaml" && extension != ".yml" {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			errs = append(errs, fmt.Errorf("inspect %q: %w", path, err))
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		marked, err := hasGarbageFlag(path, extension)
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if !marked {
			return nil
		}

		if err := os.Remove(path); err != nil {
			errs = append(errs, fmt.Errorf("delete %q: %w", path, err))
			return nil
		}
		deleted = append(deleted, path)
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}

	return deleted, errors.Join(errs...)
}

func hasGarbageFlag(path, extension string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read %q: %w", path, err)
	}

	var garbage bool
	if extension == ".json" {
		var values map[string]json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return false, fmt.Errorf("parse JSON %q: %w", path, err)
		}
		value, ok := values["garbage"]
		if !ok {
			return false, nil
		}
		if err := json.Unmarshal(value, &garbage); err != nil {
			return false, fmt.Errorf("parse garbage value in %q: %w", path, err)
		}
		return garbage, nil
	}

	var values map[string]any
	if err := yaml.Unmarshal(data, &values); err != nil {
		return false, fmt.Errorf("parse YAML %q: %w", path, err)
	}
	value, ok := values["garbage"]
	if !ok {
		return false, nil
	}
	garbage, ok = value.(bool)
	if !ok {
		return false, fmt.Errorf("garbage value in %q must be a boolean", path)
	}
	return garbage, nil
}
