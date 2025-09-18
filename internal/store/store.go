package store

import (
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/store/repos"
)

type Store struct {
	*repos.Repos

	sqlite *sql.DB
}

func New(sqlite *sql.DB) (*Store, error) {
	return &Store{
		Repos:  repos.New(sqlite),
		sqlite: sqlite,
	}, nil
}

func (s *Store) SQLite() *sql.DB { return s.sqlite }
