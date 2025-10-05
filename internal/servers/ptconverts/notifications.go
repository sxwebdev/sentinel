package ptconverts

import (
	notificationsv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/notifications/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertHistoryViewToProto converts models.NotificationHistoryView to its protobuf representation
func ConvertHistoryViewToProto(h *models.NotificationHistoryView) *notificationsv1.HistoryItem {
	item := &notificationsv1.HistoryItem{
		Id:        h.NotificationHistory.ID,
		Message:   h.NotificationHistory.Message,
		Status:    h.NotificationHistory.Status,
		Attempts:  h.NotificationHistory.Attempts,
		CreatedAt: timestamppb.New(h.NotificationHistory.CreatedAt),
		UpdatedAt: timestamppb.New(h.NotificationHistory.UpdatedAt),
	}

	if h.NotificationHistory.Response != nil {
		item.Response = h.NotificationHistory.Response
	}
	if h.NotificationHistory.ErrorMessage != nil {
		item.ErrorMessage = h.NotificationHistory.ErrorMessage
	}
	if h.NotificationHistory.LastAttemptAt != nil {
		item.LastAttemptAt = timestamppb.New(*h.NotificationHistory.LastAttemptAt)
	}
	if h.NotificationHistory.SentAt != nil {
		item.SentAt = timestamppb.New(*h.NotificationHistory.SentAt)
	}
	return item
}
