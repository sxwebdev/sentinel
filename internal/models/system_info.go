package models

import (
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
)

type SystemInfo struct {
	Version         string
	CommitHash      string
	BuildDate       string
	GoVersion       string
	SqliteVersion   string
	OS              string
	Arch            string
	Hostname        string
	KernelVersion   string
	CpuModel        string
	AvailableUpdate *AvailableUpdate
}

type AvailableUpdate struct {
	IsAvailableManual bool   `json:"is_available_manual"`
	TagName           string `json:"tag_name"`
	URL               string `json:"url"`
	Description       string `json:"description,omitempty"`
}

func GetSystemInfo(version, commitHash, buildDate string) SystemInfo {
	info := SystemInfo{
		Version:    version,
		CommitHash: commitHash,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
	}

	info.Hostname, _ = os.Hostname()

	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		info.CpuModel = cpuInfo[0].ModelName
	}

	info.KernelVersion, _ = host.KernelVersion()

	return info
}
