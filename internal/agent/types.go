package agent

type subscribeServiceType int

const (
	subscribeServiceTypeUpsert subscribeServiceType = iota + 1
	subscribeServiceTypeDelete
)
