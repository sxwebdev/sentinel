package models

import (
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
)

type SystemInfo struct {
	Version         string           `json:"version"`
	CommitHash      string           `json:"commit_hash"`
	BuildDate       string           `json:"build_date"`
	GoVersion       string           `json:"go_version"`
	SqliteVersion   string           `json:"sqlite_version"`
	OS              string           `json:"os"`
	Arch            string           `json:"arch"`
	Hostname        string           `json:"hostname"`
	KernelVersion   string           `json:"kernel_version"`
	CpuModel        string           `json:"cpu_model"`
	IpAddress       string           `json:"ip_address"`
	AvailableUpdate *AvailableUpdate `json:"available_update,omitempty"`
	StartedAt       *time.Time       `json:"started_at"`
}

type AvailableUpdate struct {
	IsAvailableManual bool   `json:"is_available_manual"`
	TagName           string `json:"tag_name"`
	URL               string `json:"url"`
	Description       string `json:"description,omitempty"`
}

var startedAt = time.Now()

func GetSystemInfo(version, commitHash, buildDate string) SystemInfo {
	info := SystemInfo{
		Version:    version,
		CommitHash: commitHash,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		StartedAt:  &startedAt,
	}

	info.Hostname, _ = os.Hostname()

	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		info.CpuModel = cpuInfo[0].ModelName
	}

	info.KernelVersion, _ = host.KernelVersion()

	return info
}
