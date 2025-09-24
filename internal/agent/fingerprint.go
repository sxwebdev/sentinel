package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/v4/host"
)

func (a *Agent) getFingerprint() string {
	// first look for a fingerprint in the data directory
	if a.config.DataDir != "" {
		if fp, err := os.ReadFile(filepath.Join(a.config.AgentDataDir(), "fingerprint")); err == nil {
			return string(fp)
		}
	}

	// if no fingerprint is found, generate one
	fingerprint, err := host.HostID()
	if err != nil || fingerprint == "" {
		fingerprint = a.systemInfo.Hostname + a.systemInfo.CpuModel
	}

	// hash fingerprint
	sum := sha256.Sum256([]byte(fingerprint))
	fingerprint = hex.EncodeToString(sum[:24])

	// save fingerprint to data directory
	if a.config.DataDir != "" {
		err = os.WriteFile(filepath.Join(a.config.AgentDataDir(), "fingerprint"), []byte(fingerprint), 0o644)
		if err != nil {
			slog.Warn("Failed to save fingerprint", "err", err)
		}
	}

	return fingerprint
}
