package game

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Map struct {
	Name      string
	WorldPath string
	configRaw []byte
}

func (m *Map) CopyWorldTo(path string) error {
	return copyDir(m.WorldPath, path)
}

func (m *Map) UnmarshalConfig(v any) error {
	return yaml.Unmarshal(m.configRaw, v)
}

func loadMaps(dir string) ([]*Map, error) {
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	maps := make([]*Map, 0, len(dirs))
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		worldPath := filepath.Join(dir, d.Name(), "world")
		stat, err := os.Stat(worldPath)
		if err != nil || !stat.IsDir() {
			continue
		}
		configRaw, err := os.ReadFile(filepath.Join(dir, d.Name(), "config.yml"))
		if err != nil {
			return nil, err
		}
		maps = append(maps, &Map{
			Name:      d.Name(),
			WorldPath: worldPath,
			configRaw: configRaw,
		})
	}

	return maps, nil
}
