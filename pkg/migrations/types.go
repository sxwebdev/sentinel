package migrations

type logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

type migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}
