package dispatcher

type EventType string

const (
	EventTypeCreate EventType = "create"
	EventTypeUpdate EventType = "update"
	EventTypeDelete EventType = "delete"
)

type BaseMessage struct {
	ProjectID string
	EventType EventType
}

func NewBaseMessage(eventType EventType, projectID string) BaseMessage {
	return BaseMessage{ProjectID: projectID, EventType: eventType}
}
