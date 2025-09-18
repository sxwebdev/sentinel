package sqlite

import "context"

func (s *SQLite) Name() string { return "sqlite" }

func (s *SQLite) Start(_ context.Context) error { return nil }

func (s *SQLite) Stop(_ context.Context) error {
	return s.DB.Close()
}

func (s *SQLite) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

// Enabled returns true if the database is enabled
func (s *SQLite) Enabled() bool {
	return true
}
