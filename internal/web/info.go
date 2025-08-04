package web

type ServerInfo struct {
	Version         string
	CommitHash      string
	BuildDate       string
	GoVersion       string
	SqliteVersion   string
	OS              string
	Arch            string
	AvailableUpdate *AvailableUpdate
}

type AvailableUpdate struct {
	IsAvailableManual bool   `json:"is_available_manual"`
	TagName           string `json:"tag_name"`
	URL               string `json:"url"`
	Description       string `json:"description,omitempty"`
}
