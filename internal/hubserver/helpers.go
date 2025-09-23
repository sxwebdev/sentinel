package hubserver

import (
	commonv1 "github.com/sxwebdev/sentinel/internal/hubserver/api/sentinel/common/v1"
	servicev1 "github.com/sxwebdev/sentinel/internal/hubserver/api/sentinel/service/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// convertServiceToProto converts a models.Service to its protobuf representation
func convertServiceToProto(svc *models.Service) *servicev1.Service {
	tags := []string{}
	_ = svc.Tags.ConvertToAny(&tags)

	var config *servicev1.ServiceConfig
	_ = protojson.Unmarshal(svc.Config, config)

	return &servicev1.Service{
		Id:                     svc.ID,
		Name:                   svc.Name,
		Protocol:               convertServiceProtocolToProto(svc.Protocol),
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

func convertServiceProtocolToProto(protocol models.ServiceProtocolType) commonv1.ServiceProtocol {
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

func convertServiceStatusToProto(status models.ServiceStatus) commonv1.ServiceStatus {
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
