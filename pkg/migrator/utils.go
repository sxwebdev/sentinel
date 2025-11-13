package migrator

import (
	"fmt"
	"os"
	"strings"
)

func GetMaxVersion(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read migrations dir: %w", err)
	}

	var maxVersion int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip non-SQL files
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var version int
		_, err := fmt.Sscanf(entry.Name(), "%d_", &version)
		if err != nil {
			return 0, fmt.Errorf("failed to parse migration file name %s: %w", entry.Name(), err)
		}

		if version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion, nil
}
