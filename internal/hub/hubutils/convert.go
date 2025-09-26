package hubutils

import (
	"encoding/json"

	agentv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agent/v1"
	commonv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/common/v1"
	servicev1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/service/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertServiceToProto converts a models.Service to its protobuf representation
func ConvertServiceToProto(svc *models.Service) *servicev1.Service {
	tags := []string{}
	_ = svc.Tags.ConvertToAny(&tags)

	var config *servicev1.ServiceConfig
	_ = protojson.Unmarshal(svc.Config, config)

	return &servicev1.Service{
		Id:                     svc.ID,
		Name:                   svc.Name,
		Protocol:               ConvertServiceProtocolToProto(svc.Protocol),
		Interval:               svc.Interval,
		Timeout:                svc.Timeout,
		Retries:                svc.Retries,
		Tags:                   tags,
		Config:                 config,
		IsNotificationsEnabled: svc.IsNotificationsEnabled,
		IsEnabled:              svc.IsEnabled,
		CreatedAt:              timestamppb.New(svc.CreatedAt),
		UpdatedAt:              timestamppb.New(svc.UpdatedAt),
	}
}

// ConvertServiceProtocolToProto converts models.ServiceProtocolType to its protobuf representation
func ConvertServiceProtocolToProto(protocol models.ServiceProtocolType) commonv1.ServiceProtocol {
	switch protocol {
	case models.ServiceProtocolTypeHTTP:
		return commonv1.ServiceProtocol_SERVICE_PROTOCOL_HTTP
	case models.ServiceProtocolTypeTCP:
		return commonv1.ServiceProtocol_SERVICE_PROTOCOL_TCP
	case models.ServiceProtocolTypeGRPC:
		return commonv1.ServiceProtocol_SERVICE_PROTOCOL_GRPC
	default:
		return commonv1.ServiceProtocol_SERVICE_PROTOCOL_UNSPECIFIED
	}
}

// ConvertServiceProtocolFromProto converts commonv1.ServiceProtocol to models.ServiceProtocolType
func ConvertServiceProtocolFromProto(protocol commonv1.ServiceProtocol) models.ServiceProtocolType {
	switch protocol {
	case commonv1.ServiceProtocol_SERVICE_PROTOCOL_HTTP:
		return models.ServiceProtocolTypeHTTP
	case commonv1.ServiceProtocol_SERVICE_PROTOCOL_TCP:
		return models.ServiceProtocolTypeTCP
	case commonv1.ServiceProtocol_SERVICE_PROTOCOL_GRPC:
		return models.ServiceProtocolTypeGRPC
	default:
		return models.ServiceProtocolTypeUnknown
	}
}

// ConvertServiceStatusToProto converts models.ServiceStatus to its protobuf representation
func ConvertServiceStatusToProto(status models.ServiceStatus) commonv1.ServiceStatus {
	switch status {
	case models.ServiceStatusUnknown:
		return commonv1.ServiceStatus_SERVICE_STATUS_UNKNOWN
	case models.ServiceStatusUp:
		return commonv1.ServiceStatus_SERVICE_STATUS_UP
	case models.ServiceStatusDown:
		return commonv1.ServiceStatus_SERVICE_STATUS_DOWN
	default:
		return commonv1.ServiceStatus_SERVICE_STATUS_UNSPECIFIED
	}
}

// ConvertServiceStatusFromProto converts commonv1.ServiceStatus to models.ServiceStatus
func ConvertServiceStatusFromProto(status commonv1.ServiceStatus) models.ServiceStatus {
	switch status {
	case commonv1.ServiceStatus_SERVICE_STATUS_UNKNOWN:
		return models.ServiceStatusUnknown
	case commonv1.ServiceStatus_SERVICE_STATUS_UP:
		return models.ServiceStatusUp
	case commonv1.ServiceStatus_SERVICE_STATUS_DOWN:
		return models.ServiceStatusDown
	default:
		return models.ServiceStatusUnknown
	}
}

// ConvertAgentStatusFromProto converts models.AgentStatusType to its protobuf representation
func ConvertAgentStatusFromProto(status agentv1.AgentStatus) models.AgentStatusType {
	switch status {
	case agentv1.AgentStatus_AGENT_STATUS_ACTIVE:
		return models.AgentStatusTypeActive
	case agentv1.AgentStatus_AGENT_STATUS_INACTIVE:
		return models.AgentStatusTypeInactive
	default:
		return models.AgentStatusTypeUnknown
	}
}

// ConvertAgentStatusToProto converts models.AgentStatusType to its protobuf representation
func ConvertAgentStatusToProto(status models.AgentStatusType) agentv1.AgentStatus {
	switch status {
	case models.AgentStatusTypeActive:
		return agentv1.AgentStatus_AGENT_STATUS_ACTIVE
	case models.AgentStatusTypeInactive:
		return agentv1.AgentStatus_AGENT_STATUS_INACTIVE
	default:
		return agentv1.AgentStatus_AGENT_STATUS_UNSPECIFIED
	}
}

// ConvertSystemInfoToProto converts models.SystemInfo to its protobuf representation
func ConvertSystemInfoToProto(info models.SystemInfo) *commonv1.SystemInfo {
	res := &commonv1.SystemInfo{
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

// ConvertSystemInfoFromProto converts commonv1.SystemInfo to its models representation
func ConvertSystemInfoFromProto(info *commonv1.SystemInfo) models.SystemInfo {
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

// ConvertServiceFromProto converts a protobuf Service to its models representation
func ConvertServiceFromProto(svc *servicev1.Service) *models.Service {
	var tags storecmn.JSONField
	tags, _ = json.Marshal(svc.Tags)

	config, _ := protojson.Marshal(svc.Config)

	return &models.Service{
		ID:                     svc.Id,
		Name:                   svc.Name,
		Protocol:               ConvertServiceProtocolFromProto(svc.Protocol),
		Interval:               svc.Interval,
		Timeout:                svc.Timeout,
		Retries:                svc.Retries,
		Tags:                   tags,
		Config:                 config,
		IsNotificationsEnabled: svc.IsNotificationsEnabled,
		IsEnabled:              svc.IsEnabled,
		CreatedAt:              svc.CreatedAt.AsTime(),
		UpdatedAt:              svc.UpdatedAt.AsTime(),
	}
}
