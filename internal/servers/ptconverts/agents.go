package ptconverts

import (
	"encoding/json"

	agentsv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agents/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertAgentToProto converts a models.Agent to its protobuf representation
func ConvertAgentToProto(a *models.Agent) (*agentsv1.Agent, error) {
	tags := []string{}
	_ = a.Tags.ConvertToAny(&tags)

	config := &agentsv1.AgentConfig{}
	_ = protojson.Unmarshal(a.Config, config)

	var systemInfo models.SystemInfo
	if err := a.SystemInfo.ConvertToAny(&systemInfo); err != nil {
		return nil, err
	}

	res := &agentsv1.Agent{
		Id:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		TokenHint:   a.TokenHint,
		Fingerprint: a.Fingerprint,
		Status:      ConvertAgentStatusToProto(a.Status),
		IsEnabled:   a.IsEnabled,
		Tags:        tags,
		Config:      config,
		SystemInfo:  ConvertSystemInfoToProto(systemInfo),
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}

	if a.LastSeenAt != nil {
		res.LastSeenAt = timestamppb.New(*a.LastSeenAt)
	}

	return res, nil
}

// ConvertAgentFromProto converts a protobuf representation of an agent to its models.Agent representation
func ConvertAgentFromProto(a *agentsv1.Agent) (*models.Agent, error) {
	var tags storecmn.JSONField
	tags, _ = json.Marshal(a.Tags)

	config, _ := protojson.Marshal(a.Config)

	return &models.Agent{
		ID:          a.Id,
		Name:        a.Name,
		Description: a.Description,
		TokenHint:   a.TokenHint,
		Fingerprint: a.Fingerprint,
		Status:      ConvertAgentStatusFromProto(a.Status),
		IsEnabled:   a.IsEnabled,
		Tags:        tags,
		Config:      config,
		CreatedAt:   a.CreatedAt.AsTime(),
		UpdatedAt:   a.UpdatedAt.AsTime(),
	}, nil
}

// ConvertAgentStatusFromProto converts models.AgentStatusType to its protobuf representation
func ConvertAgentStatusFromProto(status agentsv1.AgentStatus) models.AgentStatusType {
	switch status {
	case agentsv1.AgentStatus_AGENT_STATUS_ACTIVE:
		return models.AgentStatusTypeActive
	case agentsv1.AgentStatus_AGENT_STATUS_INACTIVE:
		return models.AgentStatusTypeInactive
	default:
		return models.AgentStatusTypeUnknown
	}
}

// ConvertAgentStatusToProto converts models.AgentStatusType to its protobuf representation
func ConvertAgentStatusToProto(status models.AgentStatusType) agentsv1.AgentStatus {
	switch status {
	case models.AgentStatusTypeActive:
		return agentsv1.AgentStatus_AGENT_STATUS_ACTIVE
	case models.AgentStatusTypeInactive:
		return agentsv1.AgentStatus_AGENT_STATUS_INACTIVE
	default:
		return agentsv1.AgentStatus_AGENT_STATUS_UNSPECIFIED
	}
}

// ConvertAgentConfigToProto converts models.AgentConfig to its protobuf representation
func ConvertAgentConfigToProto(cfg models.AgentConfig) (*agentsv1.AgentConfig, error) {
	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	var config agentsv1.AgentConfig
	err = protojson.Unmarshal(cfgBytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
