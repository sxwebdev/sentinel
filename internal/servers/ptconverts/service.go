package ptconverts

import (
	servicev1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/service/v1"
	"github.com/sxwebdev/sentinel/internal/models"
)

// ConvertServiceToProto converts a models.Service to its protobuf representation
// func ConvertServiceToProto(svc *models.Service) (*servicev1.Service, error) {
// 	tags := []string{}
// 	_ = svc.Tags.ConvertToAny(&tags)

// 	var cfg monitors.Config
// 	if err := json.Unmarshal(svc.Config, &cfg); err != nil {
// 		return nil, err
// 	}

// 	config := &servicev1.ServiceConfig{}
// 	if cfg.HTTP != nil {
// 		endpoints := make([]*servicev1.HTTPConfig_EndpointConfig, 0, len(cfg.HTTP.Endpoints))
// 		for _, ep := range cfg.HTTP.Endpoints {
// 			endpoints = append(endpoints, &servicev1.HTTPConfig_EndpointConfig{
// 				Name:           ep.Name,
// 				Url:            ep.URL,
// 				Method:         ep.Method,
// 				Headers:        ep.Headers,
// 				Body:           ep.Body,
// 				ExpectedStatus: int32(ep.ExpectedStatus),
// 				JsonPath:       ep.JSONPath,
// 				Username:       ep.Username,
// 				Password:       ep.Password,
// 			})
// 		}

// 		config.HttpConfig = &servicev1.HTTPConfig{
// 			Endpoints: endpoints,
// 			Condition: cfg.HTTP.Condition,
// 		}
// 	}

// 	if cfg.TCP != nil {
// 		config.TcpConfig = &servicev1.TCPConfig{
// 			Endpoint:   cfg.TCP.Endpoint,
// 			SendData:   cfg.TCP.SendData,
// 			ExpectData: cfg.TCP.ExpectData,
// 		}
// 	}

// 	if cfg.GRPC != nil {
// 		config.GrpcConfig = &servicev1.GRPCConfig{
// 			Endpoint:    cfg.GRPC.Endpoint,
// 			CheckType:   cfg.GRPC.CheckType,
// 			ServiceName: cfg.GRPC.ServiceName,
// 			Tls:         cfg.GRPC.TLS,
// 			InsecureTls: cfg.GRPC.InsecureTLS,
// 		}
// 	}

// 	return &servicev1.Service{
// 		Id:                     svc.ID,
// 		Name:                   svc.Name,
// 		Protocol:               ConvertServiceProtocolToProto(svc.Protocol),
// 		Interval:               svc.Interval,
// 		Timeout:                svc.Timeout,
// 		Retries:                svc.Retries,
// 		Tags:                   tags,
// 		Config:                 config,
// 		IsNotificationsEnabled: svc.IsNotificationsEnabled,
// 		IsEnabled:              svc.IsEnabled,
// 		CreatedAt:              timestamppb.New(svc.CreatedAt),
// 		UpdatedAt:              timestamppb.New(svc.UpdatedAt),
// 	}, nil
// }

// ConvertServiceFromProto converts a protobuf Service to its models representation
// func ConvertServiceFromProto(svc *servicev1.Service) (*models.Service, error) {
// 	var tags storecmn.JSONField
// 	tags, _ = json.Marshal(svc.Tags)

// 	cfg := monitors.Config{}

// 	if c := svc.GetConfig(); c != nil {
// 		if httpCfg := c.GetHttpConfig(); httpCfg != nil {
// 			endpoints := make([]monitors.EndpointConfig, 0, len(httpCfg.Endpoints))
// 			for _, ep := range httpCfg.Endpoints {
// 				endpoints = append(endpoints, monitors.EndpointConfig{
// 					Name:           ep.Name,
// 					URL:            ep.Url,
// 					Method:         ep.Method,
// 					Headers:        ep.Headers,
// 					Body:           ep.Body,
// 					ExpectedStatus: int(ep.ExpectedStatus),
// 					JSONPath:       ep.JsonPath,
// 					Username:       ep.Username,
// 					Password:       ep.Password,
// 				})
// 			}

// 			cfg.HTTP = &monitors.HTTPConfig{
// 				Endpoints: endpoints,
// 				Condition: httpCfg.Condition,
// 			}
// 		}

