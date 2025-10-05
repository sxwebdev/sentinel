package ptconverts

import (
	systemv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/system/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertSystemInfoToProto converts models.SystemInfo to its protobuf representation
func ConvertSystemInfoToProto(info models.SystemInfo) *systemv1.SystemInfo {
	res := &systemv1.SystemInfo{
		Version:       info.Version,
		CommitHash:    info.CommitHash,
		BuildDate:     info.BuildDate,
		GoVersion:     info.GoVersion,
		Os:            info.OS,
		Arch:          info.Arch,
		Hostname:      info.Hostname,
		KernelVersion: info.KernelVersion,
		IpAddress:     info.IpAddress,
		CpuModel:      info.CpuModel,
	}

	if info.StartedAt != nil && !info.StartedAt.IsZero() {
		res.StartedAt = timestamppb.New(*info.StartedAt)
	}

	return res
}

// ConvertSystemInfoFromProto converts systemv1.SystemInfo to its models representation
func ConvertSystemInfoFromProto(info *systemv1.SystemInfo) models.SystemInfo {
	res := models.SystemInfo{
		Version:       info.Version,
		CommitHash:    info.CommitHash,
		BuildDate:     info.BuildDate,
		GoVersion:     info.GoVersion,
		OS:            info.Os,
		Arch:          info.Arch,
		Hostname:      info.Hostname,
		KernelVersion: info.KernelVersion,
		IpAddress:     info.IpAddress,
		CpuModel:      info.CpuModel,
	}

	if info.StartedAt != nil && info.StartedAt.IsValid() && !info.StartedAt.AsTime().IsZero() {
		t := info.StartedAt.AsTime()
		res.StartedAt = &t
	}

	return res
}
