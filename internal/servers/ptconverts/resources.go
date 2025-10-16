package ptconverts

import (
	resourcesv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/resources/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertResourceToProto converts a models.Resource to its protobuf representation
func ConvertResourceToProto(r *models.Resource) (*resourcesv1.Resource, error) {
	if r == nil {
		return new(resourcesv1.Resource), nil
	}

	return &resourcesv1.Resource{
		Id:          r.ID,
		ProjectId:   r.ProjectID,
		Name:        r.Name,
		Description: r.Description,
		Kind:        ConvertResourceKindToProto(r.Kind),
		Tags:        r.Tags,
		Payload:     ConvertResourcePayloadToProto(r.Payload),
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}, nil
}

// ConvertResourceFromProto converts a protobuf representation of a resource to its models.Resource representation
func ConvertResourceFromProto(r *resourcesv1.Resource) (*models.Resource, error) {
	return &models.Resource{
		ID:          r.Id,
		ProjectID:   r.ProjectId,
		Name:        r.Name,
		Description: r.Description,
		Kind:        ConvertResourceKindFromProto(r.Kind),
		Tags:        r.Tags,
		Payload:     ConvertResourcePayloadFromProto(r.Payload),
		CreatedAt:   r.CreatedAt.AsTime(),
		UpdatedAt:   r.UpdatedAt.AsTime(),
	}, nil
}

// ConvertResourceKindToProto converts a models.ResourceKind to its protobuf representation
func ConvertResourceKindToProto(k models.ResourceKindType) resourcesv1.ResourceKind {
	switch k {
	case models.ResourceKindTypeUnknown:
		return resourcesv1.ResourceKind_RESOURCE_KIND_UNSPECIFIED
	case models.ResourceKindTypeServer:
		return resourcesv1.ResourceKind_RESOURCE_KIND_SERVER
	case models.ResourceKindTypeService:
		return resourcesv1.ResourceKind_RESOURCE_KIND_SERVICE
	default:
		return resourcesv1.ResourceKind_RESOURCE_KIND_UNSPECIFIED
	}
}

// ConvertResourceKindFromProto converts a protobuf ResourceKind to its models.ResourceKind representation
func ConvertResourceKindFromProto(k resourcesv1.ResourceKind) models.ResourceKindType {
	switch k {
	case resourcesv1.ResourceKind_RESOURCE_KIND_UNSPECIFIED:
		return models.ResourceKindTypeUnknown
	case resourcesv1.ResourceKind_RESOURCE_KIND_SERVER:
		return models.ResourceKindTypeServer
	case resourcesv1.ResourceKind_RESOURCE_KIND_SERVICE:
		return models.ResourceKindTypeService
	default:
		return models.ResourceKindTypeUnknown
	}
}

// ConvertResourcePayloadToProto converts a models.ResourcePayload to its protobuf representation
func ConvertResourcePayloadToProto(p models.ResourcePayload) *resourcesv1.Payload {
	return &resourcesv1.Payload{}
}

// ConvertResourcePayloadFromProto converts a protobuf Payload to its models.ResourcePayload representation
func ConvertResourcePayloadFromProto(p *resourcesv1.Payload) models.ResourcePayload {
	return models.ResourcePayload{}
}
