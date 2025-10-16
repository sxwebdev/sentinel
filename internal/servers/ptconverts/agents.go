package ptconverts

import (
	"encoding/json"

	agentsv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agents/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertAgentToProto converts a models.Agent to its protobuf representation
func ConvertAgentToProto(a *models.Agent) (*agentsv1.Agent, error) {
	tags := []string{}
	_ = a.Tags.ConvertToAny(&tags)

	config := &agentsv1.AgentConfig{}
	_ = protojson.Unmarshal(a.Config, config)

	res := &agentsv1.Agent{
		Id:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		TokenHint:   a.TokenHint,
		Fingerprint: a.Fingerprint,
		Status:      ConvertAgentStatusToProto(a.Status),
		Kind:        ConvertAgentKindToProto(a.Kind),
		IsEnabled:   a.IsEnabled,
		Location:    a.Location,
		Tags:        tags,
		Config:      config,
		SystemInfo:  ConvertSystemInfoToProto(a.SystemInfo),
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

	res := &models.Agent{
		ID:          a.Id,
		Name:        a.Name,
		Description: a.Description,
		TokenHint:   a.TokenHint,
		Fingerprint: a.Fingerprint,
		Status:      ConvertAgentStatusFromProto(a.Status),
		Kind:        ConvertAgentKindFromProto(a.Kind),
		Location:    a.Location,
		IsEnabled:   a.IsEnabled,
		Tags:        tags,
		Config:      config,
		CreatedAt:   a.CreatedAt.AsTime(),
		UpdatedAt:   a.UpdatedAt.AsTime(),
	}

	if a.LastSeenAt != nil {
		res.LastSeenAt = utils.Pointer(a.LastSeenAt.AsTime())
	}

	return res, nil
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

// ConvertAgentKindToProto converts models.AgentKindType to its protobuf representation
func ConvertAgentKindToProto(kind models.AgentKindType) agentsv1.AgentKind {
	switch kind {
	case models.AgentKindTypeHub:
		return agentsv1.AgentKind_AGENT_KIND_HUB
	case models.AgentKindTypeExternal:
		return agentsv1.AgentKind_AGENT_KIND_EXTERNAL
	default:
		return agentsv1.AgentKind_AGENT_KIND_UNSPECIFIED
	}
}

// ConvertAgentKindFromProto converts protobuf representation of an agent kind to its models.AgentKindType representation
func ConvertAgentKindFromProto(kind agentsv1.AgentKind) models.AgentKindType {
	switch kind {
	case agentsv1.AgentKind_AGENT_KIND_HUB:
		return models.AgentKindTypeHub
	case agentsv1.AgentKind_AGENT_KIND_EXTERNAL:
		return models.AgentKindTypeExternal
	default:
		return ""
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