// 		if tcpCfg := c.GetTcpConfig(); tcpCfg != nil {
// 			cfg.TCP = &monitors.TCPConfig{
// 				Endpoint:   tcpCfg.Endpoint,
// 				SendData:   tcpCfg.SendData,
// 				ExpectData: tcpCfg.ExpectData,
// 			}
// 		}

// 		if grpcCfg := c.GetGrpcConfig(); grpcCfg != nil {
// 			cfg.GRPC = &monitors.GRPCConfig{
// 				Endpoint:    grpcCfg.Endpoint,
// 				CheckType:   grpcCfg.CheckType,
// 				ServiceName: grpcCfg.ServiceName,
// 				TLS:         grpcCfg.Tls,
// 				InsecureTLS: grpcCfg.InsecureTls,
// 			}
// 		}
// 	}

// 	config, err := json.Marshal(cfg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &models.Service{
// 		ID:                     svc.Id,
// 		Name:                   svc.Name,
// 		Protocol:               ConvertServiceProtocolFromProto(svc.Protocol),
// 		Interval:               svc.Interval,
// 		Timeout:                svc.Timeout,
// 		Retries:                svc.Retries,
// 		Tags:                   tags,
// 		Config:                 config,
// 		IsNotificationsEnabled: svc.IsNotificationsEnabled,
// 		IsEnabled:              svc.IsEnabled,
// 		CreatedAt:              svc.CreatedAt.AsTime(),
// 		UpdatedAt:              svc.UpdatedAt.AsTime(),
// 	}, nil
// }

// ConvertServiceProtocolToProto converts models.ServiceProtocolType to its protobuf representation
func ConvertServiceProtocolToProto(protocol models.ServiceProtocolType) servicev1.ServiceProtocol {
	switch protocol {
	case models.ServiceProtocolTypeHTTP:
		return servicev1.ServiceProtocol_SERVICE_PROTOCOL_HTTP
	case models.ServiceProtocolTypeTCP:
		return servicev1.ServiceProtocol_SERVICE_PROTOCOL_TCP
	case models.ServiceProtocolTypeGRPC:
		return servicev1.ServiceProtocol_SERVICE_PROTOCOL_GRPC
	default:
		return servicev1.ServiceProtocol_SERVICE_PROTOCOL_UNSPECIFIED
	}
}

// ConvertServiceProtocolFromProto converts servicev1.ServiceProtocol to models.ServiceProtocolType
func ConvertServiceProtocolFromProto(protocol servicev1.ServiceProtocol) models.ServiceProtocolType {
	switch protocol {
	case servicev1.ServiceProtocol_SERVICE_PROTOCOL_HTTP:
		return models.ServiceProtocolTypeHTTP
	case servicev1.ServiceProtocol_SERVICE_PROTOCOL_TCP:
		return models.ServiceProtocolTypeTCP
	case servicev1.ServiceProtocol_SERVICE_PROTOCOL_GRPC:
		return models.ServiceProtocolTypeGRPC
	default:
		return models.ServiceProtocolTypeUnknown
	}
}

// ConvertServiceStatusToProto converts models.ServiceStatus to its protobuf representation
func ConvertServiceStatusToProto(status models.ServiceStatus) servicev1.ServiceStatus {
	switch status {
	case models.ServiceStatusUnknown:
		return servicev1.ServiceStatus_SERVICE_STATUS_UNKNOWN
	case models.ServiceStatusUp:
		return servicev1.ServiceStatus_SERVICE_STATUS_UP
	case models.ServiceStatusDown:
		return servicev1.ServiceStatus_SERVICE_STATUS_DOWN
	default:
		return servicev1.ServiceStatus_SERVICE_STATUS_UNSPECIFIED
	}
}

// ConvertServiceStatusFromProto converts servicev1.ServiceStatus to models.ServiceStatus
func ConvertServiceStatusFromProto(status servicev1.ServiceStatus) models.ServiceStatus {
	switch status {
	case servicev1.ServiceStatus_SERVICE_STATUS_UNKNOWN:
		return models.ServiceStatusUnknown
	case servicev1.ServiceStatus_SERVICE_STATUS_UP:
		return models.ServiceStatusUp
	case servicev1.ServiceStatus_SERVICE_STATUS_DOWN:
		return models.ServiceStatusDown
	default:
		return models.ServiceStatusUnknown
	}
}
