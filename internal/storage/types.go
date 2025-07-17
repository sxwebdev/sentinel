package storage

import "time"

type CreateUpdateServiceRequest struct {
	Name      string              `json:"name" yaml:"name"`
	Protocol  ServiceProtocolType `json:"protocol" yaml:"protocol"`
	Interval  time.Duration       `json:"interval" yaml:"interval" swaggertype:"primitive,integer"`
	Timeout   time.Duration       `json:"timeout" yaml:"timeout" swaggertype:"primitive,integer"`
	Retries   int                 `json:"retries" yaml:"retries"`
	Tags      []string            `json:"tags" yaml:"tags"`
	Config    map[string]any      `json:"config" yaml:"config"`
	IsEnabled bool                `json:"is_enabled" yaml:"is_enabled"`
}
