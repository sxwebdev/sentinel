package migrator

import (
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
)

type Service struct {
	logger         logger
	fs             embed.FS
	migrationsPath string
	opsmigrations  DataMigrations
}

func New(logger logger, fs embed.FS, migrationsPath string, opsmigrations DataMigrations) *Service {
	return &Service{
		logger:         logger,
		fs:             fs,
		migrationsPath: migrationsPath,
		opsmigrations:  opsmigrations,
	}
}

func (m *Service) info(format string, args ...any) {
	if m.logger != nil {
		m.logger.Infof(format, args...)
	}
}

func (m *Service) load() ([]migration, error) {
	// read all migration files from the embedded filesystem
	entries, err := m.fs.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded migrations: %v", err)
	}

	// clear existing migrationsMap to replace with embedded ones
	migrationsMap := map[int]migration{}

	// regex to parse filenames like: 1_init_repo.up.sql or 2_added_users.down.sql
	migRe := regexp.MustCompile(`^(\d+)_([^.]+)\.(up|down)\.sql$`)

	for _, file := range entries {
		if file.IsDir() {
			continue
		}

		// parse file name to get version, name and direction
		matches := migRe.FindStringSubmatch(file.Name())
		if matches == nil || len(matches) != 4 {
			return nil, fmt.Errorf("failed to parse migration file name %s: unexpected format", file.Name())
		}

		// extract parts
		var version int
		if _, err := fmt.Sscanf(matches[1], "%d", &version); err != nil {
			return nil, fmt.Errorf("failed to parse migration version from %s: %v", file.Name(), err)
		}
		name := matches[2]
		direction := matches[3]

		// read migration SQL from file
		fullPath := filepath.Join(m.migrationsPath, file.Name())
		data, err := m.fs.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %v", file.Name(), err)
		}

		// create migration entry if not exists
		item, exists := migrationsMap[version]
		if !exists {
			migrationsMap[version] = migration{
				name: name,
			}
		}

		switch direction {
		case "up":
			item.upSQL = string(data)
		case "down":
			item.downSQL = string(data)
		default:
			return nil, fmt.Errorf("invalid migration direction in file %s", file.Name())
		}

		migrationsMap[version] = item
	}

	// convert map to slice and sort by version
	migrationsSlice := make([]migration, 0, len(migrationsMap))
	for version, item := range migrationsMap {
		migrationsSlice = append(migrationsSlice, migration{
			version: version,
			name:    item.name,
			upSQL:   item.upSQL,
			downSQL: item.downSQL,
		})
	}

	// sort migrations by version
	sort.Slice(migrationsSlice, func(i, j int) bool {
		return migrationsSlice[i].version < migrationsSlice[j].version
	})

	// merge operational migrations by version
	for _, opMig := range m.opsmigrations {
		for i, mig := range migrationsSlice {
			if mig.version == opMig.Version {
				// merge hooks
				migrationsSlice[i].beforeUpFn = opMig.BeforeUpFn
				migrationsSlice[i].afterUpFn = opMig.AfterUpFn
				migrationsSlice[i].beforeDownFn = opMig.BeforeDownFn
				migrationsSlice[i].afterDownFn = opMig.AfterDownFn
				break
			}
		}
	}

	return migrationsSlice, nil
}
