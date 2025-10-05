package store

import (
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/store/repos"
	"github.com/sxwebdev/tokenmanager"
)

type Store struct {
	*repos.Repos

	tokenRepo tokenmanager.ITokenStore

	sqlite *sql.DB
}

func New(sqlite *sql.DB) (*Store, error) {
	return &Store{
		Repos:     repos.New(sqlite),
		tokenRepo: tokenmanager.NewMemoryTokenStore(),
		sqlite:    sqlite,
	}, nil
}

func (s *Store) SQLite() *sql.DB { return s.sqlite }

func (s *Store) TokenRepo() tokenmanager.ITokenStore { return s.tokenRepo }
