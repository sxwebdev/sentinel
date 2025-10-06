package store

import (
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/store/repos"
	"github.com/sxwebdev/tokenmanager"
)

type Store struct {
	*repos.Repos

	cache tokenmanager.ITokenStore

	sqlite *sql.DB
}

func New(sqlite *sql.DB, kvStore tokenmanager.ITokenStore) (*Store, error) {
	return &Store{
		Repos:  repos.New(sqlite),
		cache:  kvStore,
		sqlite: sqlite,
	}, nil
}

func (s *Store) SQLite() *sql.DB { return s.sqlite }

func (s *Store) Cache() tokenmanager.ITokenStore { return s.cache }
