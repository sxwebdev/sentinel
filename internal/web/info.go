package web

type ServerInfo struct {
	Version         string
	CommitHash      string
	BuildDate       string
	GoVersion       string
	OS              string
	Arch            string
	AvailableUpdate *AvailableUpdate
}

type AvailableUpdate struct {
	TagName     string `json:"tag_name"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}
